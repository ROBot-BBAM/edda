package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/edda/backend/internal/database"
	"github.com/edda/backend/internal/middleware/auth"
)

type URLResponse struct {
	ID            string  `json:"id"`
	URL           string  `json:"url"`
	Path          string  `json:"path"`
	Method        string  `json:"method"`
	StatusCode    *int    `json:"status_code,omitempty"`
	ContentLength *int    `json:"content_length,omitempty"`
	Words         *int    `json:"words,omitempty"`
	Lines         *int    `json:"lines,omitempty"`
	Reviewed      bool    `json:"reviewed"`
	FindingCount  int     `json:"finding_count"`
	NoteCount     int     `json:"note_count"`
	NotePreview   string  `json:"note_preview,omitempty"`
	CreatedAt     string  `json:"created_at"`
	HostID        *string `json:"host_id,omitempty"`
	Host          string  `json:"host"` // display: ip_address or hostname, empty if no host
}

func (h *Handlers) ListURLs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := strings.TrimSpace(q.Get("search"))
	var reviewed *bool
	if v := q.Get("reviewed"); v == "true" {
		t := true
		reviewed = &t
	} else if v == "false" {
		f := false
		reviewed = &f
	}
	var statusCode *int
	if v := q.Get("status"); v != "" {
		if code, err := strconv.Atoi(v); err == nil && code >= 0 && code < 1000 {
			statusCode = &code
		}
	}
	urls, err := h.db.ListURLs(search, reviewed, statusCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list URLs")
		return
	}

	urlIDs := make([]uuid.UUID, len(urls))
	for i, u := range urls {
		urlIDs[i] = u.ID
	}
	findingCounts, _ := h.db.CountFindingsByURLIDs(urlIDs)
	noteCounts, _ := h.db.CountNotesByURLIDs(urlIDs)

	// Build host display map (host_id -> display string) so we can show host on each URL
	hostDisplay := make(map[string]string)
	for _, u := range urls {
		if u.HostID.Valid {
			idStr := u.HostID.UUID.String()
			if _, ok := hostDisplay[idStr]; ok {
				continue
			}
			host, err := h.db.GetHostByID(u.HostID.UUID)
			if err == nil && host != nil {
				if host.Hostname.Valid && host.Hostname.String != "" {
					hostDisplay[idStr] = host.Hostname.String
				} else {
					hostDisplay[idStr] = host.IPAddress
				}
			}
		}
	}

	responses := make([]URLResponse, len(urls))
	for i, u := range urls {
		responses[i] = urlToResponse(u)
		responses[i].FindingCount = findingCounts[u.ID]
		responses[i].NoteCount = noteCounts[u.ID]
		if preview, ok := h.db.GetLatestNotePreview(nil, nil, &u.ID, 80); ok {
			responses[i].NotePreview = preview
		}
		if u.HostID.Valid {
			idStr := u.HostID.UUID.String()
			responses[i].HostID = &idStr
			responses[i].Host = hostDisplay[idStr]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func urlToResponse(u *database.URL) URLResponse {
	resp := URLResponse{
		ID:           u.ID.String(),
		URL:          u.URL,
		Path:         u.Path,
		Method:       u.Method,
		Reviewed:     u.Reviewed,
		FindingCount: 0, // set by caller when building list
		NoteCount:    0,
		CreatedAt:    u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if u.StatusCode.Valid {
		statusCode := int(u.StatusCode.Int64)
		resp.StatusCode = &statusCode
	}
	if u.ContentLength.Valid {
		contentLength := int(u.ContentLength.Int64)
		resp.ContentLength = &contentLength
	}
	if u.Words.Valid {
		words := int(u.Words.Int64)
		resp.Words = &words
	}
	if u.Lines.Valid {
		lines := int(u.Lines.Int64)
		resp.Lines = &lines
	}

	return resp
}

func (h *Handlers) SetURLReviewed(w http.ResponseWriter, r *http.Request) {
	urlIDStr := chi.URLParam(r, "id")
	urlID, err := uuid.Parse(urlIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid URL ID")
		return
	}

	userIDStr := auth.GetUserID(r)
	if userIDStr == "" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Invalid user")
		return
	}

	var req SetReviewedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.db.SetURLReviewed(urlID, req.Reviewed, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update URL")
		return
	}

	u, err := h.db.GetURLByID(urlID)
	if err != nil || u == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(urlToResponse(u))
}
