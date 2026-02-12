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

type PortResponse struct {
	ID             string  `json:"id"`
	HostID         string  `json:"host_id"`
	Port           int     `json:"port"`
	Protocol       string  `json:"protocol"`
	State          *string `json:"state,omitempty"`
	ServiceName    *string `json:"service_name,omitempty"`
	ServiceProduct *string `json:"service_product,omitempty"`
	ServiceVersion *string `json:"service_version,omitempty"`
	Reviewed       bool    `json:"reviewed"`
	CreatedAt      string  `json:"created_at"`
}

// PortAggregateResponse is one row per (port, protocol); reviewed is true only when all host-ports are reviewed.
type PortAggregateResponse struct {
	Port           int     `json:"port"`
	Protocol       string  `json:"protocol"`
	State          *string `json:"state,omitempty"`
	ServiceName    *string `json:"service_name,omitempty"`
	ServiceProduct *string `json:"service_product,omitempty"`
	ServiceVersion *string `json:"service_version,omitempty"`
	Reviewed       bool    `json:"reviewed"`
	HostCount      int     `json:"host_count"`
}

func portAggregateToResponse(a *database.PortAggregate) PortAggregateResponse {
	resp := PortAggregateResponse{
		Port:      a.Port,
		Protocol:  a.Protocol,
		Reviewed:  a.Reviewed,
		HostCount: a.HostCount,
	}
	if a.State.Valid {
		resp.State = &a.State.String
	}
	if a.ServiceName.Valid {
		resp.ServiceName = &a.ServiceName.String
	}
	if a.ServiceProduct.Valid {
		resp.ServiceProduct = &a.ServiceProduct.String
	}
	if a.ServiceVersion.Valid {
		resp.ServiceVersion = &a.ServiceVersion.String
	}
	return resp
}

func (h *Handlers) ListPorts(w http.ResponseWriter, r *http.Request) {
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
	aggregated, err := h.db.ListPortsAggregated(search, reviewed)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list ports")
		return
	}

	responses := make([]PortAggregateResponse, len(aggregated))
	for i, a := range aggregated {
		responses[i] = portAggregateToResponse(a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func portToResponse(p *database.Port) PortResponse {
	resp := PortResponse{
		ID:       p.ID.String(),
		HostID:   p.HostID.String(),
		Port:     p.Port,
		Protocol: p.Protocol,
		Reviewed: p.Reviewed,
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if p.State.Valid {
		resp.State = &p.State.String
	}
	if p.ServiceName.Valid {
		resp.ServiceName = &p.ServiceName.String
	}
	if p.ServiceProduct.Valid {
		resp.ServiceProduct = &p.ServiceProduct.String
	}
	if p.ServiceVersion.Valid {
		resp.ServiceVersion = &p.ServiceVersion.String
	}

	return resp
}

type PortDetailResponse struct {
	Port  PortResponse   `json:"port"`
	Hosts []HostResponse `json:"hosts"`
}

// PortDetailByNumberResponse is used when fetching by port number + protocol; reviewed is aggregated.
type PortDetailByNumberResponse struct {
	Port  PortInfoByNumber     `json:"port"`
	Hosts []HostPortRowResponse `json:"hosts"`
}

type PortInfoByNumber struct {
	Port           int     `json:"port"`
	Protocol       string  `json:"protocol"`
	State          *string `json:"state,omitempty"`
	ServiceName    *string `json:"service_name,omitempty"`
	ServiceProduct *string `json:"service_product,omitempty"`
	ServiceVersion *string `json:"service_version,omitempty"`
	Reviewed       bool    `json:"reviewed"` // true only when all host-ports are reviewed
}

type HostPortRowResponse struct {
	Host         HostResponse `json:"host"`
	PortID       string      `json:"port_id"`
	PortReviewed bool        `json:"port_reviewed"`
}

func (h *Handlers) GetPortByNumber(w http.ResponseWriter, r *http.Request) {
	portStr := chi.URLParam(r, "port")
	protocol := chi.URLParam(r, "protocol")
	if portStr == "" || protocol == "" {
		writeError(w, http.StatusBadRequest, "Missing port or protocol")
		return
	}
	portNum, err := strconv.Atoi(portStr)
	if err != nil || portNum < 1 || portNum > 65535 {
		writeError(w, http.StatusBadRequest, "Invalid port number")
		return
	}

	portInfo, hostRows, err := h.db.GetPortDetailByNumber(portNum, protocol)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get port detail")
		return
	}
	if portInfo == nil || len(hostRows) == 0 {
		writeError(w, http.StatusNotFound, "Port not found")
		return
	}

	// Aggregated reviewed: true only when all host-ports are reviewed
	allReviewed := true
	for _, row := range hostRows {
		if !row.PortReviewed {
			allReviewed = false
			break
		}
	}

	portResp := PortInfoByNumber{
		Port:     portInfo.Port,
		Protocol: portInfo.Protocol,
		Reviewed: allReviewed,
	}
	if portInfo.State.Valid {
		portResp.State = &portInfo.State.String
	}
	if portInfo.ServiceName.Valid {
		portResp.ServiceName = &portInfo.ServiceName.String
	}
	if portInfo.ServiceProduct.Valid {
		portResp.ServiceProduct = &portInfo.ServiceProduct.String
	}
	if portInfo.ServiceVersion.Valid {
		portResp.ServiceVersion = &portInfo.ServiceVersion.String
	}

	hostResponses := make([]HostPortRowResponse, len(hostRows))
	for i, row := range hostRows {
		hostResponses[i] = HostPortRowResponse{
			Host:         hostToResponse(row.Host),
			PortID:       row.PortID.String(),
			PortReviewed: row.PortReviewed,
		}
	}

	response := PortDetailByNumberResponse{
		Port:  portResp,
		Hosts: hostResponses,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handlers) SetPortReviewed(w http.ResponseWriter, r *http.Request) {
	portIDStr := chi.URLParam(r, "id")
	portID, err := uuid.Parse(portIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid port ID")
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

	if err := h.db.SetPortReviewed(portID, req.Reviewed, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update port")
		return
	}

	port, err := h.db.GetPortByID(portID)
	if err != nil || port == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(portToResponse(port))
}
