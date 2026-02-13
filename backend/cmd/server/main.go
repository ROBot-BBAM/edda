package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"

	"github.com/edda/backend/internal/database"
	"github.com/edda/backend/internal/handlers"
	authMiddleware "github.com/edda/backend/internal/middleware/auth"
)

func main() {
	var port = flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://edda:edda_dev_password@localhost:5432/edda?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize database layer
	dbLayer := database.New(db)

	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}

	// Initialize handlers
	h := handlers.New(dbLayer)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Public routes (registration disabled, but endpoint exists for clarity)
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Middleware(jwtSecret))
			r.Get("/me", h.GetMe)

			// Admin routes
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.AdminMiddleware(dbLayer))
				r.Post("/admin/users", h.CreateUser)
				r.Get("/admin/users", h.ListUsers)
				r.Delete("/admin/users/{id}", h.DeleteUser)
			})

			// Scan files routes
			r.Route("/scan-files", func(r chi.Router) {
				r.Get("/", h.ListScanFiles)
				r.Post("/", h.UploadScanFile)
			})

			// Data routes (hosts, ports, URLs extracted from scan files)
			r.Get("/hosts", h.ListHosts)
			r.Get("/hosts/{id}", h.GetHost)
			r.Get("/hosts/{id}/notes", h.ListNotesByHost)
			r.Patch("/hosts/{id}/reviewed", h.SetHostReviewed)
			r.Post("/notes", h.CreateNote)
			r.Get("/ports", h.ListPorts)
			r.Get("/ports/by-number/{port}/{protocol}", h.GetPortByNumber)
			r.Get("/ports/{id}/notes", h.ListNotesByPort)
			r.Patch("/ports/{id}/reviewed", h.SetPortReviewed)
			r.Get("/urls", h.ListURLs)
			r.Get("/urls/{id}/notes", h.ListNotesByURL)
			r.Patch("/urls/{id}/reviewed", h.SetURLReviewed)

			// Search (global)
			r.Get("/search", h.Search)

			// Export (full engagement for AI/narrative)
			r.Get("/export/narrative", h.ExportNarrative)

			// Findings (vulnerabilities)
			r.Get("/findings", h.ListFindings)
			r.Get("/findings/summary", h.GetFindingsSummary)
			r.Post("/findings", h.CreateFinding)
			r.Get("/findings/{id}", h.GetFinding)
			r.Patch("/findings/{id}", h.UpdateFinding)
			r.Delete("/findings/{id}", h.DeleteFinding)

			// Legacy project routes (kept for compatibility, but will be simplified)
			r.Route("/projects", func(r chi.Router) {
				r.Get("/", h.ListProjects)
				r.Post("/", h.CreateProject)
				r.Get("/{id}", h.GetProject)
				r.Patch("/{id}", h.UpdateProject)
				r.Delete("/{id}", h.DeleteProject)
			})
		})
	})

	// Start server
	srv := &http.Server{
		Addr:    ":" + *port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on port %s", *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
