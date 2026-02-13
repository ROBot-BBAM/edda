package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/edda/backend/internal/database"
	"github.com/edda/backend/internal/middleware/auth"
	"github.com/edda/backend/internal/parsers"
)

type ScanFileResponse struct {
	ID           string  `json:"id"`
	Type         string  `json:"type"`
	Filename     string  `json:"filename"`
	UploadedBy   string  `json:"uploaded_by"`
	UploadedAt   string  `json:"uploaded_at"`
	Status       string  `json:"status"`
	ErrorMessage *string `json:"error_message,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

const (
	maxUploadSize = 100 * 1024 * 1024 // 100MB
	uploadDir     = "/tmp/edda-uploads"
)

func init() {
	// Create upload directory if it doesn't exist
	os.MkdirAll(uploadDir, 0755)
}

func (h *Handlers) UploadScanFile(w http.ResponseWriter, r *http.Request) {
	log.Printf("UploadScanFile called: Method=%s, Content-Type=%s", r.Method, r.Header.Get("Content-Type"))
	
	userIDStr := auth.GetUserID(r)
	if userIDStr == "" {
		log.Printf("UploadScanFile: Unauthorized - no user ID")
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("UploadScanFile: Invalid user ID: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	log.Printf("UploadScanFile: User ID=%s, parsing multipart form", userIDStr)

	// Parse multipart form (limit to 100MB)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		log.Printf("UploadScanFile: Failed to parse multipart form: %v", err)
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse form: %v", err))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("UploadScanFile: Failed to get file from form: %v", err)
		writeError(w, http.StatusBadRequest, fmt.Sprintf("No file provided: %v", err))
		return
	}
	defer file.Close()
	
	log.Printf("UploadScanFile: File received - filename=%s, size=%d", header.Filename, header.Size)

	// Determine file type from extension (and for JSON, peek content to distinguish Postman/OpenAPI/FFuf)
	filename := header.Filename
	ext := strings.ToLower(filepath.Ext(filename))
	var fileType string
	var peek []byte
	const jsonPeekSize = 8192

	switch ext {
	case ".xml":
		fileType = "nmap_xml"
	case ".json":
		peekBuf := make([]byte, jsonPeekSize)
		n, _ := file.Read(peekBuf)
		peek = peekBuf[:n]
		fileType = detectJSONFileType(peek)
	case ".csv":
		fileType = "ffuf_csv"
	case ".yaml", ".yml":
		fileType = "openapi_yaml"
	default:
		log.Printf("UploadScanFile: Unsupported file type: %s", ext)
		writeError(w, http.StatusBadRequest, "Unsupported file type. Supported: .xml (nmap), .json (ffuf/postman/openapi), .csv (ffuf), .yaml/.yml (openapi)")
		return
	}

	// Create unique filename
	fileID := uuid.New()
	storageFilename := fmt.Sprintf("%s_%s", fileID.String(), filename)
	storagePath := filepath.Join(uploadDir, storageFilename)

	// Save file to disk
	log.Printf("UploadScanFile: Saving file to %s", storagePath)
	dst, err := os.Create(storagePath)
	if err != nil {
		log.Printf("UploadScanFile: Failed to create file: %v", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create file: %v", err))
		return
	}
	defer dst.Close()

	if len(peek) > 0 {
		if _, err := dst.Write(peek); err != nil {
			log.Printf("UploadScanFile: Failed to write file: %v", err)
			os.Remove(storagePath)
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save file: %v", err))
			return
		}
	}
	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("UploadScanFile: Failed to copy file: %v", err)
		os.Remove(storagePath) // Clean up on error
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save file: %v", err))
		return
	}
	log.Printf("UploadScanFile: File saved successfully")

	// Create database record
	log.Printf("UploadScanFile: Creating database record for file %s", filename)
	scanFile, err := h.db.CreateScanFile(filename, fileType, storagePath, userID)
	if err != nil {
		log.Printf("UploadScanFile: Failed to create database record: %v", err)
		os.Remove(storagePath) // Clean up on error
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create scan file record: %v", err))
		return
	}

	log.Printf("UploadScanFile: Successfully uploaded file %s (ID: %s)", filename, scanFile.ID.String())

	// Start parsing in background
	go func() {
		if err := h.parseScanFile(scanFile); err != nil {
			log.Printf("Background parsing failed for file %s: %v", scanFile.Filename, err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(scanFileToResponse(scanFile))
}

func (h *Handlers) ListScanFiles(w http.ResponseWriter, r *http.Request) {
	scanFiles, err := h.db.ListScanFiles()
	if err != nil {
		http.Error(w, "Failed to list scan files", http.StatusInternalServerError)
		return
	}

	responses := make([]ScanFileResponse, len(scanFiles))
	for i, sf := range scanFiles {
		responses[i] = *scanFileToResponse(sf)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// detectJSONFileType peeks at the first bytes of a JSON file to distinguish
// Postman collection, OpenAPI/Swagger spec, or FFuf results.
func detectJSONFileType(peek []byte) string {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(peek, &raw); err != nil {
		return "ffuf_json"
	}
	if _, ok := raw["openapi"]; ok {
		return "openapi_json"
	}
	if _, ok := raw["swagger"]; ok {
		return "openapi_json"
	}
	if _, hasInfo := raw["info"]; hasInfo {
		if _, hasItem := raw["item"]; hasItem {
			return "postman_json"
		}
	}
	return "ffuf_json"
}

func scanFileToResponse(sf *database.ScanFile) *ScanFileResponse {
	resp := &ScanFileResponse{
		ID:         sf.ID.String(),
		Type:       sf.Type,
		Filename:   sf.Filename,
		UploadedBy: sf.UploadedBy.String(),
		UploadedAt: sf.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		Status:     sf.Status,
		CreatedAt:  sf.UploadedAt.Format("2006-01-02T15:04:05Z07:00"), // Use uploaded_at as created_at
	}

	if sf.ErrorMessage.Valid {
		resp.ErrorMessage = &sf.ErrorMessage.String
	}

	return resp
}

func (h *Handlers) parseScanFile(scanFile *database.ScanFile) error {
	return parsers.ParseScanFile(scanFile, h.db)
}
