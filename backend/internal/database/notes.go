package database

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Note is a note attached to a host, port, or URL.
type Note struct {
	ID        uuid.UUID
	HostID    NullUUID
	PortID    NullUUID
	URLID     NullUUID
	Content   string
	CreatedBy NullUUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateNote inserts a note. Exactly one of hostID, portID, urlID must be non-nil.
func (db *DB) CreateNote(hostID, portID, urlID *uuid.UUID, content string, createdBy *uuid.UUID) (*Note, error) {
	var hostVal, portVal, urlVal NullUUID
	n := 0
	if hostID != nil && *hostID != uuid.Nil {
		hostVal = NullUUID{UUID: *hostID, Valid: true}
		n++
	}
	if portID != nil && *portID != uuid.Nil {
		portVal = NullUUID{UUID: *portID, Valid: true}
		n++
	}
	if urlID != nil && *urlID != uuid.Nil {
		urlVal = NullUUID{UUID: *urlID, Valid: true}
		n++
	}
	if n != 1 {
		return nil, sql.ErrNoRows
	}
	var createdByVal NullUUID
	if createdBy != nil && *createdBy != uuid.Nil {
		createdByVal = NullUUID{UUID: *createdBy, Valid: true}
	}
	query := `INSERT INTO notes (host_id, port_id, url_id, content, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, host_id, port_id, url_id, content, created_by, created_at, updated_at`
	note := &Note{}
	err := db.conn.QueryRow(query, hostVal, portVal, urlVal, content, createdByVal).Scan(
		&note.ID, &note.HostID, &note.PortID, &note.URLID, &note.Content, &note.CreatedBy, &note.CreatedAt, &note.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return note, nil
}

// ListNotes returns notes for the given entity (exactly one of hostID, portID, urlID must be set).
func (db *DB) ListNotes(hostID, portID, urlID *uuid.UUID) ([]*Note, error) {
	query := `SELECT id, host_id, port_id, url_id, content, created_by, created_at, updated_at FROM notes WHERE 1=1`
	var args []interface{}
	n := 1
	if hostID != nil && *hostID != uuid.Nil {
		query += " AND host_id = $" + strconv.Itoa(n)
		args = append(args, *hostID)
		n++
	}
	if portID != nil && *portID != uuid.Nil {
		query += " AND port_id = $" + strconv.Itoa(n)
		args = append(args, *portID)
		n++
	}
	if urlID != nil && *urlID != uuid.Nil {
		query += " AND url_id = $" + strconv.Itoa(n)
		args = append(args, *urlID)
		n++
	}
	query += " ORDER BY created_at DESC"
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Note
	for rows.Next() {
		note := &Note{}
		err := rows.Scan(&note.ID, &note.HostID, &note.PortID, &note.URLID, &note.Content, &note.CreatedBy, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, err
		}
		out = append(out, note)
	}
	return out, rows.Err()
}

// ListAllNotes returns all notes, ordered by created_at DESC.
func (db *DB) ListAllNotes() ([]*Note, error) {
	query := `SELECT id, host_id, port_id, url_id, content, created_by, created_at, updated_at FROM notes ORDER BY created_at DESC`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Note
	for rows.Next() {
		note := &Note{}
		err := rows.Scan(&note.ID, &note.HostID, &note.PortID, &note.URLID, &note.Content, &note.CreatedBy, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, err
		}
		out = append(out, note)
	}
	return out, rows.Err()
}

// CountNotesByHostIDs returns host_id -> count.
func (db *DB) CountNotesByHostIDs(hostIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(hostIDs) == 0 {
		return nil, nil
	}
	ids := make([]string, len(hostIDs))
	for i, id := range hostIDs {
		ids[i] = id.String()
	}
	query := `SELECT host_id, COUNT(*) FROM notes WHERE host_id = ANY(SELECT unnest($1::text[])::uuid) GROUP BY host_id`
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

// CountNotesByPortIDs returns port_id -> count.
func (db *DB) CountNotesByPortIDs(portIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(portIDs) == 0 {
		return nil, nil
	}
	ids := make([]string, len(portIDs))
	for i, id := range portIDs {
		ids[i] = id.String()
	}
	query := `SELECT port_id, COUNT(*) FROM notes WHERE port_id = ANY(SELECT unnest($1::text[])::uuid) GROUP BY port_id`
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

// CountNotesByURLIDs returns url_id -> count.
func (db *DB) CountNotesByURLIDs(urlIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(urlIDs) == 0 {
		return nil, nil
	}
	ids := make([]string, len(urlIDs))
	for i, id := range urlIDs {
		ids[i] = id.String()
	}
	query := `SELECT url_id, COUNT(*) FROM notes WHERE url_id = ANY(SELECT unnest($1::text[])::uuid) GROUP BY url_id`
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

// GetLatestNotePreview returns the most recent note's content truncated to maxLen for the given entity.
func (db *DB) GetLatestNotePreview(hostID, portID, urlID *uuid.UUID, maxLen int) (string, bool) {
	query := `SELECT content FROM notes WHERE 1=1`
	var args []interface{}
	n := 1
	if hostID != nil && *hostID != uuid.Nil {
		query += " AND host_id = $" + strconv.Itoa(n)
		args = append(args, *hostID)
		n++
	}
	if portID != nil && *portID != uuid.Nil {
		query += " AND port_id = $" + strconv.Itoa(n)
		args = append(args, *portID)
		n++
	}
	if urlID != nil && *urlID != uuid.Nil {
		query += " AND url_id = $" + strconv.Itoa(n)
		args = append(args, *urlID)
		n++
	}
	query += " ORDER BY created_at DESC LIMIT 1"
	var content string
	err := db.conn.QueryRow(query, args...).Scan(&content)
	if err != nil {
		return "", false
	}
	if maxLen > 0 && len(content) > maxLen {
		content = content[:maxLen] + "..."
	}
	return content, true
}
