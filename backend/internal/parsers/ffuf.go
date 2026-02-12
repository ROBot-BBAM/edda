package parsers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/edda/backend/internal/database"
)

// FFuf JSON format - single result. Supports both wrapper "results" array and NDJSON.
// Field tags match ffuf's actual JSON: "status", "ContentLength", "ContentWords", "ContentLines", "url".
type FFufResult struct {
	Input    map[string]string `json:"input"` // FUZZ etc.
	Position int               `json:"position"`
	Status   int               `json:"status"`
	Length   int               `json:"ContentLength"` // ffuf uses PascalCase for these
	Words    int               `json:"ContentWords"`
	Lines    int               `json:"ContentLines"`
	URL      string            `json:"url"`
	Redirect string            `json:"redirectlocation"`
}

// FFuf JSON wrapper - standard ffuf -of json output (single object with "results" array)
type FFufJSONOutput struct {
	Results []FFufResult `json:"results"`
}

func ParseFFufJSON(filePath string, scanFileID uuid.UUID, db *database.DB) error {
	log.Printf("Parsing ffuf JSON file: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	count := 0

	// Try standard ffuf format first: single object with "results" array
	var wrapper FFufJSONOutput
	if err := decoder.Decode(&wrapper); err == nil {
		for _, result := range wrapper.Results {
			if err := processFFufResult(result, scanFileID, db); err != nil {
				log.Printf("Failed to process ffuf result: %v", err)
				continue
			}
			count++
		}
		log.Printf("Successfully parsed %d URLs from ffuf JSON file (wrapper format)", count)
		return nil
	}

	// Decode failed - re-open and try NDJSON (one JSON object per line)
	file, err = os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to re-open file: %w", err)
	}
	defer file.Close()

	decoder = json.NewDecoder(file)
	for {
		var result FFufResult
		if err := decoder.Decode(&result); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("failed to decode JSON: %w", err)
		}

		if err := processFFufResult(result, scanFileID, db); err != nil {
			log.Printf("Failed to process ffuf result: %v", err)
			continue
		}
		count++
	}

	log.Printf("Successfully parsed %d URLs from ffuf JSON file", count)
	return nil
}

func ParseFFufCSV(filePath string, scanFileID uuid.UUID, db *database.DB) error {
	log.Printf("Parsing ffuf CSV file: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header (trim BOM from first column if present, e.g. Excel export)
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}
	if len(headers) > 0 {
		headers[0] = strings.TrimPrefix(headers[0], "\ufeff")
	}

	// Find column indices. FFuf CSV uses: url, status_code, content_length, content_words, content_lines
	urlIdx := -1
	statusIdx := -1
	lengthIdx := -1
	wordsIdx := -1
	linesIdx := -1

	for i, header := range headers {
		header = strings.ToLower(strings.TrimSpace(header))
		switch header {
		case "url":
			urlIdx = i
		case "status", "status_code":
			statusIdx = i
		case "length", "content_length":
			lengthIdx = i
		case "words", "content_words":
			wordsIdx = i
		case "lines", "content_lines":
			linesIdx = i
		}
	}

	if urlIdx == -1 {
		return fmt.Errorf("URL column not found in CSV (headers: %q)", headers)
	}

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV record: %w", err)
		}

		if len(record) <= urlIdx {
			continue
		}

		urlStr := strings.TrimSpace(record[urlIdx])
		if urlStr == "" {
			continue
		}
		var status, length, words, lines *int

		if statusIdx >= 0 && statusIdx < len(record) && record[statusIdx] != "" {
			if val, err := strconv.Atoi(record[statusIdx]); err == nil {
				status = &val
			}
		}
		if lengthIdx >= 0 && lengthIdx < len(record) && record[lengthIdx] != "" {
			if val, err := strconv.Atoi(record[lengthIdx]); err == nil {
				length = &val
			}
		}
		if wordsIdx >= 0 && wordsIdx < len(record) && record[wordsIdx] != "" {
			if val, err := strconv.Atoi(record[wordsIdx]); err == nil {
				words = &val
			}
		}
		if linesIdx >= 0 && linesIdx < len(record) && record[linesIdx] != "" {
			if val, err := strconv.Atoi(record[linesIdx]); err == nil {
				lines = &val
			}
		}

		// Parse URL to extract path
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			log.Printf("Failed to parse URL %s: %v", urlStr, err)
			continue
		}

		path := parsedURL.Path
		if parsedURL.RawQuery != "" {
			path += "?" + parsedURL.RawQuery
		}

		// Get or create host so host count reflects hosts discovered via ffuf
		var hostID *uuid.UUID
		hostname := parsedURL.Hostname()
		if hostname != "" {
			hostID = getOrCreateHostForURL(db, hostname, scanFileID)
		}

		_, err = db.UpsertURL(urlStr, path, "GET", status, length, words, lines, hostID, nil, scanFileID)
		if err != nil {
			log.Printf("Failed to upsert URL %s: %v", urlStr, err)
			continue
		}
		count++
	}

	log.Printf("Successfully parsed %d URLs from ffuf CSV file", count)
	return nil
}

// getOrCreateHostForURL returns the host ID for the given hostname (from a URL).
// Tries ip_address first (ffuf-created or nmap by IP), then hostname (nmap host with that hostname),
// then creates a new host so the host count updates and URLs link to it.
func getOrCreateHostForURL(db *database.DB, hostname string, scanFileID uuid.UUID) *uuid.UUID {
	// Match host by ip_address (covers: nmap host by IP, or ffuf-created host keyed by hostname)
	host, err := db.GetHostByIP(hostname)
	if err == nil && host != nil {
		return &host.ID
	}
	// Match by hostname so URLs land on existing nmap host (e.g. ip_address=192.168.1.1, hostname=target.local)
	host, err = db.GetHostByHostname(hostname)
	if err == nil && host != nil {
		return &host.ID
	}
	// No existing host; create one so host count updates and URLs have a host to link to.
	host, err = db.UpsertHost(hostname, hostname, "", scanFileID)
	if err != nil {
		log.Printf("Failed to create host for %s: %v", hostname, err)
		return nil
	}
	return &host.ID
}

func processFFufResult(result FFufResult, scanFileID uuid.UUID, db *database.DB) error {
	if result.URL == "" {
		return nil
	}

	// Parse URL to extract path
	parsedURL, err := url.Parse(result.URL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	path := parsedURL.Path
	if parsedURL.RawQuery != "" {
		path += "?" + parsedURL.RawQuery
	}

	// Get or create host so host count reflects hosts discovered via ffuf
	var hostID *uuid.UUID
	hostname := parsedURL.Hostname()
	if hostname != "" {
		hostID = getOrCreateHostForURL(db, hostname, scanFileID)
	}

	status := result.Status
	length := result.Length
	words := result.Words
	lines := result.Lines

	_, err = db.UpsertURL(result.URL, path, "GET", &status, &length, &words, &lines, hostID, nil, scanFileID)
	return err
}
