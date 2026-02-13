package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/edda/backend/internal/database"
)

type FindingResponse struct {
	ID          string  `json:"id"`
	HostID      *string `json:"host_id,omitempty"`
	HostDisplay string  `json:"host_display,omitempty"`
	PortID      *string `json:"port_id,omitempty"`
	PortDisplay string  `json:"port_display,omitempty"`
	URLID       *string `json:"url_id,omitempty"`
	URLDisplay  string  `json:"url_display,omitempty"`
	Title       string  `json:"title"`
	Severity    string  `json:"severity"`
	Description string  `json:"description,omitempty"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CreateFindingRequest struct {
	HostID      *string `json:"host_id,omitempty"`
	PortID      *string `json:"port_id,omitempty"`
	URLID       *string `json:"url_id,omitempty"`
	Title       string  `json:"title"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
}

func findingToResponse(f *database.Finding, hostDisplay, portDisplay, urlDisplay string) FindingResponse {
	resp := FindingResponse{
		ID:          f.ID.String(),
		Title:       f.Title,
		Severity:    f.Severity,
		Status:      f.Status,
		CreatedAt:   f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		HostDisplay: hostDisplay,
		PortDisplay: portDisplay,
		URLDisplay:  urlDisplay,
	}
	if f.Description.Valid {
		resp.Description = f.Description.String
	}
	if f.HostID.Valid {
		s := f.HostID.UUID.String()
		resp.HostID = &s
	}
	if f.PortID.Valid {
		s := f.PortID.UUID.String()
		resp.PortID = &s
	}
	if f.URLID.Valid {
		s := f.URLID.UUID.String()
		resp.URLID = &s
	}
	return resp
}

