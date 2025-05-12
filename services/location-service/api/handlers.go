package api

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "time"
    "location-service/models"
    "location-service/queue"
    "location-service/service"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
    svc   service.LocationServiceInterface
    queue queue.LocationQueue
}

// NewHandler creates a new Handler instance
func NewHandler(svc service.LocationServiceInterface, queue queue.LocationQueue) *Handler {
    return &Handler{svc: svc, queue: queue}
}

// GetLatestLocation handles GET /location/{rideID} requests
func (h *Handler) GetLatestLocation(w http.ResponseWriter, r *http.Request) {
    // Extract rideID from URL path (e.g., /location/123 -> rideID = 123)
    rideIDStr := r.URL.Path[len("/location/"):]
    rideID, err := strconv.Atoi(rideIDStr)
    if err != nil {
        http.Error(w, "Invalid rideID", http.StatusBadRequest)
        return
    }

    // Call service to get the latest location
    update, err := h.svc.GetLatestLocation(rideID)
    if err != nil {
        log.Printf("Failed to get latest location for ride %d: %v", rideID, err)
        http.Error(w, "Failed to get location", http.StatusInternalServerError)
        return
    }
    if update == nil {
        http.Error(w, "Location not found", http.StatusNotFound)
        return
    }

    // Respond with JSON
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(update); err != nil {
        log.Printf("Failed to encode response: %v", err)
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
    }
}

// PostLocation handles POST /location requests
func (h *Handler) PostLocation(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var update models.LocationUpdate
    if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate required fields
    if update.RideID == 0 || update.Latitude == 0 || update.Longitude == 0 {
        http.Error(w, "Missing required fields", http.StatusBadRequest)
        return
    }

    // Set timestamp and expiration if not provided
    if update.Timestamp.IsZero() {
        update.Timestamp = time.Now()
    }
    if update.ExpiresAt.IsZero() {
        update.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
    }

    // Log after setting defaults
    log.Printf("Received and processed update: %+v", update)

    // Publish to queue
    if err := h.queue.Publish(update); err != nil {
        log.Printf("Failed to publish location update: %v", err)
        http.Error(w, "Failed to publish location update", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Location update published"})
}