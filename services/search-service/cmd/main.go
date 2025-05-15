package main

import (
	"log"
	"net/http"
	"os"

	"search-service/internal/cache"
	"search-service/internal/config"
	"search-service/internal/db"
	"search-service/internal/grpc"
	"search-service/internal/service"
	"search-service/internal/transport"
)

func main() {
	// Load environment-based config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	// Connect to Redis
	redisClient := cache.NewRedisClient(cfg.RedisAddress)

	// Connect to geo-distance-service via gRPC
	geoClient := grpc.NewGeoClient(cfg)

	// Inject the repository logic
	rideRepo := &db.MockRideRepo{}
	// Inject into search logic
	searchLogic := service.NewSearchService(geoClient, redisClient)

	// HTTP handler setup
	handler := transport.NewHandler(searchLogic, rideRepo)
	router := transport.SetupRouter(handler)

	// Port from .env or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Search service listening on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
