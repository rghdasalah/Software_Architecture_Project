// internal/api/middleware/auth.go
package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/service"
	"github.com/google/uuid"
)

// Define a type for the context key to avoid collisions
type ContextKey string
const UserIDKey ContextKey = "userID"

// AuthMiddleware authenticates requests using the user service
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the user service from the request context
		// This needs to be set in the main handler setup
		userService, ok := r.Context().Value("userService").(*service.UserService)
		if !ok {
			log.Println("User service not found in context")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check for Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			http.Error(w, "Empty token", http.StatusUnauthorized)
			return
		}

		// Validate token using user service
		userID, err := userService.ValidateToken(r.Context(), token)
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Store user ID in request context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext retrieves the authenticated user's ID from the request context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, errors.New("user ID not found in context")
	}
	return userID, nil
}

// For backwards compatibility with test script
func validateTokenPlaceholder(token string) (uuid.UUID, error) {
    if token == "dummy-auth-token-for-testing" {
        return uuid.Parse("test-user-id-123")
    }
    return uuid.UUID{}, errors.New("invalid token")
}