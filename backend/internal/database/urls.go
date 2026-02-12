package database

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type URL struct {
	ID                 uuid.UUID
	ProjectID          sql.NullString
	HostID             NullUUID
	PortID             NullUUID
	URL                string
	Path               string
	Method             string
	StatusCode         sql.NullInt64
	ContentLength      sql.NullInt64
	Words              sql.NullInt64
	Lines              sql.NullInt64
	FirstSeenScanFileID NullUUID
	LastSeenScanFileID  NullUUID
	Reviewed           bool
	ReviewedBy         NullUUID
	ReviewedAt         sql.NullTime
	ReviewNotes        sql.NullString
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (db *DB) UpsertURL(url, path, method string, statusCode, contentLength, words, lines *int, hostID, portID *uuid.UUID, scanFileID uuid.UUID) (*URL, error) {
	urlObj := &URL{
		URL:     url,
		Path:    path,
		Method:  method,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if statusCode != nil {
		urlObj.StatusCode = sql.NullInt64{Int64: int64(*statusCode), Valid: true}
	}
	if contentLength != nil {
		urlObj.ContentLength = sql.NullInt64{Int64: int64(*contentLength), Valid: true}
	}
	if words != nil {
		urlObj.Words = sql.NullInt64{Int64: int64(*words), Valid: true}
	}
	if lines != nil {
		urlObj.Lines = sql.NullInt64{Int64: int64(*lines), Valid: true}
	}
	if hostID != nil {
		urlObj.HostID = NullUUID{UUID: *hostID, Valid: true}
	}
	if portID != nil {
		urlObj.PortID = NullUUID{UUID: *portID, Valid: true}
	}
	urlObj.FirstSeenScanFileID = NullUUID{UUID: scanFileID, Valid: true}
	urlObj.LastSeenScanFileID = NullUUID{UUID: scanFileID, Valid: true}

	// Check if URL already exists
	existing, err := db.GetURLByURL(url)
	if err == nil && existing != nil {
		// Update last_seen_scan_file_id and host_id when provided (so URLs link to host and stay linked)
		var errScan error
		if hostID != nil {
			updateQuery := `
				UPDATE urls
				SET last_seen_scan_file_id = $1, host_id = $2, updated_at = CURRENT_TIMESTAMP
				WHERE url = $3
				RETURNING id, project_id, url, path, method, status_code, content_length, words, lines, host_id, port_id, first_seen_scan_file_id, last_seen_scan_file_id, reviewed, reviewed_by, reviewed_at, review_notes, created_at, updated_at
			`
			errScan = db.conn.QueryRow(updateQuery, scanFileID, *hostID, url).Scan(
				&urlObj.ID,
				&urlObj.ProjectID,
				&urlObj.URL,
				&urlObj.Path,
				&urlObj.Method,
				&urlObj.StatusCode,
				&urlObj.ContentLength,
				&urlObj.Words,
				&urlObj.Lines,
				&urlObj.HostID,
				&urlObj.PortID,
				&urlObj.FirstSeenScanFileID,
				&urlObj.LastSeenScanFileID,
				&urlObj.Reviewed,
				&urlObj.ReviewedBy,
				&urlObj.ReviewedAt,
				&urlObj.ReviewNotes,
				&urlObj.CreatedAt,
				&urlObj.UpdatedAt,
			)
		} else {
			updateQuery := `
				UPDATE urls
				SET last_seen_scan_file_id = $1, updated_at = CURRENT_TIMESTAMP
				WHERE url = $2
				RETURNING id, project_id, url, path, method, status_code, content_length, words, lines, host_id, port_id, first_seen_scan_file_id, last_seen_scan_file_id, reviewed, reviewed_by, reviewed_at, review_notes, created_at, updated_at
			`
			errScan = db.conn.QueryRow(updateQuery, scanFileID, url).Scan(
				&urlObj.ID,
				&urlObj.ProjectID,
				&urlObj.URL,
				&urlObj.Path,
				&urlObj.Method,
				&urlObj.StatusCode,
				&urlObj.ContentLength,
				&urlObj.Words,
				&urlObj.Lines,
				&urlObj.HostID,
				&urlObj.PortID,
				&urlObj.FirstSeenScanFileID,
				&urlObj.LastSeenScanFileID,
				&urlObj.Reviewed,
				&urlObj.ReviewedBy,
				&urlObj.ReviewedAt,
				&urlObj.ReviewNotes,
				&urlObj.CreatedAt,
				&urlObj.UpdatedAt,
			)
		}
		if errScan != nil {
			return nil, errScan
		}
		return urlObj, nil
	}

	query := `
		INSERT INTO urls (id, project_id, url, path, method, status_code, content_length, words, lines, host_id, port_id, first_seen_scan_file_id, last_seen_scan_file_id, created_at, updated_at)
		VALUES (gen_random_uuid(), NULL, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, project_id, url, path, method, status_code, content_length, words, lines, host_id, port_id, first_seen_scan_file_id, last_seen_scan_file_id, reviewed, reviewed_by, reviewed_at, review_notes, created_at, updated_at
	`

	var statusCodeVal, contentLengthVal, wordsVal, linesVal sql.NullInt64
	if statusCode != nil {
		statusCodeVal = sql.NullInt64{Int64: int64(*statusCode), Valid: true}
	}
	if contentLength != nil {
		contentLengthVal = sql.NullInt64{Int64: int64(*contentLength), Valid: true}
	}
	if words != nil {
		wordsVal = sql.NullInt64{Int64: int64(*words), Valid: true}
	}
	if lines != nil {
		linesVal = sql.NullInt64{Int64: int64(*lines), Valid: true}
	}

	var hostIDVal, portIDVal NullUUID
	if hostID != nil {
		hostIDVal = NullUUID{UUID: *hostID, Valid: true}
	}
	if portID != nil {
		portIDVal = NullUUID{UUID: *portID, Valid: true}
	}

	err = db.conn.QueryRow(
		query,
		url,
		path,
		method,
		statusCodeVal,
		contentLengthVal,
		wordsVal,
		linesVal,
		hostIDVal,
		portIDVal,
		scanFileID,
	).Scan(
		&urlObj.ID,
		&urlObj.ProjectID,
		&urlObj.URL,
		&urlObj.Path,
		&urlObj.Method,
		&urlObj.StatusCode,
		&urlObj.ContentLength,
		&urlObj.Words,
		&urlObj.Lines,
		&urlObj.HostID,
		&urlObj.PortID,
		&urlObj.FirstSeenScanFileID,
		&urlObj.LastSeenScanFileID,
		&urlObj.Reviewed,
		&urlObj.ReviewedBy,
		&urlObj.ReviewedAt,
		&urlObj.ReviewNotes,
		&urlObj.CreatedAt,
		&urlObj.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Conflict - URL already exists, try to get it
		return db.GetURLByURL(url)
	}
	if err != nil {
		return nil, err
	}

	return urlObj, nil
}

func (db *DB) GetURLByURL(url string) (*URL, error) {
	urlObj := &URL{}
	query := `
		SELECT id, project_id, url, path, method, status_code, content_length, words, lines, host_id, port_id, first_seen_scan_file_id, last_seen_scan_file_id, reviewed, reviewed_by, reviewed_at, review_notes, created_at, updated_at
		FROM urls
		WHERE url = $1
	`

	err := db.conn.QueryRow(query, url).Scan(
		&urlObj.ID,
		&urlObj.ProjectID,
		&urlObj.URL,
		&urlObj.Path,
		&urlObj.Method,
		&urlObj.StatusCode,
		&urlObj.ContentLength,
		&urlObj.Words,
		&urlObj.Lines,
		&urlObj.HostID,
		&urlObj.PortID,
		&urlObj.FirstSeenScanFileID,
		&urlObj.LastSeenScanFileID,
		&urlObj.Reviewed,
		&urlObj.ReviewedBy,
		&urlObj.ReviewedAt,
		&urlObj.ReviewNotes,
		&urlObj.CreatedAt,
		&urlObj.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return urlObj, nil
}

// URLWithHost is a URL plus the host it was discovered on (for list views).
type URLWithHost struct {
	URL
	HostIPAddress sql.NullString
	HostHostname  sql.NullString
}

// ListURLs optionally filters by search (url/path substring), reviewed, and status_code.
func (db *DB) ListURLs(search string, reviewed *bool, statusCode *int) ([]*URL, error) {
	query := `
		SELECT id, project_id, url, path, method, status_code, content_length, words, lines, host_id, port_id, first_seen_scan_file_id, last_seen_scan_file_id, reviewed, reviewed_by, reviewed_at, review_notes, created_at, updated_at
		FROM urls
	`
	var args []interface{}
	var where []string
	n := 1
	if search != "" {
		where = append(where, "(url ILIKE '%' || $"+strconv.Itoa(n)+" || '%' OR path ILIKE '%' || $"+strconv.Itoa(n)+" || '%')")
		args = append(args, search)
		n++
	}
	if reviewed != nil {
		where = append(where, "reviewed = $"+strconv.Itoa(n))
		args = append(args, *reviewed)
		n++
	}
	if statusCode != nil {
		where = append(where, "status_code = $"+strconv.Itoa(n))
		args = append(args, *statusCode)
	}
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []*URL
	for rows.Next() {
		urlObj := &URL{}
		err := rows.Scan(
			&urlObj.ID,
			&urlObj.ProjectID,
			&urlObj.URL,
			&urlObj.Path,
			&urlObj.Method,
			&urlObj.StatusCode,
			&urlObj.ContentLength,
			&urlObj.Words,
			&urlObj.Lines,
			&urlObj.HostID,
			&urlObj.PortID,
			&urlObj.FirstSeenScanFileID,
			&urlObj.LastSeenScanFileID,
			&urlObj.Reviewed,
			&urlObj.ReviewedBy,
			&urlObj.ReviewedAt,
			&urlObj.ReviewNotes,
			&urlObj.CreatedAt,
			&urlObj.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		urls = append(urls, urlObj)
	}

	return urls, rows.Err()
}

// ListURLsWithHost returns all URLs with host ip_address and hostname joined (for list view with host column).
func (db *DB) ListURLsWithHost() ([]*URLWithHost, error) {
	query := `
		SELECT u.id, u.project_id, u.url, u.path, u.method, u.status_code, u.content_length, u.words, u.lines,
			u.host_id, u.port_id, u.first_seen_scan_file_id, u.last_seen_scan_file_id, u.reviewed, u.reviewed_by, u.reviewed_at, u.review_notes, u.created_at, u.updated_at,
			h.ip_address AS host_ip_address, h.hostname AS host_hostname
		FROM urls u
		LEFT JOIN hosts h ON u.host_id = h.id
		ORDER BY u.created_at DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*URLWithHost
	for rows.Next() {
		row := &URLWithHost{}
		err := rows.Scan(
			&row.ID,
			&row.ProjectID,
			&row.URL,
			&row.Path,
			&row.Method,
			&row.StatusCode,
			&row.ContentLength,
			&row.Words,
			&row.Lines,
			&row.HostID,
			&row.PortID,
			&row.FirstSeenScanFileID,
			&row.LastSeenScanFileID,
			&row.Reviewed,
			&row.ReviewedBy,
			&row.ReviewedAt,
			&row.ReviewNotes,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.HostIPAddress,
			&row.HostHostname,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}

	return result, rows.Err()
}

func (db *DB) SetURLReviewed(id uuid.UUID, reviewed bool, reviewedBy uuid.UUID) error {
	if reviewed {
		_, err := db.conn.Exec(`UPDATE urls SET reviewed = true, reviewed_by = $1, reviewed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, reviewedBy, id)
		return err
	}
	_, err := db.conn.Exec(`UPDATE urls SET reviewed = false, reviewed_by = NULL, reviewed_at = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, id)
	return err
}

func (db *DB) GetURLByID(id uuid.UUID) (*URL, error) {
	u := &URL{}
	query := `
		SELECT id, project_id, url, path, method, status_code, content_length, words, lines, host_id, port_id, first_seen_scan_file_id, last_seen_scan_file_id, reviewed, reviewed_by, reviewed_at, review_notes, created_at, updated_at
		FROM urls
		WHERE id = $1
	`
	err := db.conn.QueryRow(query, id).Scan(
		&u.ID,
		&u.ProjectID,
		&u.URL,
		&u.Path,
		&u.Method,
		&u.StatusCode,
		&u.ContentLength,
		&u.Words,
		&u.Lines,
		&u.HostID,
		&u.PortID,
		&u.FirstSeenScanFileID,
		&u.LastSeenScanFileID,
		&u.Reviewed,
		&u.ReviewedBy,
		&u.ReviewedAt,
		&u.ReviewNotes,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (db *DB) GetURLsByHostID(hostID uuid.UUID) ([]*URL, error) {
	query := `
		SELECT id, project_id, url, path, method, status_code, content_length, words, lines, host_id, port_id, first_seen_scan_file_id, last_seen_scan_file_id, reviewed, reviewed_by, reviewed_at, review_notes, created_at, updated_at
		FROM urls
		WHERE host_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.conn.Query(query, hostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []*URL
	for rows.Next() {
		urlObj := &URL{}
		err := rows.Scan(
			&urlObj.ID,
			&urlObj.ProjectID,
			&urlObj.URL,
			&urlObj.Path,
			&urlObj.Method,
			&urlObj.StatusCode,
			&urlObj.ContentLength,
			&urlObj.Words,
			&urlObj.Lines,
			&urlObj.HostID,
			&urlObj.PortID,
			&urlObj.FirstSeenScanFileID,
			&urlObj.LastSeenScanFileID,
			&urlObj.Reviewed,
			&urlObj.ReviewedBy,
			&urlObj.ReviewedAt,
			&urlObj.ReviewNotes,
			&urlObj.CreatedAt,
			&urlObj.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		urls = append(urls, urlObj)
	}

	return urls, rows.Err()
}
