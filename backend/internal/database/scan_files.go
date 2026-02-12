package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ScanFile struct {
	ID           uuid.UUID
	ProjectID    sql.NullString // Nullable for single engagement model
	Type         string         // 'nmap_xml', 'ffuf_json', 'ffuf_csv'
	Filename     string
	UploadedBy   uuid.UUID
	UploadedAt   time.Time
	Status       string // 'pending', 'parsing', 'parsed', 'failed'
	ErrorMessage sql.NullString
	StoragePath  sql.NullString
}

func (db *DB) CreateScanFile(filename, fileType, storagePath string, uploadedBy uuid.UUID) (*ScanFile, error) {
	scanFile := &ScanFile{
		ID:         uuid.New(),
		Type:       fileType,
		Filename:   filename,
		UploadedBy: uploadedBy,
		Status:     "pending",
		UploadedAt: time.Now(),
	}

	if storagePath != "" {
		scanFile.StoragePath = sql.NullString{String: storagePath, Valid: true}
	}

	query := `
		INSERT INTO scan_files (id, project_id, type, filename, uploaded_by, uploaded_at, status, storage_path)
		VALUES ($1, NULL, $2, $3, $4, $5, $6, $7)
		RETURNING id, project_id, type, filename, uploaded_by, uploaded_at, status, error_message, storage_path
	`

	err := db.conn.QueryRow(
		query,
		scanFile.ID,
		scanFile.Type,
		scanFile.Filename,
		scanFile.UploadedBy,
		scanFile.UploadedAt,
		scanFile.Status,
		scanFile.StoragePath,
	).Scan(
		&scanFile.ID,
		&scanFile.ProjectID,
		&scanFile.Type,
		&scanFile.Filename,
		&scanFile.UploadedBy,
		&scanFile.UploadedAt,
		&scanFile.Status,
		&scanFile.ErrorMessage,
		&scanFile.StoragePath,
	)

	if err != nil {
		return nil, err
	}

	return scanFile, nil
}

func (db *DB) ListScanFiles() ([]*ScanFile, error) {
	query := `
		SELECT id, project_id, type, filename, uploaded_by, uploaded_at, status, error_message, storage_path
		FROM scan_files
		ORDER BY uploaded_at DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scanFiles []*ScanFile
	for rows.Next() {
		sf := &ScanFile{}
		err := rows.Scan(
			&sf.ID,
			&sf.ProjectID,
			&sf.Type,
			&sf.Filename,
			&sf.UploadedBy,
			&sf.UploadedAt,
			&sf.Status,
			&sf.ErrorMessage,
			&sf.StoragePath,
		)
		if err != nil {
			return nil, err
		}
		scanFiles = append(scanFiles, sf)
	}

	return scanFiles, rows.Err()
}

func (db *DB) GetScanFileByID(id uuid.UUID) (*ScanFile, error) {
	sf := &ScanFile{}
	query := `
		SELECT id, project_id, type, filename, uploaded_by, uploaded_at, status, error_message, storage_path
		FROM scan_files
		WHERE id = $1
	`

	err := db.conn.QueryRow(query, id).Scan(
		&sf.ID,
		&sf.ProjectID,
		&sf.Type,
		&sf.Filename,
		&sf.UploadedBy,
		&sf.UploadedAt,
		&sf.Status,
		&sf.ErrorMessage,
		&sf.StoragePath,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return sf, nil
}

func (db *DB) UpdateScanFileStatus(id uuid.UUID, status string, errorMessage string) error {
	var errMsg sql.NullString
	if errorMessage != "" {
		errMsg = sql.NullString{String: errorMessage, Valid: true}
	}

	query := `
		UPDATE scan_files
		SET status = $1, error_message = $2
		WHERE id = $3
	`

	_, err := db.conn.Exec(query, status, errMsg, id)
	return err
}
