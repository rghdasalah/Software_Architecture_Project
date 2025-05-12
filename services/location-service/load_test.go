package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "testing"
    "time"
    "location-service/models"

    "github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
    // Target: 200 updates per minute = 200 / 60 = ~3.33 updates per second
    updatesPerSecond := 3.33
    duration := 60 * time.Second // Run for 1 minute
    totalUpdates := int(updatesPerSecond * duration.Seconds()) // 200 updates

    var wg sync.WaitGroup
    wg.Add(totalUpdates)

    // Rate limiter to control the frequency of requests
    ticker := time.NewTicker(time.Duration(float64(time.Second) / updatesPerSecond))
    defer ticker.Stop()

    successCount := 0
    mu := sync.Mutex{}

    start := time.Now()
    for i := 0; i < totalUpdates; i++ {
        <-ticker.C
        go func(rideID int) {
            defer wg.Done()

            update := models.LocationUpdate{
                RideID:    rideID,
                Latitude:  40.7128 + float64(rideID)/1000, // Slight variation in coordinates
                Longitude: -74.0060 + float64(rideID)/1000,
                Timestamp: time.Now(),                     // Set current time
                ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // Set expiration
            }
            body, _ := json.Marshal(update)
            resp, err := http.Post("http://localhost:8080/location/", "application/json", bytes.NewReader(body))
            if err != nil {
                t.Logf("Failed to publish update for ride %d: %v", rideID, err)
                return
            }
            defer resp.Body.Close()

            if resp.StatusCode == http.StatusCreated {
                mu.Lock()
                successCount++
                mu.Unlock()
            } else {
                t.Logf("Failed to publish update for ride %d: status %d", rideID, resp.StatusCode)
            }
        }(i + 1) // Use unique rideIDs
    }

    wg.Wait()
    elapsed := time.Since(start)

    t.Logf("Published %d/%d updates successfully in %v", successCount, totalUpdates, elapsed)
    assert.GreaterOrEqual(t, successCount, totalUpdates*9/10, "At least 90% of updates should succeed")
}

func TestLoadRetrieve(t *testing.T) {
    // Wait for the service to process all messages (adjust based on system performance)
    time.Sleep(10 * time.Second)

    // Retrieve a few updates to verify they were saved
    for rideID := 1; rideID <= 5; rideID++ {
        resp, err := http.Get(fmt.Sprintf("http://localhost:8080/location/%d", rideID))
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode, "Should retrieve location for ride %d", rideID)
        resp.Body.Close()
    }
}