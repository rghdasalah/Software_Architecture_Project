package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// RegisterRequest represents a request to register a new user
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	PhoneNumber string `json:"phoneNumber"`
	DateOfBirth string `json:"dateOfBirth"` // ISO format: YYYY-MM-DD
}

// LoginRequest represents a request to login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterUser handles user registration
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	// Check content type
	contentType := r.Header.Get("Content-Type")
	log.Printf("RegisterUser received Content-Type: %s", contentType)
	
	// Parse request body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body: " + err.Error(),
			"details": "Make sure you're sending properly formatted JSON with Content-Type: application/json",
		})
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.PhoneNumber == "" || req.DateOfBirth == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Missing required fields",
		})
		return
	}

	// Parse date of birth
	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid date format for date of birth (use YYYY-MM-DD)",
		})
		return
	}

	// Check if user is at least 18 years old
	minAge := 18
	if time.Now().AddDate(-minAge, 0, 0).Before(dob) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "User must be at least 18 years old",
		})
		return
	}

	// Create user using service
	user, err := h.userService.RegisterUser(
		r.Context(),
		req.Email,
		req.Password,
		req.FirstName,
		req.LastName,
		req.PhoneNumber,
		dob,
	)

	if err != nil {
		log.Printf("Error registering user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to register user: " + err.Error(),
		})
		return
	}

	// For testing compatibility, use the user ID as the token
	// In production, we would use a proper JWT token
	token := user.ID.String()

	// Return successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	
	response := map[string]interface{}{
		"userId": user.ID.String(),
		"token":  token,
	}
	
	// Debug log
	responseBytes, _ := json.Marshal(response)
	log.Printf("RegisterUser response: %s", string(responseBytes))
	
	json.NewEncoder(w).Encode(response)
}

// LoginUser handles user login
func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	// Check content type
	contentType := r.Header.Get("Content-Type")
	log.Printf("LoginUser received Content-Type: %s", contentType)
	
	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body: " + err.Error(),
			"details": "Make sure you're sending properly formatted JSON with Content-Type: application/json",
		})
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Missing email or password",
		})
		return
	}

	// Authenticate user using service
	user, err := h.userService.LoginUser(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Printf("Error authenticating user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid email or password",
		})
		return
	}

	// For testing compatibility, use the user ID as the token
	// In production, we would use a proper JWT token
	token := user.ID.String()

	// Return successful response with token
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"token":     token,
		"userId":    user.ID.String(),
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"email":     user.Email,
	}
	
	// Debug log
	responseBytes, _ := json.Marshal(response)
	log.Printf("LoginUser response: %s", string(responseBytes))
	
	json.NewEncoder(w).Encode(response)
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	vars := mux.Vars(r)
	userIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Get user using service
	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		http.Error(w, "Failed to get user: "+err.Error(), http.StatusNotFound)
		return
	}

	// Return user details (excluding sensitive data)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        user.ID.String(),
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"email":     user.Email,
		"role":      user.Role,
		"rating":    user.AverageRating,
	})
}