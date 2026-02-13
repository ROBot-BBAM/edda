package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/edda/backend/internal/database"
	"github.com/edda/backend/internal/middleware/auth"
)

type HostResponse struct {
	ID           string  `json:"id"`
	IPAddress    string  `json:"ip_address"`
	Hostname     *string `json:"hostname,omitempty"`
	OS           *string `json:"os,omitempty"`
	Reviewed     bool    `json:"reviewed"`
	FindingCount int     `json:"finding_count"`
	NoteCount    int     `json:"note_count"`
	NotePreview  string  `json:"note_preview,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

func (h *Handlers) ListHosts(w http.ResponseWriter, r *http.Request) {
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
	hosts, err := h.db.ListHosts(search, reviewed)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list hosts")
		return
	}

	hostIDs := make([]uuid.UUID, len(hosts))
	for i, host := range hosts {
		hostIDs[i] = host.ID
	}
	findingCounts, _ := h.db.CountFindingsByHostIDs(hostIDs)
	noteCounts, _ := h.db.CountNotesByHostIDs(hostIDs)

	responses := make([]HostResponse, len(hosts))
	for i, host := range hosts {
		resp := hostToResponse(host)
		resp.FindingCount = findingCounts[host.ID]
		resp.NoteCount = noteCounts[host.ID]
		if preview, ok := h.db.GetLatestNotePreview(&host.ID, nil, nil, 80); ok {
			resp.NotePreview = preview
		}
		responses[i] = resp
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

type HostDetailResponse struct {
	Host  HostResponse   `json:"host"`
	Ports []PortResponse  `json:"ports"`
	URLs  []URLResponse   `json:"urls"`
}

func (h *Handlers) GetHost(w http.ResponseWriter, r *http.Request) {
	hostIDStr := chi.URLParam(r, "id")
	hostID, err := uuid.Parse(hostIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid host ID")
		return
	}

	host, err := h.db.GetHostByID(hostID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get host")
		return
	}
	if host == nil {
		writeError(w, http.StatusNotFound, "Host not found")
		return
	}

	// Get ports for this host
	ports, err := h.db.GetPortsByHostID(hostID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get ports")
		return
	}

	// Get URLs for this host
	urls, err := h.db.GetURLsByHostID(hostID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get URLs")
		return
	}

	portIDs := make([]uuid.UUID, len(ports))
	for i, p := range ports {
		portIDs[i] = p.ID
	}
	urlIDs := make([]uuid.UUID, len(urls))
	for i, u := range urls {
		urlIDs[i] = u.ID
	}
	portFindingCounts, _ := h.db.CountFindingsByPortIDs(portIDs)
	urlFindingCounts, _ := h.db.CountFindingsByURLIDs(urlIDs)
	portNoteCounts, _ := h.db.CountNotesByPortIDs(portIDs)
	urlNoteCounts, _ := h.db.CountNotesByURLIDs(urlIDs)

	portResponses := make([]PortResponse, len(ports))
	for i, port := range ports {
		resp := portToResponse(port)
		resp.FindingCount = portFindingCounts[port.ID]
		resp.NoteCount = portNoteCounts[port.ID]
		if preview, ok := h.db.GetLatestNotePreview(nil, &port.ID, nil, 80); ok {
			resp.NotePreview = preview
		}
		portResponses[i] = resp
	}

	urlResponses := make([]URLResponse, len(urls))
	for i, urlObj := range urls {
		resp := urlToResponse(urlObj)
		resp.FindingCount = urlFindingCounts[urlObj.ID]
		resp.NoteCount = urlNoteCounts[urlObj.ID]
		if preview, ok := h.db.GetLatestNotePreview(nil, nil, &urlObj.ID, 80); ok {
			resp.NotePreview = preview
		}
		urlResponses[i] = resp
	}

	hostResp := hostToResponse(host)
	if noteCounts, err := h.db.CountNotesByHostIDs([]uuid.UUID{host.ID}); err == nil {
		hostResp.NoteCount = noteCounts[host.ID]
	}
	if preview, ok := h.db.GetLatestNotePreview(&host.ID, nil, nil, 80); ok {
		hostResp.NotePreview = preview
	}
	response := HostDetailResponse{
		Host:  hostResp,
		Ports: portResponses,
		URLs:  urlResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func hostToResponse(h *database.Host) HostResponse {
	resp := HostResponse{
		ID:           h.ID.String(),
		IPAddress:    h.IPAddress,
		Reviewed:     h.Reviewed,
		FindingCount: 0, // set by caller when building list
		NoteCount:    0,
		CreatedAt:    h.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if h.Hostname.Valid {
		resp.Hostname = &h.Hostname.String
	}
	if h.OS.Valid {
		resp.OS = &h.OS.String
	}

	return resp
}

func (h *Handlers) SetHostReviewed(w http.ResponseWriter, r *http.Request) {
	hostIDStr := chi.URLParam(r, "id")
	hostID, err := uuid.Parse(hostIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid host ID")
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

	if err := h.db.SetHostReviewed(hostID, req.Reviewed, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update host")
		return
	}

	host, err := h.db.GetHostByID(hostID)
	if err != nil || host == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hostToResponse(host))
}
