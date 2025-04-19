package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/api/middleware"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/models"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type VehicleHandler struct {
	vehicleService *service.VehicleService
}

func NewVehicleHandler(vehicleService *service.VehicleService) *VehicleHandler {
	return &VehicleHandler{vehicleService: vehicleService}
}

// CreateVehicle handles creating a new vehicle for the authenticated user
func (h *VehicleHandler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if (err != nil) {
		log.Printf("Error getting user ID from context: %v", err)
		
		// Return detailed error for debugging
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Authentication error: " + err.Error(),
			"details": "Failed to get user ID from request context. Make sure your token is valid and properly formatted.",
		})
		return
	}
	
	log.Printf("Creating vehicle for user ID: %s", userID.String())

	// Parse request body
	var req models.CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request format: " + err.Error(),
			"details": "Request body could not be parsed. Ensure you're sending valid JSON.",
		})
		return
	}
	
	// Log the request data
	reqData, _ := json.Marshal(req)
	log.Printf("Vehicle creation request data: %s", string(reqData))

	// Create vehicle using service
	vehicle, err := h.vehicleService.CreateVehicle(r.Context(), userID, &req)
	if err != nil {
		log.Printf("Error creating vehicle: %v", err)
		
		// Return a properly formatted JSON error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := map[string]string{
			"error": "Failed to create vehicle: " + err.Error(),
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Return successful response using direct marshaling to avoid any encoding issues
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	
	response := map[string]string{
		"vehicleId": vehicle.ID.String(),
		"message": "Vehicle created successfully",
	}
	
	// Debug log the response before sending
	responseBytes, _ := json.Marshal(response)
	log.Printf("DEBUG: Sending vehicle creation response: %s", string(responseBytes))
	
	// Write the response
	json.NewEncoder(w).Encode(response)
}

// GetUserVehiclesForAuthUser gets vehicles for the authenticated user
func (h *VehicleHandler) GetUserVehiclesForAuthUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Printf("Error getting user ID from context: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get vehicles using service
	vehicles, err := h.vehicleService.GetVehiclesByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching vehicles: %v", err)
		http.Error(w, "Failed to get vehicles", http.StatusInternalServerError)
		return
	}

	// Return vehicles
	w.Header().Set("Content-Type", "application/json")
	vehicleBytes, err := json.Marshal(vehicles)
	if err != nil {
		log.Printf("Error marshalling vehicle list: %v", err)
		http.Error(w, "Failed to serialize vehicles", http.StatusInternalServerError)
		return
	}
	w.Write(vehicleBytes)
}

// GetUserVehicles gets vehicles for a specific user ID (public endpoint)
func (h *VehicleHandler) GetUserVehicles(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL params
	vars := mux.Vars(r)
	userIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get vehicles using service
	vehicles, err := h.vehicleService.GetVehiclesByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching vehicles: %v", err)
		http.Error(w, "Failed to get vehicles", http.StatusInternalServerError)
		return
	}

	// Return vehicles
	w.Header().Set("Content-Type", "application/json")
	vehicleBytes, err := json.Marshal(vehicles)
	if err != nil {
		log.Printf("Error marshalling vehicle list: %v", err)
		http.Error(w, "Failed to serialize vehicles", http.StatusInternalServerError)
		return
	}
	w.Write(vehicleBytes)
}

// DeleteVehicle soft deletes a vehicle
func (h *VehicleHandler) DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Printf("Error getting user ID from context: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get vehicle ID from URL
	vars := mux.Vars(r)
	vehicleIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing vehicle ID", http.StatusBadRequest)
		return
	}

	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	// Delete the vehicle
	if err := h.vehicleService.DeleteVehicle(r.Context(), userID, vehicleID); err != nil {
		log.Printf("Error deleting vehicle: %v", err)
		http.Error(w, "Failed to delete vehicle", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}