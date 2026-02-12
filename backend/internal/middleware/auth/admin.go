package auth

import (
	"net/http"

	"github.com/edda/backend/internal/database"
	"github.com/google/uuid"
)

// AdminMiddleware checks if the authenticated user is an admin
func AdminMiddleware(db *database.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDStr := GetUserID(r)
			if userIDStr == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusBadRequest)
				return
			}

			user, err := db.GetUserByID(userID)
			if err != nil || user == nil {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}

			if !user.IsAdmin {
				http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
