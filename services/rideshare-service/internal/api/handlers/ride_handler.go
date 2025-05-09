package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/api/middleware"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/models"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type RideHandler struct {
    rideService    *service.RideService
    vehicleService *service.VehicleService
}

func NewRideHandler(rideService *service.RideService, vehicleService *service.VehicleService) *RideHandler {
    return &RideHandler{
        rideService: rideService,
        vehicleService: vehicleService,
    }
}

// CreateRide handles ride creation
func (h *RideHandler) CreateRide(w http.ResponseWriter, r *http.Request) {
    // Get user ID from context (set by auth middleware)
    userID, err := middleware.GetUserIDFromContext(r.Context())
    if err != nil {
        log.Printf("Error getting user ID from context: %v", err)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Parse request body
    var req models.CreateRideRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("Error decoding request body: %v", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate vehicle ownership
    vehicle, err := h.vehicleService.GetVehicleByID(r.Context(), req.VehicleID)
    if err != nil {
        log.Printf("Error fetching vehicle: %v", err)
        http.Error(w, "Failed to validate vehicle", http.StatusBadRequest)
        return
    }
    
    if vehicle.UserID != userID {
        log.Printf("Vehicle %s does not belong to user %s", req.VehicleID, userID)
        http.Error(w, "Vehicle does not belong to authenticated user", http.StatusForbidden)
        return
    }

    // Create ride using service
    ride, err := h.rideService.CreateRide(r.Context(), userID, &req)
    if err != nil {
        log.Printf("Error creating ride: %v", err)
        http.Error(w, "Failed to create ride: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Return successful response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "id": ride.ID.String(),
        "status": ride.Status,
        "departureTime": ride.DepartureTime,
    })
}

// GetRide retrieves a ride by ID
func (h *RideHandler) GetRide(w http.ResponseWriter, r *http.Request) {
    // Get ride ID from URL params
    vars := mux.Vars(r)
    rideIDStr, ok := vars["id"]
    if !ok {
        http.Error(w, "Missing ride ID", http.StatusBadRequest)
        return
    }

    rideID, err := uuid.Parse(rideIDStr)
    if err != nil {
        http.Error(w, "Invalid ride ID format", http.StatusBadRequest)
        return
    }

    // Get ride using service
    ride, err := h.rideService.GetRide(r.Context(), rideID)
    if err != nil {
        log.Printf("Error fetching ride: %v", err)
        http.Error(w, "Failed to get ride: "+err.Error(), http.StatusNotFound)
        return
    }

    // Return ride details
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ride)
}

// FindNearbyRides finds rides near a location
func (h *RideHandler) FindNearbyRides(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters
    query := r.URL.Query()
    
    latStr := query.Get("lat")
    lonStr := query.Get("lon")
    destLatStr := query.Get("destLat")
    destLonStr := query.Get("destLon")
    radiusStr := query.Get("radius")
    departureAfterStr := query.Get("departureAfter")
    
    // Validate required parameters
    if latStr == "" || lonStr == "" || destLatStr == "" || destLonStr == "" {
        http.Error(w, "Missing required location parameters", http.StatusBadRequest)
        return
    }
    
    // Parse location parameters
    lat, err := strconv.ParseFloat(latStr, 64)
    if err != nil {
        http.Error(w, "Invalid latitude format", http.StatusBadRequest)
        return
    }
    
    lon, err := strconv.ParseFloat(lonStr, 64)
    if err != nil {
        http.Error(w, "Invalid longitude format", http.StatusBadRequest)
        return
    }
    
    destLat, err := strconv.ParseFloat(destLatStr, 64)
    if err != nil {
        http.Error(w, "Invalid destination latitude format", http.StatusBadRequest)
        return
    }
    
    destLon, err := strconv.ParseFloat(destLonStr, 64)
    if err != nil {
        http.Error(w, "Invalid destination longitude format", http.StatusBadRequest)
        return
    }
    
    // Default radius to 5000 meters if not provided
    radius := 5000.0
    if radiusStr != "" {
        radius, err = strconv.ParseFloat(radiusStr, 64)
        if err != nil {
            http.Error(w, "Invalid radius format", http.StatusBadRequest)
            return
        }
        
        // Convert kilometers to meters if needed
        if radius < 100 {
            radius = radius * 1000
        }
    }
    
    // Default departure time to now if not provided
    departureAfter := time.Now()
    if departureAfterStr != "" {
        departureAfter, err = time.Parse(time.RFC3339, departureAfterStr)
        if err != nil {
            http.Error(w, "Invalid departure time format", http.StatusBadRequest)
            return
        }
    }
    
    // Find nearby rides using service
    rides, err := h.rideService.FindNearbyRides(
        r.Context(), lat, lon, destLat, destLon, radius, departureAfter)
    if err != nil {
        log.Printf("Error finding nearby rides: %v", err)
        http.Error(w, "Failed to find nearby rides", http.StatusInternalServerError)
        return
    }
    
    // Return rides
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(rides)
}

// CancelRide cancels a ride
func (h *RideHandler) CancelRide(w http.ResponseWriter, r *http.Request) {
    // Get user ID from context
    userID, err := middleware.GetUserIDFromContext(r.Context())
    if err != nil {
        log.Printf("Error getting user ID from context: %v", err)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Get ride ID from URL
    vars := mux.Vars(r)
    rideIDStr, ok := vars["id"]
    if !ok {
        http.Error(w, "Missing ride ID", http.StatusBadRequest)
        return
    }

    rideID, err := uuid.Parse(rideIDStr)
    if err != nil {
        http.Error(w, "Invalid ride ID", http.StatusBadRequest)
        return
    }

    // Cancel the ride
    if err := h.rideService.CancelRide(r.Context(), userID, rideID); err != nil {
        log.Printf("Error cancelling ride: %v", err)
        http.Error(w, "Failed to cancel ride: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Return success response
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "cancelled",
    })
}

// SearchRides searches for rides based on criteria
func (h *RideHandler) SearchRides(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters
    query := r.URL.Query()
    
    // Extract search parameters
    fromCity := query.Get("fromCity")
    toCity := query.Get("toCity")
    departureAfterStr := query.Get("departureAfter")
    
    // Validate required parameters
    if fromCity == "" || toCity == "" {
        http.Error(w, "Missing required city parameters", http.StatusBadRequest)
        return
    }
    
    // Default departure time to now if not provided
    departureAfter := time.Now()
    var err error
    if departureAfterStr != "" {
        departureAfter, err = time.Parse(time.RFC3339, departureAfterStr)
        if err != nil {
            http.Error(w, "Invalid departure time format", http.StatusBadRequest)
            return
        }
    }
    
    // Find rides using service
    rides, err := h.rideService.SearchRides(r.Context(), fromCity, toCity, departureAfter)
    if err != nil {
        log.Printf("Error searching rides: %v", err)
        http.Error(w, "Failed to search rides", http.StatusInternalServerError)
        return
    }
    
    // Return rides
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(rides)
}

// GetUserRides gets all rides for the authenticated user
func (h *RideHandler) GetUserRides(w http.ResponseWriter, r *http.Request) {
    // Get user ID from context (set by auth middleware)
    userID, err := middleware.GetUserIDFromContext(r.Context())
    if err != nil {
        log.Printf("Error getting user ID from context: %v", err)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Get rides using service
    rides, err := h.rideService.GetUserRides(r.Context(), userID)
    if err != nil {
        log.Printf("Error fetching user rides: %v", err)
        http.Error(w, "Failed to get user rides", http.StatusInternalServerError)
        return
    }

    // Return rides
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(rides)
}