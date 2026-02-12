package database

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Port struct {
	ID                 uuid.UUID
	HostID             uuid.UUID
	Port               int
	Protocol           string
	State              sql.NullString
	ServiceName        sql.NullString
	ServiceProduct     sql.NullString
	ServiceVersion     sql.NullString
	FirstSeenScanFileID NullUUID
	LastSeenScanFileID  NullUUID
	Reviewed           bool
	ReviewedBy         NullUUID
	ReviewedAt         sql.NullTime
	ReviewNotes        sql.NullString
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (db *DB) UpsertPort(hostID uuid.UUID, port int, protocol, state, serviceName, serviceProduct, serviceVersion string, scanFileID uuid.UUID) (*Port, error) {
	portObj := &Port{
		HostID:   hostID,
		Port:     port,
		Protocol: protocol,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if state != "" {
		portObj.State = sql.NullString{String: state, Valid: true}
	}
	if serviceName != "" {
		portObj.ServiceName = sql.NullString{String: serviceName, Valid: true}
	}
	if serviceProduct != "" {
		portObj.ServiceProduct = sql.NullString{String: serviceProduct, Valid: true}
	}
	if serviceVersion != "" {
		portObj.ServiceVersion = sql.NullString{String: serviceVersion, Valid: true}
	}
	portObj.FirstSeenScanFileID = NullUUID{UUID: scanFileID, Valid: true}
	portObj.LastSeenScanFileID = NullUUID{UUID: scanFileID, Valid: true}

	query := `
		INSERT INTO ports (id, host_id, port, protocol, state, service_name, service_product, service_version, first_seen_scan_file_id, last_seen_scan_file_id, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (host_id, port, protocol) DO UPDATE
		SET 
			state = COALESCE(EXCLUDED.state, ports.state),
			service_name = COALESCE(EXCLUDED.service_name, ports.service_name),
			service_product = COALESCE(EXCLUDED.service_product, ports.service_product),
			service_version = COALESCE(EXCLUDED.service_version, ports.service_version),
			last_seen_scan_file_id = EXCLUDED.last_seen_scan_file_id,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, host_id, port, protocol, state, service_name, service_product, service_version, first_seen_scan_file_id, last_seen_scan_file_id, reviewed, reviewed_by, reviewed_at, review_notes, created_at, updated_at
	`

	var stateVal, serviceNameVal, serviceProductVal, serviceVersionVal sql.NullString
	if state != "" {
		stateVal = sql.NullString{String: state, Valid: true}
	}
	if serviceName != "" {
		serviceNameVal = sql.NullString{String: serviceName, Valid: true}
	}
	if serviceProduct != "" {
		serviceProductVal = sql.NullString{String: serviceProduct, Valid: true}
	}
	if serviceVersion != "" {
		serviceVersionVal = sql.NullString{String: serviceVersion, Valid: true}
	}

	err := db.conn.QueryRow(
		query,
		hostID,
		port,
		protocol,
		stateVal,
		serviceNameVal,
		serviceProductVal,
		serviceVersionVal,
		scanFileID,
	).Scan(
		&portObj.ID,
		&portObj.HostID,
		&portObj.Port,
		&portObj.Protocol,
		&portObj.State,
		&portObj.ServiceName,
		&portObj.ServiceProduct,
		&portObj.ServiceVersion,
		&portObj.FirstSeenScanFileID,
		&portObj.LastSeenScanFileID,
		&portObj.Reviewed,
		&portObj.ReviewedBy,
		&portObj.ReviewedAt,
		&portObj.ReviewNotes,
		&portObj.CreatedAt,
		&portObj.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return portObj, nil
}

func (db *DB) ListPorts() ([]*Port, error) {
	query := `
		SELECT id, host_id, port, protocol, state, service_name, service_product, service_version, first_seen_scan_file_id, last_seen_scan_file_id, reviewed, reviewed_by, reviewed_at, review_notes, created_at, updated_at
		FROM ports
		ORDER BY created_at DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []*Port
	for rows.Next() {
		port := &Port{}
		err := rows.Scan(
			&port.ID,
			&port.HostID,
			&port.Port,
			&port.Protocol,
			&port.State,
			&port.ServiceName,
			&port.ServiceProduct,
			&port.ServiceVersion,
			&port.FirstSeenScanFileID,
			&port.LastSeenScanFileID,
			&port.Reviewed,
			&port.ReviewedBy,
			&port.ReviewedAt,
			&port.ReviewNotes,
			&port.CreatedAt,
			&port.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		ports = append(ports, port)
	}

	return ports, rows.Err()
}

func (db *DB) SetPortReviewed(id uuid.UUID, reviewed bool, reviewedBy uuid.UUID) error {
	if reviewed {
		_, err := db.conn.Exec(`UPDATE ports SET reviewed = true, reviewed_by = $1, reviewed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, reviewedBy, id)
		return err
	}
	_, err := db.conn.Exec(`UPDATE ports SET reviewed = false, reviewed_by = NULL, reviewed_at = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, id)
	return err
}

func (db *DB) GetPortByID(id uuid.UUID) (*Port, error) {
	port := &Port{}
	query := `
		SELECT id, host_id, port, protocol, state, service_name, service_product, service_version, first_seen_scan_file_id, last_seen_scan_file_id, reviewed, reviewed_by, reviewed_at, review_notes, created_at, updated_at
		FROM ports
		WHERE id = $1
	`

	err := db.conn.QueryRow(query, id).Scan(
		&port.ID,
		&port.HostID,
		&port.Port,
		&port.Protocol,
		&port.State,
		&port.ServiceName,
		&port.ServiceProduct,
		&port.ServiceVersion,
		&port.FirstSeenScanFileID,
		&port.LastSeenScanFileID,
		&port.Reviewed,
		&port.ReviewedBy,
		&port.ReviewedAt,
		&port.ReviewNotes,
		&port.CreatedAt,
		&port.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return port, nil
}

// PortAggregate is one row per (port, protocol) with reviewed = true only when all host-ports are reviewed.
type PortAggregate struct {
	Port          int
	Protocol      string
	State         sql.NullString
	ServiceName   sql.NullString
	ServiceProduct sql.NullString
	ServiceVersion sql.NullString
	Reviewed      bool // true only when BOOL_AND(reviewed) for this port+protocol
	HostCount     int
}

// ListPortsAggregated optionally filters by search (port number or service name substring) and reviewed.
func (db *DB) ListPortsAggregated(search string, reviewed *bool) ([]*PortAggregate, error) {
	query := `
		SELECT port, protocol,
			MAX(state) AS state, MAX(service_name) AS service_name, MAX(service_product) AS service_product, MAX(service_version) AS service_version,
			BOOL_AND(reviewed) AS reviewed,
			COUNT(*) AS host_count
		FROM ports
	`
	var args []interface{}
	n := 1
	if search != "" {
		query += " WHERE (port::text LIKE '%' || $" + strconv.Itoa(n) + " || '%' OR service_name ILIKE '%' || $" + strconv.Itoa(n) + " || '%')"
		args = append(args, search)
		n++
	}
	query += " GROUP BY port, protocol"
	if reviewed != nil {
		query += " HAVING BOOL_AND(reviewed) = $" + strconv.Itoa(n)
		args = append(args, *reviewed)
	}
	query += " ORDER BY port, protocol"
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*PortAggregate
	for rows.Next() {
		a := &PortAggregate{}
		err := rows.Scan(
			&a.Port,
			&a.Protocol,
			&a.State,
			&a.ServiceName,
			&a.ServiceProduct,
			&a.ServiceVersion,
			&a.Reviewed,
			&a.HostCount,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// PortHostRow is one host's port row for a given port+protocol (for detail by number).
type PortHostRow struct {
	Host         *Host
	PortID       uuid.UUID
	PortReviewed bool
}

func (db *DB) GetPortDetailByNumber(portNum int, protocol string) (portInfo *Port, rows []*PortHostRow, err error) {
	query := `
		SELECT p.id, p.host_id, p.port, p.protocol, p.state, p.service_name, p.service_product, p.service_version, p.reviewed,
			h.id, h.project_id, h.ip_address, h.hostname, h.os, h.first_seen_scan_file_id, h.last_seen_scan_file_id, h.reviewed, h.reviewed_by, h.reviewed_at, h.review_notes, h.created_at, h.updated_at
		FROM ports p
		INNER JOIN hosts h ON h.id = p.host_id
		WHERE p.port = $1 AND p.protocol = $2
		ORDER BY h.created_at DESC
	`
	portRows, qErr := db.conn.Query(query, portNum, protocol)
	if qErr != nil {
		return nil, nil, qErr
	}
	defer portRows.Close()

	for portRows.Next() {
		var p Port
		var h Host
		err := portRows.Scan(
			&p.ID, &p.HostID, &p.Port, &p.Protocol, &p.State, &p.ServiceName, &p.ServiceProduct, &p.ServiceVersion, &p.Reviewed,
			&h.ID, &h.ProjectID, &h.IPAddress, &h.Hostname, &h.OS, &h.FirstSeenScanFileID, &h.LastSeenScanFileID, &h.Reviewed, &h.ReviewedBy, &h.ReviewedAt, &h.ReviewNotes, &h.CreatedAt, &h.UpdatedAt,
		)
		if err != nil {
			return nil, nil, err
		}
		if portInfo == nil {
			portInfo = &p
		}
		// Host reviewed is derived: all ports and all URLs reviewed (or none)
		if computed, err := db.ComputeHostReviewed(h.ID); err == nil {
			h.Reviewed = computed
		}
		rows = append(rows, &PortHostRow{Host: &h, PortID: p.ID, PortReviewed: p.Reviewed})
	}
	return portInfo, rows, portRows.Err()
}

func (db *DB) GetPortsByHostID(hostID uuid.UUID) ([]*Port, error) {
	query := `
		SELECT id, host_id, port, protocol, state, service_name, service_product, service_version, first_seen_scan_file_id, last_seen_scan_file_id, reviewed, reviewed_by, reviewed_at, review_notes, created_at, updated_at
		FROM ports
		WHERE host_id = $1
		ORDER BY port ASC
	`

	rows, err := db.conn.Query(query, hostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []*Port
	for rows.Next() {
		port := &Port{}
		err := rows.Scan(
			&port.ID,
			&port.HostID,
			&port.Port,
			&port.Protocol,
			&port.State,
			&port.ServiceName,
			&port.ServiceProduct,
			&port.ServiceVersion,
			&port.FirstSeenScanFileID,
			&port.LastSeenScanFileID,
			&port.Reviewed,
			&port.ReviewedBy,
			&port.ReviewedAt,
			&port.ReviewNotes,
			&port.CreatedAt,
			&port.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		ports = append(ports, port)
	}

	return ports, rows.Err()
}

func (db *DB) GetHostsByPort(portNum int, protocol string) ([]*Host, error) {
	query := `
		SELECT DISTINCT h.id, h.project_id, h.ip_address, h.hostname, h.os, h.first_seen_scan_file_id, h.last_seen_scan_file_id, h.reviewed, h.reviewed_by, h.reviewed_at, h.review_notes, h.created_at, h.updated_at
		FROM hosts h
		INNER JOIN ports p ON h.id = p.host_id
		WHERE p.port = $1 AND p.protocol = $2
		ORDER BY h.created_at DESC
	`

	rows, err := db.conn.Query(query, portNum, protocol)
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