func (h *Handlers) ListFindings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var hostID *uuid.UUID
	if v := strings.TrimSpace(q.Get("host_id")); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			hostID = &id
		}
	}
	var portIDs []uuid.UUID
	for _, v := range q["port_id"] {
		if id, err := uuid.Parse(strings.TrimSpace(v)); err == nil {
			portIDs = append(portIDs, id)
		}
	}
	var urlID *uuid.UUID
	if v := strings.TrimSpace(q.Get("url_id")); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			urlID = &id
		}
	}
	severity := strings.TrimSpace(q.Get("severity"))
	status := strings.TrimSpace(q.Get("status"))

	findings, err := h.db.ListFindings(hostID, portIDs, urlID, severity, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list findings")
		return
	}

	// Enrich with display strings for host, port, URL
	responses := make([]FindingResponse, len(findings))
	for i, f := range findings {
		var hostDisp, portDisp, urlDisp string
		if f.HostID.Valid {
			if host, err := h.db.GetHostByID(f.HostID.UUID); err == nil && host != nil {
				if host.Hostname.Valid && host.Hostname.String != "" {
					hostDisp = host.Hostname.String
				} else {
					hostDisp = host.IPAddress
				}
			}
		}
		if f.PortID.Valid {
			if port, err := h.db.GetPortByID(f.PortID.UUID); err == nil && port != nil {
				portDisp = port.Protocol + "/" + strconv.Itoa(port.Port)
			}
		}
		if f.URLID.Valid {
			if u, err := h.db.GetURLByID(f.URLID.UUID); err == nil && u != nil {
				urlDisp = u.Path
				if len(urlDisp) > 60 {
					urlDisp = urlDisp[:57] + "..."
				}
			}
		}
		responses[i] = findingToResponse(f, hostDisp, portDisp, urlDisp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

type FindingsSummaryResponse struct {
	BySeverity map[string]int `json:"by_severity"`
	OpenCount  int             `json:"open_count"`
}

func (h *Handlers) GetFindingsSummary(w http.ResponseWriter, r *http.Request) {
	bySeverity, openCount, err := h.db.GetFindingsSummary()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get findings summary")
		return
	}
	if bySeverity == nil {
		bySeverity = make(map[string]int)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(FindingsSummaryResponse{BySeverity: bySeverity, OpenCount: openCount})
}

func (h *Handlers) CreateFinding(w http.ResponseWriter, r *http.Request) {
	var req CreateFindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
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
	if hostID == nil && portID == nil && urlID == nil {
		writeError(w, http.StatusBadRequest, "at least one of host_id, port_id, or url_id is required")
		return
	}

	severity := strings.TrimSpace(req.Severity)
	if severity == "" {
		severity = "medium"
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "open"
	}

	finding, err := h.db.CreateFinding(hostID, portID, urlID, req.Title, severity, strings.TrimSpace(req.Description), status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create finding")
		return
	}

	var hostDisp, portDisp, urlDisp string
	if finding.HostID.Valid {
		if host, err := h.db.GetHostByID(finding.HostID.UUID); err == nil && host != nil {
			if host.Hostname.Valid && host.Hostname.String != "" {
				hostDisp = host.Hostname.String
			} else {
				hostDisp = host.IPAddress
			}
		}
	}
	if finding.PortID.Valid {
		if port, err := h.db.GetPortByID(finding.PortID.UUID); err == nil && port != nil {
			portDisp = port.Protocol + "/" + strconv.Itoa(port.Port)
		}
	}
	if finding.URLID.Valid {
		if u, err := h.db.GetURLByID(finding.URLID.UUID); err == nil && u != nil {
			urlDisp = u.Path
			if len(urlDisp) > 60 {
				urlDisp = urlDisp[:57] + "..."
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(findingToResponse(finding, hostDisp, portDisp, urlDisp))
}

func (h *Handlers) GetFinding(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid finding ID")
		return
	}
	finding, err := h.db.GetFindingByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get finding")
		return
	}
	if finding == nil {
		writeError(w, http.StatusNotFound, "Finding not found")
		return
	}
	var hostDisp, portDisp, urlDisp string
	if finding.HostID.Valid {
		if host, err := h.db.GetHostByID(finding.HostID.UUID); err == nil && host != nil {
			if host.Hostname.Valid && host.Hostname.String != "" {
				hostDisp = host.Hostname.String
			} else {
				hostDisp = host.IPAddress
			}
		}
	}
	if finding.PortID.Valid {
		if port, err := h.db.GetPortByID(finding.PortID.UUID); err == nil && port != nil {
			portDisp = port.Protocol + "/" + strconv.Itoa(port.Port)
		}
	}
	if finding.URLID.Valid {
		if u, err := h.db.GetURLByID(finding.URLID.UUID); err == nil && u != nil {
			urlDisp = u.Path
			if len(urlDisp) > 60 {
				urlDisp = urlDisp[:57] + "..."
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(findingToResponse(finding, hostDisp, portDisp, urlDisp))
}

func (h *Handlers) UpdateFinding(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid finding ID")
		return
	}
	existing, err := h.db.GetFindingByID(id)
	if err != nil || existing == nil {
		writeError(w, http.StatusNotFound, "Finding not found")
		return
	}
	var req CreateFindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	// Merge: use request values when provided, else keep existing
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = existing.Title
	}
	severity := strings.TrimSpace(req.Severity)
	if severity == "" {
		severity = existing.Severity
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = existing.Status
	}
	description := strings.TrimSpace(req.Description)
	if description == "" && existing.Description.Valid {
		description = existing.Description.String
	}
	hostID, portID, urlID := existing.HostID, existing.PortID, existing.URLID
	if req.HostID != nil {
		if v := strings.TrimSpace(*req.HostID); v != "" {
			if parsed, err := uuid.Parse(v); err == nil {
				hostID = database.NullUUID{UUID: parsed, Valid: true}
			}
		} else {
			hostID = database.NullUUID{}
		}
	}
	if req.PortID != nil {
		if v := strings.TrimSpace(*req.PortID); v != "" {
			if parsed, err := uuid.Parse(v); err == nil {
				portID = database.NullUUID{UUID: parsed, Valid: true}
			}
		} else {
			portID = database.NullUUID{}
		}
	}
	if req.URLID != nil {
		if v := strings.TrimSpace(*req.URLID); v != "" {
			if parsed, err := uuid.Parse(v); err == nil {
				urlID = database.NullUUID{UUID: parsed, Valid: true}
			}
		} else {
			urlID = database.NullUUID{}
		}
	}
	if !hostID.Valid && !portID.Valid && !urlID.Valid {
		writeError(w, http.StatusBadRequest, "at least one of host_id, port_id, or url_id is required")
		return
	}
	finding, err := h.db.UpdateFinding(id, hostID, portID, urlID, title, severity, description, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update finding")
		return
	}
	var hostDisp, portDisp, urlDisp string
	if finding.HostID.Valid {
		if host, err := h.db.GetHostByID(finding.HostID.UUID); err == nil && host != nil {
			if host.Hostname.Valid && host.Hostname.String != "" {
				hostDisp = host.Hostname.String
			} else {
				hostDisp = host.IPAddress
			}
		}
	}
	if finding.PortID.Valid {
		if port, err := h.db.GetPortByID(finding.PortID.UUID); err == nil && port != nil {
			portDisp = port.Protocol + "/" + strconv.Itoa(port.Port)
		}
	}
	if finding.URLID.Valid {
		if u, err := h.db.GetURLByID(finding.URLID.UUID); err == nil && u != nil {
			urlDisp = u.Path
			if len(urlDisp) > 60 {
				urlDisp = urlDisp[:57] + "..."
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(findingToResponse(finding, hostDisp, portDisp, urlDisp))
}

func (h *Handlers) DeleteFinding(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid finding ID")
		return
	}
	err = h.db.DeleteFinding(id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "Finding not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to delete finding")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
