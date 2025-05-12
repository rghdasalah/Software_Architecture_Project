package main

import (
    "context"
    "log"
    "net/http"
    "location-service/api"
    "location-service/queue"
    "location-service/repository"
    "location-service/service"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    // Connect to MongoDB
    client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    defer client.Disconnect(context.Background())

    // Initialize repository
    repo := repository.NewLocationRepository(client)

    // Connect to RabbitMQ
    q, err := queue.NewLocationQueue("amqp://guest:guest@localhost:5672/")
    if err != nil {
        log.Fatalf("Failed to connect to RabbitMQ: %v", err)
    }

    // Initialize service
    svc := service.NewLocationService(repo, q)

    // Start consuming messages in a separate goroutine
    go func() {
        if err := svc.Start(context.Background()); err != nil {
            log.Fatalf("Failed to start service: %v", err)
        }
    }()

    // Initialize API handler with both service and queue
    handler := api.NewHandler(svc, q)

    // Set up HTTP routes
    http.HandleFunc("/location/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost && r.URL.Path == "/location/" {
            handler.PostLocation(w, r)
        } else {
            handler.GetLatestLocation(w, r)
        }
    })

    // Start HTTP server
    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}