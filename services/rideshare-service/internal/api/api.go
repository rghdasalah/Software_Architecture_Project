package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/api/handlers"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/api/middleware"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/db"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/service"
	"github.com/gorilla/mux"
)

// Server encapsulates the API server components
type Server struct {
    router *mux.Router
    db     *sql.DB
    config ServerConfig
}

// ServerConfig holds configuration for the API server
type ServerConfig struct {
    Port string
}

// NewServer creates a new API server instance
func NewServer(db *sql.DB, config ServerConfig) *Server {
    router := mux.NewRouter()
    
    server := &Server{
        router: router,
        db:     db,
        config: config,
    }
    
    // Register all routes
    server.registerRoutes()
    
    return server
}

// registerRoutes sets up all API endpoints
func (s *Server) registerRoutes() {
    log.Println("Registering API routes...")
    
    // Register the health endpoint
    s.RegisterHealthHandler()
    
    // Register user routes
    s.RegisterUserRoutes()
    
    // Register vehicle routes
    s.RegisterVehicleRoutes()
    
    // Register ride routes - commented out until implemented
    // s.RegisterRideRoutes()
}

// RegisterHealthHandler sets up the health check endpoint
func (s *Server) RegisterHealthHandler() {
    s.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }).Methods("GET")
}

// RegisterUserRoutes sets up all user-related endpoints
func (s *Server) RegisterUserRoutes() {
    // Create required services
    dbManager, _ := db.NewDBManagerFromDB(s.db)
    userService := service.NewUserService(dbManager)
    userHandler := handlers.NewUserHandler(userService)
    
    // Public routes - no auth required
    s.router.HandleFunc("/api/users/register", userHandler.RegisterUser).Methods("POST")
    s.router.HandleFunc("/api/users/login", userHandler.LoginUser).Methods("POST")
    s.router.HandleFunc("/api/users/{id}", userHandler.GetUser).Methods("GET")
    
    // Debug logging for routes
    log.Println("Registered user routes")
}

// RegisterVehicleRoutes sets up all vehicle-related endpoints
func (s *Server) RegisterVehicleRoutes() {
    // Create required services
    dbManager, _ := db.NewDBManagerFromDB(s.db)
    userService := service.NewUserService(dbManager)
    vehicleService := service.NewVehicleService(dbManager)
    vehicleHandler := handlers.NewVehicleHandler(vehicleService)
    
    // Create a subrouter for authenticated endpoints
    authRouter := s.router.PathPrefix("/api").Subrouter()
    
    // Middleware to inject userService into context
    authRouter.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := context.WithValue(r.Context(), "userService", userService)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    })
    
    // Apply authentication middleware
    authRouter.Use(middleware.AuthMiddleware)
    
    // Auth required endpoints - fixing the routes to be consistent
    authRouter.HandleFunc("/vehicles", vehicleHandler.CreateVehicle).Methods("POST")
    authRouter.HandleFunc("/vehicles", vehicleHandler.GetUserVehiclesForAuthUser).Methods("GET")
    authRouter.HandleFunc("/vehicles/{id}", vehicleHandler.DeleteVehicle).Methods("DELETE")
    
    // Public endpoints
    s.router.HandleFunc("/api/users/{id}/vehicles", vehicleHandler.GetUserVehicles).Methods("GET")
    
    // Log registered routes for debugging
    log.Println("Registered vehicle routes:")
    log.Println("- POST /api/vehicles")
    log.Println("- GET /api/vehicles")
    log.Println("- DELETE /api/vehicles/{id}")
    log.Println("- GET /api/users/{id}/vehicles")
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
    addr := fmt.Sprintf(":%s", s.config.Port)
    log.Printf("API server listening on %s", addr)
    
    return http.ListenAndServe(addr, s.router)
}