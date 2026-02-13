package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NarrativeExport is the full engagement export for AI/narrative generation.
type NarrativeExport struct {
	ExportedAt string                 `json:"exported_at"`
	Summary    NarrativeExportSummary  `json:"summary"`
	Hosts      []NarrativeHost         `json:"hosts"`
	Ports      []NarrativePort         `json:"ports"`
	URLs       []NarrativeURL          `json:"urls"`
	Findings   []NarrativeFinding      `json:"findings"`
	Notes      []NarrativeNote         `json:"notes"`
}

type NarrativeExportSummary struct {
	Hosts    int `json:"hosts"`
	Ports    int `json:"ports"`
	URLs     int `json:"urls"`
	Findings int `json:"findings"`
	Notes    int `json:"notes"`
}

type NarrativeHost struct {
	ID         string  `json:"id"`
	IPAddress  string  `json:"ip_address"`
	Hostname   *string `json:"hostname,omitempty"`
	OS         *string `json:"os,omitempty"`
	Reviewed   bool    `json:"reviewed"`
	CreatedAt  string  `json:"created_at"`
}

type NarrativePort struct {
	ID          string  `json:"id"`
	HostID      string  `json:"host_id"`
	HostIP      string  `json:"host_ip"`
	Port        int     `json:"port"`
	Protocol    string  `json:"protocol"`
	State       *string `json:"state,omitempty"`
	Service     *string `json:"service_name,omitempty"`
	Product     *string `json:"service_product,omitempty"`
	Version     *string `json:"service_version,omitempty"`
	Reviewed    bool    `json:"reviewed"`
	CreatedAt   string  `json:"created_at"`
}

type NarrativeURL struct {
	ID            string  `json:"id"`
	HostID        *string `json:"host_id,omitempty"`
	HostDisplay   string  `json:"host_display,omitempty"`
	URL           string  `json:"url"`
	Path          string  `json:"path"`
	Method        string  `json:"method"`
	StatusCode    *int    `json:"status_code,omitempty"`
	ContentLength *int64  `json:"content_length,omitempty"`
	Words         *int64  `json:"words,omitempty"`
	Lines         *int64  `json:"lines,omitempty"`
	Reviewed      bool    `json:"reviewed"`
	CreatedAt     string  `json:"created_at"`
}

type NarrativeFinding struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Severity    string  `json:"severity"`
	Status      string  `json:"status"`
	Description string  `json:"description,omitempty"`
	HostDisplay string  `json:"host_display,omitempty"`
	PortDisplay string  `json:"port_display,omitempty"`
	URLDisplay  string  `json:"url_display,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type NarrativeNote struct {
	ID         string `json:"id"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
	TargetType string `json:"target_type"` // "host" | "port" | "url"
	TargetID   string `json:"target_id"`
	TargetLabel string `json:"target_label"` // e.g. "Host 192.168.1.1", "Port 443/tcp on 192.168.1.1"
}

