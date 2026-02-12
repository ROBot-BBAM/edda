package database

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Host struct {
	ID                 uuid.UUID
	ProjectID          sql.NullString
	IPAddress          string
	Hostname           sql.NullString
	OS                 sql.NullString
	FirstSeenScanFileID NullUUID
	LastSeenScanFileID  NullUUID
	Reviewed           bool
	ReviewedBy         NullUUID
	ReviewedAt         sql.NullTime
	ReviewNotes        sql.NullString
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (db *DB) UpsertHost(ipAddress, hostname, os string, scanFileID uuid.UUID) (*Host, error) {
	host := &Host{
		IPAddress: ipAddress,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if hostname != "" {
		host.Hostname = sql.NullString{String: hostname, Valid: true}
	}
	if os != "" {
		host.OS = sql.NullString{String: os, Valid: true}
	}
	host.FirstSeenScanFileID = NullUUID{UUID: scanFileID, Valid: true}
	host.LastSeenScanFileID = NullUUID{UUID: scanFileID, Valid: true}

	query := `
		INSERT INTO hosts (id, ip_address, hostname, os, first_seen_scan_file_id, last_seen_scan_file_id, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (ip_address) DO UPDATE
		SET 
			hostname = COALESCE(EXCLUDED.hostname, hosts.hostname),
			os = COALESCE(EXCLUDED.os, hosts.os),
			last_seen_scan_file_id = EXCLUDED.last_seen_scan_file_id,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, project_id, ip_address, hostname, os, first_seen_scan_file_id, last_seen_scan_file_id, reviewed, reviewed_by, reviewed_at, review_notes, created_at, updated_at
	`

	var hostnameVal, osVal sql.NullString
	if hostname != "" {
		hostnameVal = sql.NullString{String: hostname, Valid: true}
	}
	if os != "" {
		osVal = sql.NullString{String: os, Valid: true}
	}

	err := db.conn.QueryRow(
		query,
		ipAddress,
		hostnameVal,
		osVal,
		scanFileID,
	).Scan(
		&host.ID,
		&host.ProjectID,
		&host.IPAddress,
		&host.Hostname,
		&host.OS,
		&host.FirstSeenScanFileID,
		&host.LastSeenScanFileID,
		&host.Reviewed,
		&host.ReviewedBy,
		&host.ReviewedAt,
		&host.ReviewNotes,
		&host.CreatedAt,
		&host.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return host, nil
}

func (db *DB) GetHostByIP(ipAddress string) (*Host, error) {
	host := &Host{}
	query := `
		SELECT id, project_id, ip_address, hostname, os, first_seen_scan_file_id, last_seen_scan_file_id, reviewed, reviewed_by, reviewed_at, review_notes, created_at, updated_at
		FROM hosts
		WHERE ip_address = $1
	`

	err := db.conn.QueryRow(query, ipAddress).Scan(
		&host.ID,
		&host.ProjectID,
		&host.IPAddress,
		&host.Hostname,
		&host.OS,
		&host.FirstSeenScanFileID,
		&host.LastSeenScanFileID,
		&host.Reviewed,
		&host.ReviewedBy,
		&host.ReviewedAt,
		&host.ReviewNotes,
		&host.CreatedAt,
		&host.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return host, nil
}

func (db *DB) GetHostByHostname(hostname string) (*Host, error) {
	if hostname == "" {
		return nil, nil
	}
	host := &Host{}
	query := `
		SELECT id, project_id, ip_address, hostname, os, first_seen_scan_file_id, last_seen_scan_file_id, reviewed, reviewed_by, reviewed_at, review_notes, created_at, updated_at
		FROM hosts
		WHERE hostname = $1
		LIMIT 1
	`

	err := db.conn.QueryRow(query, hostname).Scan(
		&host.ID,
		&host.ProjectID,
		&host.IPAddress,
		&host.Hostname,
		&host.OS,
		&host.FirstSeenScanFileID,
		&host.LastSeenScanFileID,
		&host.Reviewed,
		&host.ReviewedBy,
		&host.ReviewedAt,
		&host.ReviewNotes,
		&host.CreatedAt,
		&host.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return host, nil
}

// HostReviewed returns true when all ports (or none) and all URLs (or none) for the host are reviewed.
func (db *DB) ComputeHostReviewed(hostID uuid.UUID) (bool, error) {
	var reviewed bool
	query := `
		SELECT (SELECT COALESCE(BOOL_AND(p.reviewed), true) FROM ports p WHERE p.host_id = $1)
		   AND (SELECT COALESCE(BOOL_AND(u.reviewed), true) FROM urls u WHERE u.host_id = $1)
	`
	err := db.conn.QueryRow(query, hostID).Scan(&reviewed)
	return reviewed, err
}

// ListHostsOpts optionally filters by search (IP/hostname substring) and reviewed (true/false).
func (db *DB) ListHosts(search string, reviewed *bool) ([]*Host, error) {
	inner := `
		SELECT id, project_id, ip_address, hostname, os, first_seen_scan_file_id, last_seen_scan_file_id,
			(SELECT COALESCE(BOOL_AND(p.reviewed), true) FROM ports p WHERE p.host_id = hosts.id)
				AND (SELECT COALESCE(BOOL_AND(u.reviewed), true) FROM urls u WHERE u.host_id = hosts.id) AS reviewed,
			reviewed_by, reviewed_at, review_notes, created_at, updated_at
		FROM hosts
	`
	var args []interface{}
	n := 1
	if search != "" {
		inner += " WHERE (ip_address ILIKE '%' || $" + strconv.Itoa(n) + " || '%' OR hostname ILIKE '%' || $" + strconv.Itoa(n) + " || '%')"
		args = append(args, search)
		n++
	}
	inner += " ORDER BY created_at DESC"

	query := inner
	if reviewed != nil {
		query = "SELECT * FROM (" + inner + ") t WHERE reviewed = $" + strconv.Itoa(n)
		args = append(args, *reviewed)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []*Host
	for rows.Next() {
		host := &Host{}
		err := rows.Scan(
			&host.ID,
			&host.ProjectID,
			&host.IPAddress,
			&host.Hostname,
			&host.OS,
			&host.FirstSeenScanFileID,
			&host.LastSeenScanFileID,
			&host.Reviewed,
			&host.ReviewedBy,
			&host.ReviewedAt,
			&host.ReviewNotes,
			&host.CreatedAt,
			&host.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}

	return hosts, rows.Err()
}

func (db *DB) SetHostReviewed(id uuid.UUID, reviewed bool, reviewedBy uuid.UUID) error {
	var query string
	if reviewed {
		query = `UPDATE hosts SET reviewed = true, reviewed_by = $1, reviewed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
		_, err := db.conn.Exec(query, reviewedBy, id)
		return err
	}
	query = `UPDATE hosts SET reviewed = false, reviewed_by = NULL, reviewed_at = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := db.conn.Exec(query, id)
	return err
}

func (db *DB) GetHostByID(id uuid.UUID) (*Host, error) {
	host := &Host{}
	query := `
		SELECT id, project_id, ip_address, hostname, os, first_seen_scan_file_id, last_seen_scan_file_id,
			(SELECT COALESCE(BOOL_AND(p.reviewed), true) FROM ports p WHERE p.host_id = hosts.id)
				AND (SELECT COALESCE(BOOL_AND(u.reviewed), true) FROM urls u WHERE u.host_id = hosts.id) AS reviewed,
			reviewed_by, reviewed_at, review_notes, created_at, updated_at
		FROM hosts
		WHERE id = $1
	`

	err := db.conn.QueryRow(query, id).Scan(
		&host.ID,
		&host.ProjectID,
		&host.IPAddress,
		&host.Hostname,
		&host.OS,
		&host.FirstSeenScanFileID,
		&host.LastSeenScanFileID,
		&host.Reviewed,
		&host.ReviewedBy,
		&host.ReviewedAt,
		&host.ReviewNotes,
		&host.CreatedAt,
		&host.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return host, nil
}
