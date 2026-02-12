package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/edda/backend/internal/database"
)

func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := strings.Builder{}
	password.Grow(length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password.WriteByte(charset[num.Int64()])
	}

	return password.String(), nil
}

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://edda:edda_dev_password@db:5432/edda?sslmode=disable"
	}

	// Get admin email from environment or use default
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@edda.local"
	}

	// Connect to database with retries
	var db *sql.DB
	var err error
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(2 * time.Second)
				continue
			}
			log.Fatalf("Failed to connect to database after %d attempts", maxRetries)
		}

		if err = db.Ping(); err == nil {
			break
		}
		log.Printf("Database not ready (attempt %d/%d), retrying...", i+1, maxRetries)
		if i < maxRetries-1 {
			db.Close()
			time.Sleep(2 * time.Second)
		}
	}
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	defer db.Close()

	// Run migrations first to ensure tables exist
	log.Println("Running database migrations...")
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	// Check if admin user already exists
	var existingID string
	err = db.QueryRow("SELECT id FROM users WHERE email = $1", adminEmail).Scan(&existingID)
	if err == nil {
		log.Printf("Admin user '%s' already exists, skipping creation", adminEmail)
		return
	}
	if err != sql.ErrNoRows {
		log.Fatalf("Failed to check for existing admin user: %v", err)
	}

	// Generate random password
	password, err := generateRandomPassword(16)
	if err != nil {
		log.Fatalf("Failed to generate random password: %v", err)
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create admin user
	_, err = db.Exec(`
		INSERT INTO users (id, email, password_hash, is_admin, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, adminEmail, string(hash))

	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	// Output credentials
	fmt.Println("=" + strings.Repeat("=", 78) + "=")
	fmt.Println("=" + strings.Repeat(" ", 78) + "=")
	fmt.Println("=" + strings.Repeat(" ", 20) + "EDDA ADMIN CREDENTIALS" + strings.Repeat(" ", 35) + "=")
	fmt.Println("=" + strings.Repeat(" ", 78) + "=")
	fmt.Println("=" + strings.Repeat("=", 78) + "=")
	fmt.Printf("Email:    %s\n", adminEmail)
	fmt.Printf("Password: %s\n", password)
	fmt.Println("=" + strings.Repeat("=", 78) + "=")
	fmt.Println("⚠️  IMPORTANT: Save these credentials! They will not be shown again.")
	fmt.Println("=" + strings.Repeat("=", 78) + "=")

	// Also write to a file for easy access
	credsFile := "/tmp/edda-admin-credentials.txt"
	file, err := os.Create(credsFile)
	if err == nil {
		fmt.Fprintf(file, "Edda Admin Credentials\n")
		fmt.Fprintf(file, "======================\n\n")
		fmt.Fprintf(file, "Email:    %s\n", adminEmail)
		fmt.Fprintf(file, "Password: %s\n", password)
		fmt.Fprintf(file, "\n⚠️  IMPORTANT: Save these credentials! They will not be shown again.\n")
		file.Close()
		fmt.Printf("\nCredentials also saved to: %s\n", credsFile)
	}
}
