package transport

import (
	"auth-service/internal/token"
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userCtxKey = contextKey("user")

// AuthMiddleware checks for valid access token and injects user info into context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := token.VerifyToken(tokenStr)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userCtxKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext retrieves user claims from request context
func GetUserFromContext(r *http.Request) *token.Claims {
	claims, ok := r.Context().Value(userCtxKey).(*token.Claims)
	if !ok {
		return nil
	}
	return claims
}
