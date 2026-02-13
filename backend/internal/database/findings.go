package database

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Finding is a vulnerability/finding linked to a host, port, and/or URL.
type Finding struct {
	ID          uuid.UUID
	HostID      NullUUID
	PortID      NullUUID
	URLID       NullUUID
	Title       string
	Severity    string
	Description sql.NullString
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateFinding inserts a new finding. At least one of hostID, portID, urlID must be non-nil.
func (db *DB) CreateFinding(hostID, portID, urlID *uuid.UUID, title, severity, description, status string) (*Finding, error) {
	if (hostID == nil || *hostID == uuid.Nil) && (portID == nil || *portID == uuid.Nil) && (urlID == nil || *urlID == uuid.Nil) {
		return nil, sql.ErrNoRows
	}
	var hostVal, portVal, urlVal NullUUID
	if hostID != nil && *hostID != uuid.Nil {
		hostVal = NullUUID{UUID: *hostID, Valid: true}
	}
	if portID != nil && *portID != uuid.Nil {
		portVal = NullUUID{UUID: *portID, Valid: true}
	}
	if urlID != nil && *urlID != uuid.Nil {
		urlVal = NullUUID{UUID: *urlID, Valid: true}
	}
	if severity == "" {
		severity = "medium"
	}
	if status == "" {
		status = "open"
	}
	var descVal sql.NullString
	if description != "" {
		descVal = sql.NullString{String: description, Valid: true}
	}
	query := `
		INSERT INTO findings (host_id, port_id, url_id, title, severity, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, host_id, port_id, url_id, title, severity, description, status, created_at, updated_at
	`
	f := &Finding{}
	err := db.conn.QueryRow(query, hostVal, portVal, urlVal, title, severity, descVal, status).Scan(
		&f.ID, &f.HostID, &f.PortID, &f.URLID, &f.Title, &f.Severity, &f.Description, &f.Status, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// ListFindingsOpts filters by host, port(s), URL, severity, and/or status. Empty severity/status means no filter.
func (db *DB) ListFindings(hostID *uuid.UUID, portIDs []uuid.UUID, urlID *uuid.UUID, severity, status string) ([]*Finding, error) {
	query := `SELECT id, host_id, port_id, url_id, title, severity, description, status, created_at, updated_at FROM findings WHERE 1=1`
	var args []interface{}
	n := 1
	if hostID != nil && *hostID != uuid.Nil {
		query += " AND host_id = $" + strconv.Itoa(n)
		args = append(args, *hostID)
		n++
	}
	if len(portIDs) > 0 {
		ids := make([]string, len(portIDs))
		for i, id := range portIDs {
			ids[i] = id.String()
		}
		query += " AND port_id = ANY(SELECT unnest($" + strconv.Itoa(n) + "::text[])::uuid)"
		args = append(args, pq.Array(ids))
		n++
	}
	if urlID != nil && *urlID != uuid.Nil {
		query += " AND url_id = $" + strconv.Itoa(n)
		args = append(args, *urlID)
		n++
	}
	if severity != "" {
		query += " AND severity = $" + strconv.Itoa(n)
		args = append(args, severity)
		n++
	}
	if status != "" {
		query += " AND status = $" + strconv.Itoa(n)
		args = append(args, status)
		n++
	}
	query += " ORDER BY created_at DESC"
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Finding
	for rows.Next() {
		f := &Finding{}
		err := rows.Scan(&f.ID, &f.HostID, &f.PortID, &f.URLID, &f.Title, &f.Severity, &f.Description, &f.Status, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// GetFindingByID returns a finding by ID, or nil if not found.
func (db *DB) GetFindingByID(id uuid.UUID) (*Finding, error) {
	f := &Finding{}
	query := `SELECT id, host_id, port_id, url_id, title, severity, description, status, created_at, updated_at FROM findings WHERE id = $1`
	err := db.conn.QueryRow(query, id).Scan(
		&f.ID, &f.HostID, &f.PortID, &f.URLID, &f.Title, &f.Severity, &f.Description, &f.Status, &f.CreatedAt, &f.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

// DeleteFinding deletes a finding by ID. Returns nil if not found.
func (db *DB) DeleteFinding(id uuid.UUID) error {
	res, err := db.conn.Exec(`DELETE FROM findings WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateFinding updates a finding by ID. All fields are updated; pass existing values for fields not changing.
func (db *DB) UpdateFinding(id uuid.UUID, hostID, portID, urlID NullUUID, title, severity, description, status string) (*Finding, error) {
	var descVal sql.NullString
	if description != "" {
		descVal = sql.NullString{String: description, Valid: true}
	}
	query := `UPDATE findings SET host_id = $1, port_id = $2, url_id = $3, title = $4, severity = $5, description = $6, status = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8 RETURNING id, host_id, port_id, url_id, title, severity, description, status, created_at, updated_at`
	f := &Finding{}
	err := db.conn.QueryRow(query, hostID, portID, urlID, title, severity, descVal, status, id).Scan(
		&f.ID, &f.HostID, &f.PortID, &f.URLID, &f.Title, &f.Severity, &f.Description, &f.Status, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// CountFindingsByHostIDs returns a map of host_id -> count of findings for each host.
func (db *DB) CountFindingsByHostIDs(hostIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(hostIDs) == 0 {
		return nil, nil
	}
	ids := make([]string, len(hostIDs))
	for i, id := range hostIDs {
		ids[i] = id.String()
	}
	query := `SELECT host_id, COUNT(*) FROM findings WHERE host_id = ANY(SELECT unnest($1::text[])::uuid) GROUP BY host_id`
	rows, err := db.conn.Query(query, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[uuid.UUID]int)
	for rows.Next() {
		var id uuid.UUID
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}
		out[id] = count
	}
	return out, rows.Err()
}

// CountFindingsByPortIDs returns a map of port_id -> count of findings for each port.
func (db *DB) CountFindingsByPortIDs(portIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(portIDs) == 0 {
		return nil, nil
	}
	ids := make([]string, len(portIDs))
	for i, id := range portIDs {
		ids[i] = id.String()
	}
	query := `SELECT port_id, COUNT(*) FROM findings WHERE port_id = ANY(SELECT unnest($1::text[])::uuid) GROUP BY port_id`
	rows, err := db.conn.Query(query, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[uuid.UUID]int)
	for rows.Next() {
		var id uuid.UUID
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}
		out[id] = count
	}
	return out, rows.Err()
}

// CountFindingsByURLIDs returns a map of url_id -> count of findings for each URL.
func (db *DB) CountFindingsByURLIDs(urlIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(urlIDs) == 0 {
		return nil, nil
	}
	ids := make([]string, len(urlIDs))
	for i, id := range urlIDs {
		ids[i] = id.String()
	}
	query := `SELECT url_id, COUNT(*) FROM findings WHERE url_id = ANY(SELECT unnest($1::text[])::uuid) GROUP BY url_id`
	rows, err := db.conn.Query(query, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[uuid.UUID]int)
	for rows.Next() {
		var id uuid.UUID
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}
		out[id] = count
	}
	return out, rows.Err()
}

// GetFindingsSummary returns counts by severity and total open count.
func (db *DB) GetFindingsSummary() (bySeverity map[string]int, openCount int, err error) {
	bySeverity = make(map[string]int)
	query := `SELECT severity, COUNT(*) FROM findings GROUP BY severity`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var sev string
		var n int
		if err := rows.Scan(&sev, &n); err != nil {
			return nil, 0, err
		}
		bySeverity[sev] = n
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	err = db.conn.QueryRow(`SELECT COUNT(*) FROM findings WHERE status = $1`, "open").Scan(&openCount)
	if err != nil {
		return nil, 0, err
	}
	return bySeverity, openCount, nil
}

// SearchFindings returns findings whose title or description matches the query (ILIKE), limited to limit.
func (db *DB) SearchFindings(q string, limit int) ([]*Finding, error) {
	if q == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}
	query := `SELECT id, host_id, port_id, url_id, title, severity, description, status, created_at, updated_at
		FROM findings WHERE title ILIKE '%' || $1 || '%' OR (description IS NOT NULL AND description ILIKE '%' || $1 || '%')
		ORDER BY created_at DESC LIMIT ` + strconv.Itoa(limit)
	rows, err := db.conn.Query(query, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Finding
	for rows.Next() {
		f := &Finding{}
		err := rows.Scan(&f.ID, &f.HostID, &f.PortID, &f.URLID, &f.Title, &f.Severity, &f.Description, &f.Status, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}
