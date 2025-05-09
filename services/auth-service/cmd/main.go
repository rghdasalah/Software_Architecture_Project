package main

import (
	"auth-service/internal/config"
	"auth-service/internal/db"
	"auth-service/internal/service"
	"auth-service/internal/transport"
	"fmt"
	"log"
	"net/http"
)

func main() {
	config.LoadConfig()
	fmt.Println("✅ Config Loaded Successfully")
	fmt.Println("🟢 Running on port:", config.AppConfig.Port)

	db.ConnectToDB()

	authService := &service.AuthService{DB: db.DB}
	router := transport.SetupRoutes(authService)

	err := http.ListenAndServe(":"+config.AppConfig.Port, router)
	if err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
