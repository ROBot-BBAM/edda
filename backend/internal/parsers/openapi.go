package parsers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"log"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/edda/backend/internal/database"
)

// OpenAPI 3: openapi, servers[] { url }, paths { path -> { get, post, ... } }
// Swagger 2: swagger, host, basePath, schemes[], paths
var openAPIOperations = []string{"get", "post", "put", "delete", "patch", "head", "options", "trace"}

func ParseOpenAPI(filePath string, scanFileID uuid.UUID, db *database.DB) error {
	log.Printf("Parsing OpenAPI/Swagger spec: %s", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var spec map[string]interface{}
	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return fmt.Errorf("failed to decode YAML: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &spec); err != nil {
			return fmt.Errorf("failed to decode JSON: %w", err)
		}
	}

	bases, err := extractBaseURLs(spec)
	if err != nil {
		return err
	}
	if len(bases) == 0 {
		bases = []string{"/"}
	}

	paths, _ := spec["paths"].(map[string]interface{})
	if paths == nil {
		log.Printf("OpenAPI: no paths found")
		return nil
	}

	count := 0
	for pathTemplate, pathItem := range paths {
		pathObj, _ := pathItem.(map[string]interface{})
		if pathObj == nil {
			continue
		}
		for _, method := range openAPIOperations {
			if _, hasOp := pathObj[method]; !hasOp {
				continue
			}
			path := normalizePath(pathTemplate)
			for _, base := range bases {
				fullURL := base + path
				if err := upsertOpenAPIURL(db, fullURL, path, strings.ToUpper(method), scanFileID); err != nil {
					log.Printf("OpenAPI: failed to upsert %s %s: %v", method, fullURL, err)
					continue
				}
				count++
			}
		}
	}

	log.Printf("Successfully parsed %d endpoints from OpenAPI/Swagger spec", count)
	return nil
}

func extractBaseURLs(spec map[string]interface{}) ([]string, error) {
	// OpenAPI 3.x: servers[].url
	if servers, ok := spec["servers"].([]interface{}); ok && len(servers) > 0 {
		var bases []string
		for _, s := range servers {
			sv, _ := s.(map[string]interface{})
			if sv == nil {
				continue
			}
			u, _ := sv["url"].(string)
			if u == "" {
				continue
			}
			u = strings.TrimSuffix(u, "/")
			bases = append(bases, u)
		}
		if len(bases) > 0 {
			return bases, nil
		}
	}

	// Swagger 2: schemes, host, basePath
	schemes, _ := spec["schemes"].([]interface{})
	host, _ := spec["host"].(string)
	basePath, _ := spec["basePath"].(string)
	if host == "" {
		return []string{"/"}, nil
	}
	if len(schemes) == 0 {
		schemes = []interface{}{"https"}
	}
	basePath = strings.TrimSuffix(basePath, "/")
	var bases []string
	for _, s := range schemes {
		scheme, _ := s.(string)
		if scheme == "" {
			continue
		}
		bases = append(bases, scheme+"://"+host+basePath)
	}
	if len(bases) == 0 {
		bases = []string{"https://" + host + basePath}
	}
	return bases, nil
}

func normalizePath(path string) string {
	if path == "" || path[0] != '/' {
		path = "/" + path
	}
	return path
}

func upsertOpenAPIURL(db *database.DB, fullURL, path, method string, scanFileID uuid.UUID) error {
	parsed, err := url.Parse(fullURL)
	if err != nil {
		return err
	}
	path = parsed.Path
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}
	hostname := parsed.Hostname()
	var hostID *uuid.UUID
	if hostname != "" {
		hostID = getOrCreateHostForURL(db, hostname, scanFileID)
	}
	_, err = db.UpsertURL(fullURL, path, method, nil, nil, nil, nil, hostID, nil, scanFileID)
	return err
}
