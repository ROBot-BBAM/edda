package handlers

import "github.com/edda/backend/internal/database"

type Handlers struct {
	db *database.DB
}

// SetReviewedRequest is the body for PATCH .../reviewed endpoints.
type SetReviewedRequest struct {
	Reviewed bool `json:"reviewed"`
}

func New(db *database.DB) *Handlers {
	return &Handlers{db: db}
}
