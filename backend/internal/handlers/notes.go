package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/edda/backend/internal/middleware/auth"
)

type NoteResponse struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type CreateNoteRequest struct {
	HostID  *string `json:"host_id,omitempty"`
	PortID  *string `json:"port_id,omitempty"`
	URLID   *string `json:"url_id,omitempty"`
	Content string  `json:"content"`
}

func (h *Handlers) CreateNote(w http.ResponseWriter, r *http.Request) {
	userIDStr := auth.GetUserID(r)
	if userIDStr == "" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, _ := uuid.Parse(userIDStr)

	var req CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	var hostID, portID, urlID *uuid.UUID
	if req.HostID != nil && strings.TrimSpace(*req.HostID) != "" {
		if id, err := uuid.Parse(*req.HostID); err == nil {
			hostID = &id
		}
	}
	if req.PortID != nil && strings.TrimSpace(*req.PortID) != "" {
		if id, err := uuid.Parse(*req.PortID); err == nil {
			portID = &id
		}
	}
	if req.URLID != nil && strings.TrimSpace(*req.URLID) != "" {
		if id, err := uuid.Parse(*req.URLID); err == nil {
			urlID = &id
		}
	}
	n := 0
	if hostID != nil {
		n++
	}
	if portID != nil {
		n++
	}
	if urlID != nil {
		n++
	}
	if n != 1 {
		writeError(w, http.StatusBadRequest, "exactly one of host_id, port_id, or url_id is required")
		return
	}

	note, err := h.db.CreateNote(hostID, portID, urlID, content, &userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create note")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(NoteResponse{
		ID:        note.ID.String(),
		Content:   note.Content,
		CreatedAt: note.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *Handlers) ListNotesByHost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	hostID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid host ID")
		return
	}
	notes, err := h.db.ListNotes(&hostID, nil, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list notes")
		return
	}
	responses := make([]NoteResponse, len(notes))
	for i, n := range notes {
		responses[i] = NoteResponse{ID: n.ID.String(), Content: n.Content, CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func (h *Handlers) ListNotesByPort(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	portID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid port ID")
		return
	}
	notes, err := h.db.ListNotes(nil, &portID, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list notes")
		return
	}
	responses := make([]NoteResponse, len(notes))
	for i, n := range notes {
		responses[i] = NoteResponse{ID: n.ID.String(), Content: n.Content, CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func (h *Handlers) ListNotesByURL(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	urlID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid URL ID")
		return
	}
	notes, err := h.db.ListNotes(nil, nil, &urlID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list notes")
		return
	}
	responses := make([]NoteResponse, len(notes))
	for i, n := range notes {
		responses[i] = NoteResponse{ID: n.ID.String(), Content: n.Content, CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}
