package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"

	"github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type DB struct {
	conn *sql.DB
}

func New(conn *sql.DB) *DB {
	return &DB{conn: conn}
}

func RunMigrations(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	migrations, err := fs.Glob(migrationsFS, "migrations/*.up.sql")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// Derive numeric versions from filenames and sort ascending so migrations
	// are applied in a deterministic order.
	type migrationFile struct {
		Path    string
		Version int
	}

	var mf []migrationFile
	for _, path := range migrations {
		var v int
		// Expect filenames like "001_initial_schema.up.sql"
		if _, err := fmt.Sscanf(path, "migrations/%d", &v); err != nil {
			return fmt.Errorf("failed to parse migration version from %q: %w", path, err)
		}
		mf = append(mf, migrationFile{Path: path, Version: v})
	}

	sort.Slice(mf, func(i, j int) bool {
		return mf[i].Version < mf[j].Version
	})

	// Simple migration runner - in production, use a proper migration tool
	for _, m := range mf {
		var existing int
		err := db.QueryRow(
			"SELECT version FROM schema_migrations WHERE version = $1",
			m.Version,
		).Scan(&existing)

		switch {
		case err == sql.ErrNoRows:
			// Not applied yet
		case err != nil:
			return fmt.Errorf("failed to check migration status: %w", err)
		default:
			// Already applied, skip
			continue
		}

		content, err := fs.ReadFile(migrationsFS, m.Path)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", m.Path, err)
		}

		// Execute migration in a transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", m.Path, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version) VALUES ($1)",
			m.Version,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration: %w", err)
		}
	}

	return nil
}

// Helper function to check if error is a unique constraint violation
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // unique_violation
	}
	return false
}
