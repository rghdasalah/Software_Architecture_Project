package transport

import (
	"encoding/json"
	"net/http"

	"search-service/internal/db"

	"search-service/internal/service"
	pb "search-service/proto"
)

type Handler struct {
	SearchService *service.SearchService
	RideRepo      *db.MockRideRepo
}

func NewHandler(searchService *service.SearchService, rideRepo *db.MockRideRepo) *Handler {
	return &Handler{SearchService: searchService, RideRepo: rideRepo}
}

type searchRequest struct {
	StartLat float64 `json:"start_lat"`
	StartLng float64 `json:"start_lng"`
	EndLat   float64 `json:"end_lat"`
	EndLng   float64 `json:"end_lng"`
}

func (h *Handler) SearchRidesHandler(w http.ResponseWriter, r *http.Request) {
	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	input := service.SearchInput{
		Start: &pb.Point{Lat: req.StartLat, Lng: req.StartLng},
		End:   &pb.Point{Lat: req.EndLat, Lng: req.EndLng},
	}

	// Fetch all available rides from the repository
	allRides, err := h.RideRepo.GetAllAvailableRides()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Call SearchService to get the filtered rides
	matches, err := h.SearchService.SearchRides(r.Context(), input, allRides)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the filtered list of rides in the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}