func (h *Handlers) ExportNarrative(w http.ResponseWriter, r *http.Request) {
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "json"
	}
	if format != "json" && format != "csv" {
		format = "json"
	}

	hosts, err := h.db.ListHosts("", nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list hosts")
		return
	}
	ports, err := h.db.ListPorts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list ports")
		return
	}
	urls, err := h.db.ListURLs("", nil, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list URLs")
		return
	}
	findings, err := h.db.ListFindings(nil, nil, nil, "", "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list findings")
		return
	}
	notes, err := h.db.ListAllNotes()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list notes")
		return
	}

	hostByID := make(map[uuid.UUID]*narrativeHostInfo)
	for _, host := range hosts {
		var hostname *string
		if host.Hostname.Valid {
			hostname = &host.Hostname.String
		}
		hostByID[host.ID] = &narrativeHostInfo{IP: host.IPAddress, Hostname: hostname}
	}

	portByID := make(map[uuid.UUID]*narrativePortInfo)
	for _, p := range ports {
		state, svc, product, version := (*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil)
		if p.State.Valid {
			state = &p.State.String
		}
		if p.ServiceName.Valid {
			svc = &p.ServiceName.String
		}
		if p.ServiceProduct.Valid {
			product = &p.ServiceProduct.String
		}
		if p.ServiceVersion.Valid {
			version = &p.ServiceVersion.String
		}
		hostIP := ""
		if h, ok := hostByID[p.HostID]; ok {
			hostIP = h.IP
		}
		portByID[p.ID] = &narrativePortInfo{HostIP: hostIP, Port: p.Port, Protocol: p.Protocol, State: state, Service: svc, Product: product, Version: version}
	}

	urlByID := make(map[uuid.UUID]*narrativeURLInfo)
	for _, u := range urls {
		hostDisp := ""
		if u.HostID.Valid {
			if h, ok := hostByID[u.HostID.UUID]; ok {
				if h.Hostname != nil && *h.Hostname != "" {
					hostDisp = *h.Hostname + " (" + h.IP + ")"
				} else {
					hostDisp = h.IP
				}
			}
		}
		var statusCode *int
		var contentLength, words, lines *int64
		if u.StatusCode.Valid {
			sc := int(u.StatusCode.Int64)
			statusCode = &sc
		}
		if u.ContentLength.Valid {
			contentLength = &u.ContentLength.Int64
		}
		if u.Words.Valid {
			words = &u.Words.Int64
		}
		if u.Lines.Valid {
			lines = &u.Lines.Int64
		}
		urlByID[u.ID] = &narrativeURLInfo{HostDisplay: hostDisp, URL: u.URL, Path: u.Path, Method: u.Method, StatusCode: statusCode, ContentLength: contentLength, Words: words, Lines: lines}
	}

	out := NarrativeExport{
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Summary: NarrativeExportSummary{
			Hosts:    len(hosts),
			Ports:    len(ports),
			URLs:     len(urls),
			Findings: len(findings),
			Notes:    len(notes),
		},
		Hosts:    make([]NarrativeHost, 0, len(hosts)),
		Ports:    make([]NarrativePort, 0, len(ports)),
		URLs:     make([]NarrativeURL, 0, len(urls)),
		Findings: make([]NarrativeFinding, 0, len(findings)),
		Notes:    make([]NarrativeNote, 0, len(notes)),
	}

	for _, host := range hosts {
		hostname, osVal := (*string)(nil), (*string)(nil)
		if host.Hostname.Valid {
			hostname = &host.Hostname.String
		}
		if host.OS.Valid {
			osVal = &host.OS.String
		}
		out.Hosts = append(out.Hosts, NarrativeHost{
			ID:        host.ID.String(),
			IPAddress: host.IPAddress,
			Hostname:  hostname,
			OS:        osVal,
			Reviewed:  host.Reviewed,
			CreatedAt: host.CreatedAt.Format(time.RFC3339),
		})
	}

	for _, p := range ports {
		hostIP := ""
		if h, ok := hostByID[p.HostID]; ok {
			hostIP = h.IP
		}
		state, svc, product, version := (*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil)
		if p.State.Valid {
			state = &p.State.String
		}
		if p.ServiceName.Valid {
			svc = &p.ServiceName.String
		}
		if p.ServiceProduct.Valid {
			product = &p.ServiceProduct.String
		}
		if p.ServiceVersion.Valid {
			version = &p.ServiceVersion.String
		}
		out.Ports = append(out.Ports, NarrativePort{
			ID:        p.ID.String(),
			HostID:    p.HostID.String(),
			HostIP:    hostIP,
			Port:      p.Port,
			Protocol:  p.Protocol,
			State:     state,
			Service:   svc,
			Product:   product,
			Version:   version,
			Reviewed:  p.Reviewed,
			CreatedAt: p.CreatedAt.Format(time.RFC3339),
		})
	}

	for _, u := range urls {
		hostDisp := ""
		var hostIDPtr *string
		if u.HostID.Valid {
			hostIDPtr = strPtr(u.HostID.UUID.String())
			if h, ok := hostByID[u.HostID.UUID]; ok {
				if h.Hostname != nil && *h.Hostname != "" {
					hostDisp = *h.Hostname + " (" + h.IP + ")"
				} else {
					hostDisp = h.IP
				}
			}
		}
		var statusCode *int
		var contentLength, words, lines *int64
		if u.StatusCode.Valid {
			sc := int(u.StatusCode.Int64)
			statusCode = &sc
		}
		if u.ContentLength.Valid {
			contentLength = &u.ContentLength.Int64
		}
		if u.Words.Valid {
			words = &u.Words.Int64
		}
		if u.Lines.Valid {
			lines = &u.Lines.Int64
		}
		out.URLs = append(out.URLs, NarrativeURL{
			ID:            u.ID.String(),
			HostID:        hostIDPtr,
			HostDisplay:   hostDisp,
			URL:           u.URL,
			Path:          u.Path,
			Method:        u.Method,
			StatusCode:    statusCode,
			ContentLength: contentLength,
			Words:         words,
			Lines:         lines,
			Reviewed:      u.Reviewed,
			CreatedAt:     u.CreatedAt.Format(time.RFC3339),
		})
	}

	for _, f := range findings {
		hostDisp, portDisp, urlDisp := "", "", ""
		if f.HostID.Valid {
			if host, err := h.db.GetHostByID(f.HostID.UUID); err == nil && host != nil {
				if host.Hostname.Valid && host.Hostname.String != "" {
					hostDisp = host.Hostname.String + " (" + host.IPAddress + ")"
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
			}
		}
		desc := ""
		if f.Description.Valid {
			desc = f.Description.String
		}
		out.Findings = append(out.Findings, NarrativeFinding{
			ID:          f.ID.String(),
			Title:       f.Title,
			Severity:    f.Severity,
			Status:      f.Status,
			Description: desc,
			HostDisplay: hostDisp,
			PortDisplay: portDisp,
			URLDisplay:  urlDisp,
			CreatedAt:   f.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   f.UpdatedAt.Format(time.RFC3339),
		})
	}

	for _, n := range notes {
		targetType := ""
		targetID := ""
		targetLabel := ""
		if n.HostID.Valid {
			targetType = "host"
			targetID = n.HostID.UUID.String()
			if host, err := h.db.GetHostByID(n.HostID.UUID); err == nil && host != nil {
				if host.Hostname.Valid && host.Hostname.String != "" {
					targetLabel = "Host " + host.Hostname.String + " (" + host.IPAddress + ")"
				} else {
					targetLabel = "Host " + host.IPAddress
				}
			} else {
				targetLabel = "Host " + targetID
			}
		} else if n.PortID.Valid {
			targetType = "port"
			targetID = n.PortID.UUID.String()
			if info, ok := portByID[n.PortID.UUID]; ok {
				targetLabel = "Port " + strconv.Itoa(info.Port) + "/" + info.Protocol + " on " + info.HostIP
			} else {
				targetLabel = "Port " + targetID
			}
		} else if n.URLID.Valid {
			targetType = "url"
			targetID = n.URLID.UUID.String()
			if info, ok := urlByID[n.URLID.UUID]; ok {
				targetLabel = "URL " + info.Path
				if len(targetLabel) > 80 {
					targetLabel = targetLabel[:77] + "..."
				}
			} else {
				targetLabel = "URL " + targetID
			}
		}
		out.Notes = append(out.Notes, NarrativeNote{
			ID:          n.ID.String(),
			Content:     n.Content,
			CreatedAt:   n.CreatedAt.Format(time.RFC3339),
			TargetType:  targetType,
			TargetID:    targetID,
			TargetLabel: targetLabel,
		})
	}

	if format == "csv" {
		writeExportZip(w, &out)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="edda-narrative-export.json"`)
	json.NewEncoder(w).Encode(out)
}

func writeExportZip(w http.ResponseWriter, out *NarrativeExport) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	date := time.Now().UTC().Format("2006-01-02")

	// hosts.csv
	{
		var b bytes.Buffer
		cw := csv.NewWriter(&b)
		_ = cw.Write([]string{"id", "ip_address", "hostname", "os", "reviewed", "created_at"})
		for _, h := range out.Hosts {
			hn, osVal := "", ""
			if h.Hostname != nil {
				hn = *h.Hostname
			}
			if h.OS != nil {
				osVal = *h.OS
			}
			_ = cw.Write([]string{h.ID, h.IPAddress, hn, osVal, strconv.FormatBool(h.Reviewed), h.CreatedAt})
		}
		cw.Flush()
		f, _ := zw.Create("hosts.csv")
		_, _ = f.Write(b.Bytes())
	}
	// ports.csv
	{
		var b bytes.Buffer
		cw := csv.NewWriter(&b)
		_ = cw.Write([]string{"id", "host_id", "host_ip", "port", "protocol", "state", "service_name", "service_product", "service_version", "reviewed", "created_at"})
		for _, p := range out.Ports {
			state, svc, product, version := "", "", "", ""
			if p.State != nil {
				state = *p.State
			}
			if p.Service != nil {
				svc = *p.Service
			}
			if p.Product != nil {
				product = *p.Product
			}
			if p.Version != nil {
				version = *p.Version
			}
			_ = cw.Write([]string{p.ID, p.HostID, p.HostIP, strconv.Itoa(p.Port), p.Protocol, state, svc, product, version, strconv.FormatBool(p.Reviewed), p.CreatedAt})
		}
		cw.Flush()
		f, _ := zw.Create("ports.csv")
		_, _ = f.Write(b.Bytes())
	}
	// urls.csv
	{
		var b bytes.Buffer
		cw := csv.NewWriter(&b)
		_ = cw.Write([]string{"id", "host_id", "host_display", "url", "path", "method", "status_code", "content_length", "words", "lines", "reviewed", "created_at"})
		for _, u := range out.URLs {
			hostID := ""
			if u.HostID != nil {
				hostID = *u.HostID
			}
			sc, cl, words, lines := "", "", "", ""
			if u.StatusCode != nil {
				sc = strconv.Itoa(*u.StatusCode)
			}
			if u.ContentLength != nil {
				cl = strconv.FormatInt(*u.ContentLength, 10)
			}
			if u.Words != nil {
				words = strconv.FormatInt(*u.Words, 10)
			}
			if u.Lines != nil {
				lines = strconv.FormatInt(*u.Lines, 10)
			}
			_ = cw.Write([]string{u.ID, hostID, u.HostDisplay, u.URL, u.Path, u.Method, sc, cl, words, lines, strconv.FormatBool(u.Reviewed), u.CreatedAt})
		}
		cw.Flush()
		f, _ := zw.Create("urls.csv")
		_, _ = f.Write(b.Bytes())
	}
	// findings.csv
	{
		var b bytes.Buffer
		cw := csv.NewWriter(&b)
		_ = cw.Write([]string{"id", "title", "severity", "status", "description", "host_display", "port_display", "url_display", "created_at", "updated_at"})
		for _, f := range out.Findings {
			_ = cw.Write([]string{f.ID, f.Title, f.Severity, f.Status, f.Description, f.HostDisplay, f.PortDisplay, f.URLDisplay, f.CreatedAt, f.UpdatedAt})
		}
		cw.Flush()
		f, _ := zw.Create("findings.csv")
		_, _ = f.Write(b.Bytes())
	}
	// notes.csv
	{
		var b bytes.Buffer
		cw := csv.NewWriter(&b)
		_ = cw.Write([]string{"id", "content", "created_at", "target_type", "target_id", "target_label"})
		for _, n := range out.Notes {
			_ = cw.Write([]string{n.ID, n.Content, n.CreatedAt, n.TargetType, n.TargetID, n.TargetLabel})
		}
		cw.Flush()
		f, _ := zw.Create("notes.csv")
		_, _ = f.Write(b.Bytes())
	}
	_ = zw.Close()

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="edda-export-`+date+`.zip"`)
	_, _ = w.Write(buf.Bytes())
}

type narrativeHostInfo struct {
	IP       string
	Hostname *string
}

type narrativePortInfo struct {
	HostIP   string
	Port     int
	Protocol string
	State    *string
	Service  *string
	Product  *string
	Version  *string
}

type narrativeURLInfo struct {
	HostDisplay   string
	URL           string
	Path          string
	Method        string
	StatusCode    *int
	ContentLength *int64
	Words         *int64
	Lines         *int64
}

func strPtr(s string) *string {
	return &s
}
