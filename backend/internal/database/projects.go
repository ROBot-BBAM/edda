package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID
	Name        string
	Description sql.NullString
	OwnerID     uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (db *DB) CreateProject(name, description string, ownerID uuid.UUID) (*Project, error) {
	project := &Project{
		ID:        uuid.New(),
		Name:      name,
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if description != "" {
		project.Description = sql.NullString{String: description, Valid: true}
	}

	query := `
		INSERT INTO projects (id, name, description, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, description, owner_id, created_at, updated_at
	`

	err := db.conn.QueryRow(
		query,
		project.ID,
		project.Name,
		project.Description,
		project.OwnerID,
		project.CreatedAt,
		project.UpdatedAt,
	).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.OwnerID,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return project, nil
}

func (db *DB) GetProjectByID(id uuid.UUID) (*Project, error) {
	project := &Project{}
	query := `
		SELECT id, name, description, owner_id, created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	err := db.conn.QueryRow(query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.OwnerID,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (db *DB) ListProjectsByUserID(userID uuid.UUID) ([]*Project, error) {
	// All users see all projects (shared engagement model)
	query := `
		SELECT p.id, p.name, p.description, p.owner_id, p.created_at, p.updated_at
		FROM projects p
		ORDER BY p.created_at DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.OwnerID,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

func (db *DB) UpdateProject(id uuid.UUID, name string, description sql.NullString) (*Project, error) {
	query := `
		UPDATE projects
		SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING id, name, description, owner_id, created_at, updated_at
	`

	project := &Project{}
	err := db.conn.QueryRow(query, name, description, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.OwnerID,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (db *DB) DeleteProject(id uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`
	result, err := db.conn.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
