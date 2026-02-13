package parsers

import (
	"log"

	"github.com/google/uuid"

	"github.com/edda/backend/internal/database"
)

// getOrCreateHostForURL returns the host ID for the given hostname (from a URL).
// Tries ip_address first, then hostname, then creates a new host so the host count
// updates and URLs link to it. Used by ffuf, postman, and openapi parsers.
func getOrCreateHostForURL(db *database.DB, hostname string, scanFileID uuid.UUID) *uuid.UUID {
	host, err := db.GetHostByIP(hostname)
	if err == nil && host != nil {
		return &host.ID
	}
	host, err = db.GetHostByHostname(hostname)
	if err == nil && host != nil {
		return &host.ID
	}
	host, err = db.UpsertHost(hostname, hostname, "", scanFileID)
	if err != nil {
		log.Printf("Failed to create host for %s: %v", hostname, err)
		return nil
	}
	return &host.ID
}
