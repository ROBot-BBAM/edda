package parsers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"log"

	"github.com/google/uuid"

	"github.com/edda/backend/internal/database"
)

// Postman collection v2.1: top-level "item" (array of requests or folders), "variable" (optional).
// Each item can have "request" (string URL or object with method, url) or nested "item".
type postmanCollection struct {
	Item     []postmanItem     `json:"item"`
	Variable []postmanVariable `json:"variable"`
}

type postmanVariable struct {
	Key   string `json:"key"`
	ID    string `json:"id"`
	Value string `json:"value"`
}

type postmanItem struct {
	Name    string          `json:"name"`
	Request json.RawMessage `json:"request"`
	Item    []postmanItem   `json:"item"`
}

// request can be a string (URL) or object with method and url
type postmanReq struct {
	Method string          `json:"method"`
	URL    json.RawMessage `json:"url"`
}

// postmanURLObject is when request.url is an object
type postmanURLObject struct {
	Raw      string   `json:"raw"`
	Protocol string   `json:"protocol"`
	Host     []string `json:"host"`
	Port     string   `json:"port"`
	Path     []string `json:"path"`
	Query    []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"query"`
}

func ParsePostman(filePath string, scanFileID uuid.UUID, db *database.DB) error {
	log.Printf("Parsing Postman collection: %s", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var col postmanCollection
	if err := json.Unmarshal(data, &col); err != nil {
		return fmt.Errorf("failed to decode Postman JSON: %w", err)
	}

	vars := make(map[string]string)
	for _, v := range col.Variable {
		k := v.Key
		if k == "" {
			k = v.ID
		}
		if k != "" {
			vars[k] = v.Value
		}
	}

	count := 0
	for i := range col.Item {
		n := extractPostmanRequests(&col.Item[i], vars, scanFileID, db)
		count += n
	}

	log.Printf("Successfully parsed %d URLs from Postman collection", count)
	return nil
}

func extractPostmanRequests(item *postmanItem, vars map[string]string, scanFileID uuid.UUID, db *database.DB) int {
	count := 0
	if len(item.Item) > 0 {
		for i := range item.Item {
			count += extractPostmanRequests(&item.Item[i], vars, scanFileID, db)
		}
		return count
	}
	if len(item.Request) == 0 {
		return 0
	}
	// This item is a request: "request" can be a string (URL) or object with method + url
	urlStr, method := resolvePostmanRequest(item.Request, vars)
	if urlStr == "" {
		return 0
	}
	if method == "" {
		method = "GET"
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("Postman: invalid URL %q: %v", urlStr, err)
		return 0
	}
	path := parsedURL.Path
	if parsedURL.RawQuery != "" {
		path += "?" + parsedURL.RawQuery
	}
	hostname := parsedURL.Hostname()
	var hostID *uuid.UUID
	if hostname != "" {
		hostID = getOrCreateHostForURL(db, hostname, scanFileID)
	}
	_, err = db.UpsertURL(urlStr, path, strings.ToUpper(method), nil, nil, nil, nil, hostID, nil, scanFileID)
	if err != nil {
		log.Printf("Postman: failed to upsert URL %s: %v", urlStr, err)
		return 0
	}
	return 1
}

func resolvePostmanRequest(requestRaw json.RawMessage, vars map[string]string) (urlStr, method string) {
	// request can be a string (URL only) or object
	var s string
	if err := json.Unmarshal(requestRaw, &s); err == nil {
		return substituteVars(s, vars), "GET"
	}
	var req postmanReq
	if err := json.Unmarshal(requestRaw, &req); err != nil {
		return "", "GET"
	}
	method = strings.TrimSpace(req.Method)
	if method == "" {
		method = "GET"
	}
	if req.URL == nil {
		return "", method
	}
	// Try URL as string first
	if err := json.Unmarshal(req.URL, &s); err == nil {
		return substituteVars(s, vars), method
	}
	// URL object form
	var u postmanURLObject
	if err := json.Unmarshal(req.URL, &u); err != nil {
		return "", method
	}
	if u.Raw != "" {
		return substituteVars(u.Raw, vars), method
	}
	// Build from protocol, host, port, path, query
	protocol := u.Protocol
	if protocol == "" {
		protocol = "https"
	}
	host := strings.Join(u.Host, ".")
	if host == "" {
		return "", method
	}
	path := "/" + strings.TrimLeft(strings.Join(u.Path, "/"), "/")
	if u.Port != "" {
		if port, err := strconv.Atoi(u.Port); err == nil && (port != 443 && port != 80) {
			host = host + ":" + u.Port
		}
	}
	var q string
	for _, p := range u.Query {
		if q != "" {
			q += "&"
		}
		q += url.QueryEscape(p.Key) + "=" + url.QueryEscape(p.Value)
	}
	if q != "" {
		path += "?" + q
	}
	urlStr = protocol + "://" + host + path
	return substituteVars(urlStr, vars), method
}

func substituteVars(s string, vars map[string]string) string {
	for k, v := range vars {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
		s = strings.ReplaceAll(s, "{{"+strings.ToLower(k)+"}}", v)
	}
	return s
}
