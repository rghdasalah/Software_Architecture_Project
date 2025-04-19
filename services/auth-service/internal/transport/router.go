package transport

import (
	"auth-service/internal/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(authService *service.AuthService) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	handler := &AuthHandler{Service: authService}

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Get("/google/callback", handler.GoogleCallbackHandler)
		r.Get("/refresh", handler.RefreshTokenHandler)

		//r.Get("/profile", handler.ProfileHandler)
		r.With(AuthMiddleware).Get("/profile", handler.ProfileHandler)
	})

	return r
}

//Register routes and middleware
