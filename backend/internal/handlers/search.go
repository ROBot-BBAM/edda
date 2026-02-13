package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

const searchLimit = 10

type SearchResultHost struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type SearchResultURL struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type SearchResultFinding struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Severity string `json:"severity"`
}

type SearchResponse struct {
	Hosts    []SearchResultHost    `json:"hosts"`
	URLs     []SearchResultURL    `json:"urls"`
	Findings []SearchResultFinding `json:"findings"`
}

func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) < 2 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SearchResponse{Hosts: nil, URLs: nil, Findings: nil})
		return
	}

	limit := searchLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 20 {
			limit = n
		}
	}

	resp := SearchResponse{}

	hosts, err := h.db.ListHosts(q, nil)
	if err == nil && len(hosts) > 0 {
		n := limit
		if len(hosts) < n {
			n = len(hosts)
		}
		resp.Hosts = make([]SearchResultHost, n)
		for i := 0; i < n; i++ {
			label := hosts[i].IPAddress
			if hosts[i].Hostname.Valid && hosts[i].Hostname.String != "" {
				label = hosts[i].Hostname.String + " (" + hosts[i].IPAddress + ")"
			}
			resp.Hosts[i] = SearchResultHost{ID: hosts[i].ID.String(), Label: label}
		}
	}

	urls, err := h.db.ListURLs(q, nil, nil)
	if err == nil && len(urls) > 0 {
		n := limit
		if len(urls) < n {
			n = len(urls)
		}
		resp.URLs = make([]SearchResultURL, n)
		for i := 0; i < n; i++ {
			label := urls[i].Path
			if len(label) > 80 {
				label = label[:77] + "..."
			}
			resp.URLs[i] = SearchResultURL{ID: urls[i].ID.String(), Label: label}
		}
	}

	findings, err := h.db.SearchFindings(q, limit)
	if err == nil && len(findings) > 0 {
		resp.Findings = make([]SearchResultFinding, len(findings))
		for i, f := range findings {
			resp.Findings[i] = SearchResultFinding{ID: f.ID.String(), Title: f.Title, Severity: f.Severity}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
