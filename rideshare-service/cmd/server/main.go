// cmd/server/main.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/api/handlers"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/api/middleware"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/config"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/db"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/service"
	"github.com/gorilla/mux"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig(".")
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Initialize database connection
    dbManager, err := db.NewDBManager(&cfg.Database)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer dbManager.Close()

    // Initialize services
    rideService := service.NewRideService(dbManager)
    userService := service.NewUserService(dbManager)
    vehicleService := service.NewVehicleService(dbManager)

    // Initialize handlers
    rideHandler := handlers.NewRideHandler(rideService, vehicleService)
    userHandler := handlers.NewUserHandler(userService)
    vehicleHandler := handlers.NewVehicleHandler(vehicleService)
    
    // Set up router
    r := mux.NewRouter()
    
    // Add health check endpoint
    r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        // Check database connection
        dbStatus := "ok"
        
        // Try executing a simple query to test connectivity
        primaryDB := dbManager.GetPrimary()
        if err := primaryDB.Ping(); err != nil {
            dbStatus = "error: " + err.Error()
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status": "ok",
            "database": dbStatus,
            "version": "1.0.0",
        })
    }).Methods("GET")    
    
    // Public routes (no authentication required)
    public := r.PathPrefix("/api").Subrouter()

    // IMPORTANT: Register most specific routes first!
    public.HandleFunc("/rides/nearby", rideHandler.FindNearbyRides).Methods("GET")
    public.HandleFunc("/rides/{id}", rideHandler.GetRide).Methods("GET")
    public.HandleFunc("/users/{id}/vehicles", vehicleHandler.GetUserVehicles).Methods("GET")
    public.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
    public.HandleFunc("/users/register", userHandler.RegisterUser).Methods("POST")
    public.HandleFunc("/users/login", userHandler.LoginUser).Methods("POST")

    // Middleware to inject user service into request context
    serviceMiddleware := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Add user service to request context
            ctx := context.WithValue(r.Context(), "userService", userService)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }

    // Protected routes with authentication
    protected := r.PathPrefix("/api").Subrouter()
    protected.Use(serviceMiddleware)  // Add service middleware first
    protected.Use(middleware.AuthMiddleware)  // Then auth middleware
    protected.HandleFunc("/vehicles", vehicleHandler.CreateVehicle).Methods("POST")
    protected.HandleFunc("/vehicles", vehicleHandler.GetUserVehiclesForAuthUser).Methods("GET")
    protected.HandleFunc("/rides", rideHandler.CreateRide).Methods("POST")
    protected.HandleFunc("/rides/{id}", rideHandler.CancelRide).Methods("DELETE")

    // Create HTTP server
    srv := &http.Server{
        Addr:         ":" + cfg.Server.Port,
        Handler:      r,
        ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
        WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
    }

    // Start server in a goroutine
    go func() {
        log.Printf("Server starting on port %s", cfg.Server.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    // Wait for interrupt signal to gracefully shut down the server
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c

    // Create a deadline to wait for
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    // Doesn't block if no connections, but will otherwise wait until the timeout
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server shutdown failed: %v", err)
    }
    log.Println("Server gracefully stopped")
}