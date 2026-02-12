package parsers

import (
	"fmt"
	"log"

	"github.com/edda/backend/internal/database"
)

func ParseScanFile(scanFile *database.ScanFile, db *database.DB) error {
	log.Printf("Starting to parse scan file: %s (type: %s)", scanFile.Filename, scanFile.Type)

	// Check if storage path is valid
	if !scanFile.StoragePath.Valid || scanFile.StoragePath.String == "" {
		err := fmt.Errorf("scan file has no storage path")
		log.Printf("Failed to parse scan file %s: %v", scanFile.Filename, err)
		if updateErr := db.UpdateScanFileStatus(scanFile.ID, "failed", err.Error()); updateErr != nil {
			log.Printf("Failed to update scan file status to failed: %v", updateErr)
		}
		return err
	}

	// Update status to parsing
	if err := db.UpdateScanFileStatus(scanFile.ID, "parsing", ""); err != nil {
		log.Printf("Failed to update scan file status to parsing: %v", err)
	}

	var err error
	switch scanFile.Type {
	case "nmap_xml":
		err = ParseNmapXML(scanFile.StoragePath.String, scanFile.ID, db)
	case "ffuf_json":
		err = ParseFFufJSON(scanFile.StoragePath.String, scanFile.ID, db)
	case "ffuf_csv":
		err = ParseFFufCSV(scanFile.StoragePath.String, scanFile.ID, db)
	default:
		err = fmt.Errorf("unknown file type: %s", scanFile.Type)
	}

	if err != nil {
		log.Printf("Failed to parse scan file %s: %v", scanFile.Filename, err)
		if updateErr := db.UpdateScanFileStatus(scanFile.ID, "failed", err.Error()); updateErr != nil {
			log.Printf("Failed to update scan file status to failed: %v", updateErr)
		}
		return err
	}

	// Update status to parsed
	if err := db.UpdateScanFileStatus(scanFile.ID, "parsed", ""); err != nil {
		log.Printf("Failed to update scan file status to parsed: %v", err)
		return err
	}

	log.Printf("Successfully parsed scan file: %s", scanFile.Filename)
	return nil
}
