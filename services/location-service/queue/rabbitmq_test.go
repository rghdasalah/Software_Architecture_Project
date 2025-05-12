package queue

import (
    "encoding/json" // Used directly for dummy operation and indirectly via Publish/Consume
    "testing"
    "time"
    "location-service/models"

    "github.com/streadway/amqp" // Used directly for dummy operation and indirectly via LocationQueue
)

func TestLocationQueue(t *testing.T) {
    // Dummy usage to avoid "imported and not used" error
    var _ = json.Marshal
    var _ = amqp.Queue{}

    // Connect to RabbitMQ
    q, err := NewLocationQueue("amqp://guest:guest@localhost:5672/")
    if err != nil {
        t.Fatalf("Failed to create LocationQueue: %v", err)
    }

    // Test Publish
    update := models.LocationUpdate{
        RideID:    456,
        Latitude:  30.0150,
        Longitude: 31.0000,
        Timestamp: time.Now(),
        ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
    }
    err = q.Publish(update)
    if err != nil {
        t.Fatalf("Failed to publish location update: %v", err)
    }

    // Test Consume
    received := make(chan models.LocationUpdate)
    err = q.Consume(func(update models.LocationUpdate) {
        received <- update
    })
    if err != nil {
        t.Fatalf("Failed to consume messages: %v", err)
    }

    // Wait for the message to be consumed
    select {
    case retrieved := <-received:
        if retrieved.RideID != 456 {
            t.Errorf("Expected ride ID 456, got %d", retrieved.RideID)
        }
        if retrieved.Latitude != 30.0150 {
            t.Errorf("Expected latitude 30.0150, got %f", retrieved.Latitude)
        }
    case <-time.After(5 * time.Second):
        t.Fatal("Timed out waiting for message")
    }
}