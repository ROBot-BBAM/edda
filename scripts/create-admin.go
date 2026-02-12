package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run scripts/create-admin.go <email> <password>")
		fmt.Println("Example: go run scripts/create-admin.go admin@example.com admin123")
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]

	// Get database URL from environment or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://edda:edda_dev_password@localhost:5432/edda?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Generate bcrypt hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Insert admin user
	_, err = db.Exec(`
		INSERT INTO users (id, email, password_hash, is_admin, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, email, string(hash))

	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("✓ Admin user created successfully!\n")
	fmt.Printf("  Email: %s\n", email)
	fmt.Printf("  Password: %s\n", password)
	fmt.Printf("\nYou can now log in at http://localhost:3000/login\n")
}
