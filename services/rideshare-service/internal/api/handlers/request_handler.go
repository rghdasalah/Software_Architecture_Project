package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/api/middleware"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/models"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/service"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// RequestHandler handles ride join request operations
type RequestHandler struct {
	rideService *service.RideService
}

// NewRequestHandler creates a new RequestHandler
func NewRequestHandler(rideService *service.RideService) *RequestHandler {
	return &RequestHandler{rideService: rideService}
}

// CreateJoinRequest handles a request to join a ride
func (h *RequestHandler) CreateJoinRequest(w http.ResponseWriter, r *http.Request) {
	// Extract ride ID from URL
	vars := mux.Vars(r)
	rideID, err := uuid.Parse(vars["rideId"])
	if err != nil {
		http.Error(w, "Invalid ride ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var joinReq models.JoinRideRequest
	if err := json.NewDecoder(r.Body).Decode(&joinReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate seats requested
	if joinReq.SeatsRequested <= 0 {
		http.Error(w, "Seats requested must be greater than zero", http.StatusBadRequest)
		return
	}

	// Set the ride ID from URL
	joinReq.RideID = rideID

	// Call service to create request
	rideRequest, err := h.rideService.RequestToJoinRide(r.Context(), rideID, userID, &joinReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create join request: %v", err), http.StatusBadRequest)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rideRequest)
}

// GetRideRequests returns all join requests for a ride
func (h *RequestHandler) GetRideRequests(w http.ResponseWriter, r *http.Request) {
	// Extract ride ID from URL
	vars := mux.Vars(r)
	rideID, err := uuid.Parse(vars["rideId"])
	if err != nil {
		http.Error(w, "Invalid ride ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Call service to get requests
	requests, err := h.rideService.GetRequestsByRide(r.Context(), rideID, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get ride requests: %v", err), http.StatusBadRequest)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// GetMyRequests returns all join requests made by the current user
func (h *RequestHandler) GetMyRequests(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Call service to get requests
	requests, err := h.rideService.GetRequestsByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get your requests: %v", err), http.StatusBadRequest)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// UpdateRequestStatus handles accepting/rejecting/cancelling requests
func (h *RequestHandler) UpdateRequestStatus(w http.ResponseWriter, r *http.Request) {
	// Extract IDs from URL
	vars := mux.Vars(r)
	rideID, err := uuid.Parse(vars["rideId"])
	if err != nil {
		http.Error(w, "Invalid ride ID", http.StatusBadRequest)
		return
	}

	requestID, err := uuid.Parse(vars["requestId"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode request body
	var statusUpdate models.RequestStatusUpdate
	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call service to update status
	err = h.rideService.UpdateRequestStatus(r.Context(), requestID, rideID, userID, statusUpdate.Status)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update request status: %v", err), http.StatusBadRequest)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Request status updated successfully"}`))
}

// GetRidePassengers returns all confirmed passengers for a ride
func (h *RequestHandler) GetRidePassengers(w http.ResponseWriter, r *http.Request) {
	// Extract ride ID from URL
	vars := mux.Vars(r)
	rideID, err := uuid.Parse(vars["rideId"])
	if err != nil {
		http.Error(w, "Invalid ride ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Call service to get passengers
	passengers, err := h.rideService.GetRidePassengers(r.Context(), rideID, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get ride passengers: %v", err), http.StatusBadRequest)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(passengers)
}