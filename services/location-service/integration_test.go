package main

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    "location-service/api"
    "location-service/models"
    "location-service/queue"
    "location-service/repository"
    "location-service/service"

    "github.com/stretchr/testify/assert"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func TestIntegration(t *testing.T) {
    // Connect to MongoDB
    client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
    assert.NoError(t, err)
    defer client.Disconnect(context.Background())

    // Clear the database before the test
    client.Database("rideshare").Collection("location_updates").Drop(context.Background())

    // Initialize repository
    repo := repository.NewLocationRepository(client)

    // Connect to RabbitMQ
    q, err := queue.NewLocationQueue("amqp://guest:guest@localhost:5672/")
    assert.NoError(t, err)

    // Initialize service
    svc := service.NewLocationService(repo, q)

    // Start the service in a separate goroutine
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go func() {
        if err := svc.Start(ctx); err != nil {
            t.Logf("Service error: %v", err)
        }
    }()

    // Initialize API handler with service and queue
    handler := api.NewHandler(svc, q)

    // Test: Publish a location update
    update := models.LocationUpdate{
        RideID:    123,
        Latitude:  40.7128,
        Longitude: -74.0060,
        Timestamp: time.Now(),
        ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
    }
    body, _ := json.Marshal(update)
    req := httptest.NewRequest("POST", "/location/", bytes.NewReader(body))
    w := httptest.NewRecorder()
    handler.PostLocation(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)
    t.Log("Location update published")

    // Wait longer for the message to be processed
    time.Sleep(2 * time.Second)

    // Verify the data was saved to MongoDB directly
    collection := client.Database("rideshare").Collection("location_updates")
    var savedUpdate models.LocationUpdate
    err = collection.FindOne(context.Background(), map[string]int{"ride_id": 123}).Decode(&savedUpdate)
    assert.NoError(t, err)
    assert.Equal(t, update.RideID, savedUpdate.RideID)
    assert.Equal(t, update.Latitude, savedUpdate.Latitude)
    assert.Equal(t, update.Longitude, savedUpdate.Longitude)

    // Test: Retrieve the location
    req = httptest.NewRequest("GET", "/location/123", nil)
    w = httptest.NewRecorder()
    handler.GetLatestLocation(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    var result models.LocationUpdate
    err = json.NewDecoder(w.Body).Decode(&result)
    assert.NoError(t, err)
    assert.Equal(t, update.RideID, result.RideID)
    assert.Equal(t, update.Latitude, result.Latitude)
    assert.Equal(t, update.Longitude, result.Longitude)
}