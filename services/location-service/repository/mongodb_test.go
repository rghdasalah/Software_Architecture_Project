package repository

import (
    "context"
    "testing"
    "time"
    "location-service/models"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func TestLocationRepository(t *testing.T) {
    // Connect to MongoDB
    client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        t.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    defer client.Disconnect(context.Background())

    repo := NewLocationRepository(client)

    // Test Save
    update := models.LocationUpdate{
        RideID:    456,
        Latitude:  30.0150,
        Longitude: 31.0000,
        Timestamp: time.Now(),
        ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
    }
    err = repo.Save(context.Background(), update)
    if err != nil {
        t.Fatalf("Failed to save location update: %v", err)
    }

    // Test GetLatest
    retrieved, err := repo.GetLatest(context.Background(), 456)
    if err != nil {
        t.Fatalf("Failed to get latest location: %v", err)
    }
    if retrieved.RideID != 456 {
        t.Errorf("Expected ride ID 456, got %d", retrieved.RideID)
    }
    if retrieved.Latitude != 30.0150 {
        t.Errorf("Expected latitude 30.0150, got %f", retrieved.Latitude)
    }
}