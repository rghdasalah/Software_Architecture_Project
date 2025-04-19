package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/db"
	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type RideService struct {
	dbManager *db.DBManager
}

func NewRideService(dbManager *db.DBManager) *RideService {
	return &RideService{dbManager: dbManager}
}

// CreateRide inserts a new ride into the database
func (s *RideService) CreateRide(ctx context.Context, hostID uuid.UUID, req *models.CreateRideRequest) (*models.Ride, error) {
	// Parse time strings
	departureTime, err := time.Parse(time.RFC3339, req.DepartureTime)
	if err != nil {
		return nil, fmt.Errorf("invalid departure time format: %w", err)
	}

	estimatedArrivalTime, err := time.Parse(time.RFC3339, req.EstimatedArrivalTime)
	if err != nil {
		return nil, fmt.Errorf("invalid estimated arrival time format: %w", err)
	}

	// Validate that departure time is in the future
	if departureTime.Before(time.Now()) {
		return nil, errors.New("departure time must be in the future")
	}

	// Validate that estimated arrival is after departure
	if estimatedArrivalTime.Before(departureTime) {
		return nil, errors.New("estimated arrival time must be after departure time")
	}

	// Validate capacity
	if req.MaxPassengers <= 0 {
		return nil, errors.New("maximum passengers must be greater than zero")
	}

	if req.AvailableSeats > req.MaxPassengers {
		return nil, errors.New("available seats cannot exceed maximum passengers")
	}

	// Insert into database
	query := `
		INSERT INTO rides (
			host_id, vehicle_id, origin_address, origin_latitude, origin_longitude,
			destination_address, destination_latitude, destination_longitude, 
			departure_time, estimated_arrival_time, max_passengers, available_seats,
			price_per_seat, status, description, luggage_capacity, is_pets_allowed, 
			is_smoking_allowed
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
		RETURNING ride_id, host_id, vehicle_id, origin_address, origin_latitude, 
			origin_longitude, destination_address, destination_latitude, destination_longitude,
			departure_time, estimated_arrival_time, max_passengers, available_seats, 
			price_per_seat, status, description, luggage_capacity, is_pets_allowed, 
			is_smoking_allowed, created_at, updated_at
	`

	var ride models.Ride
	var description, luggageCapacity sql.NullString

	if req.Description != "" {
		description.String = req.Description
		description.Valid = true
	}

	if req.LuggageCapacity != "" {
		luggageCapacity.String = req.LuggageCapacity
		luggageCapacity.Valid = true
	}

	err = s.dbManager.GetPrimary().QueryRowContext(
		ctx,
		query,
		hostID,                      // $1
		req.VehicleID,               // $2
		req.OriginAddress,           // $3
		req.OriginLatitude,          // $4
		req.OriginLongitude,         // $5
		req.DestinationAddress,      // $6
		req.DestinationLatitude,     // $7
		req.DestinationLongitude,    // $8
		departureTime,               // $9
		estimatedArrivalTime,        // $10
		req.MaxPassengers,           // $11
		req.AvailableSeats,          // $12
		req.PricePerSeat,            // $13
		"scheduled",                 // $14
		description,                 // $15
		luggageCapacity,             // $16
		req.IsPetsAllowed,           // $17
		req.IsSmokingAllowed,        // $18
	).Scan(
		&ride.ID,
		&ride.HostID,
		&ride.VehicleID,
		&ride.OriginAddress,
		&ride.OriginLatitude,
		&ride.OriginLongitude,
		&ride.DestinationAddress,
		&ride.DestinationLatitude,
		&ride.DestinationLongitude,
		&ride.DepartureTime,
		&ride.EstimatedArrivalTime,
		&ride.MaxPassengers,
		&ride.AvailableSeats,
		&ride.PricePerSeat,
		&ride.Status,
		&ride.Description,
		&ride.LuggageCapacity,
		&ride.IsPetsAllowed,
		&ride.IsSmokingAllowed,
		&ride.CreatedAt,
		&ride.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create ride: %w", err)
	}

	return &ride, nil
}

// GetRide fetches a ride by ID
func (s *RideService) GetRide(ctx context.Context, rideID uuid.UUID) (*models.Ride, error) {
	query := `
		SELECT ride_id, host_id, vehicle_id, origin_address, origin_latitude, 
			origin_longitude, destination_address, destination_latitude, destination_longitude,
			departure_time, estimated_arrival_time, max_passengers, available_seats, 
			price_per_seat, status, description, luggage_capacity, is_pets_allowed, 
			is_smoking_allowed, created_at, updated_at
		FROM rides
		WHERE ride_id = $1
	`

	var ride models.Ride
	var description, luggageCapacity sql.NullString

	err := s.dbManager.GetReplica().QueryRowContext(ctx, query, rideID).Scan(
		&ride.ID,
		&ride.HostID,
		&ride.VehicleID,
		&ride.OriginAddress,
		&ride.OriginLatitude,
		&ride.OriginLongitude,
		&ride.DestinationAddress,
		&ride.DestinationLatitude,
		&ride.DestinationLongitude,
		&ride.DepartureTime,
		&ride.EstimatedArrivalTime,
		&ride.MaxPassengers,
		&ride.AvailableSeats,
		&ride.PricePerSeat,
		&ride.Status,
		&description,
		&luggageCapacity,
		&ride.IsPetsAllowed,
		&ride.IsSmokingAllowed,
		&ride.CreatedAt,
		&ride.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("ride not found")
	} else if err != nil {
		return nil, fmt.Errorf("error fetching ride: %w", err)
	}

	// Handle null strings
	if description.Valid {
		descStr := description.String
		ride.Description = &descStr
	}

	if luggageCapacity.Valid {
		capacityStr := luggageCapacity.String
		ride.LuggageCapacity = &capacityStr
	}

	return &ride, nil
}

// FindNearbyRides uses the database function to find rides near a location
func (s *RideService) FindNearbyRides(ctx context.Context, lat, lon, destLat, destLon, radiusMeters float64, departureAfter time.Time) ([]*models.Ride, error) {
	// Use the PostgreSQL function we defined in the schema
	query := `
		SELECT * FROM find_nearby_rides($1, $2, $3, $4, $5, $6)
	`

	rows, err := s.dbManager.GetReplica().QueryContext(
		ctx,
		query,
		lat,
		lon,
		destLat,
		destLon,
		radiusMeters,
		departureAfter,
	)
	if err != nil {
		return nil, fmt.Errorf("error finding nearby rides: %w", err)
	}
	defer rows.Close()

	// Process results
	type nearbyRideResult struct {
		RideID               uuid.UUID
		HostID               uuid.UUID
		OriginAddress        string
		DestinationAddress   string
		DepartureTime        time.Time
		AvailableSeats       int
		DistanceFromOrigin   float64
		DistanceFromDestination float64
	}

	var resultIDs []uuid.UUID
	nearbyResults := make(map[uuid.UUID]nearbyRideResult)

	for rows.Next() {
		var result nearbyRideResult
		if err := rows.Scan(
			&result.RideID,
			&result.HostID,
			&result.OriginAddress,
			&result.DestinationAddress,
			&result.DepartureTime,
			&result.AvailableSeats,
			&result.DistanceFromOrigin,
			&result.DistanceFromDestination,
		); err != nil {
			return nil, fmt.Errorf("error scanning nearby ride: %w", err)
		}

		resultIDs = append(resultIDs, result.RideID)
		nearbyResults[result.RideID] = result
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through nearby rides: %w", err)
	}

	// If no rides found, return empty slice
	if len(resultIDs) == 0 {
		return []*models.Ride{}, nil
	}

	// Fetch full ride details for all matched rides
	rides, err := s.GetRidesByIDs(ctx, resultIDs)
	if err != nil {
		return nil, err
	}

	return rides, nil
}

// GetRidesByIDs fetches multiple rides by their IDs
func (s *RideService) GetRidesByIDs(ctx context.Context, rideIDs []uuid.UUID) ([]*models.Ride, error) {
	// Build a query with multiple UUID parameters
	query := `
		SELECT ride_id, host_id, vehicle_id, origin_address, origin_latitude, 
			origin_longitude, destination_address, destination_latitude, destination_longitude,
			departure_time, estimated_arrival_time, max_passengers, available_seats, 
			price_per_seat, status, description, luggage_capacity, is_pets_allowed, 
			is_smoking_allowed, created_at, updated_at
		FROM rides
		WHERE ride_id = ANY($1)
	`

	// Convert []uuid.UUID to []interface{} for the query
	idArgs := make([]interface{}, len(rideIDs))
	for i, id := range rideIDs {
		idArgs[i] = id
	}

	rows, err := s.dbManager.GetReplica().QueryContext(ctx, query, pq.Array(rideIDs))
	if err != nil {
		return nil, fmt.Errorf("error fetching rides by IDs: %w", err)
	}
	defer rows.Close()

	rides := []*models.Ride{}
	for rows.Next() {
		var ride models.Ride
		var description, luggageCapacity sql.NullString

		if err := rows.Scan(
			&ride.ID,
			&ride.HostID,
			&ride.VehicleID,
			&ride.OriginAddress,
			&ride.OriginLatitude,
			&ride.OriginLongitude,
			&ride.DestinationAddress,
			&ride.DestinationLatitude,
			&ride.DestinationLongitude,
			&ride.DepartureTime,
			&ride.EstimatedArrivalTime,
			&ride.MaxPassengers,
			&ride.AvailableSeats,
			&ride.PricePerSeat,
			&ride.Status,
			&description,
			&luggageCapacity,
			&ride.IsPetsAllowed,
			&ride.IsSmokingAllowed,
			&ride.CreatedAt,
			&ride.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error scanning ride: %w", err)
		}

		// Handle null strings
		if description.Valid {
			descStr := description.String
			ride.Description = &descStr
		}

		if luggageCapacity.Valid {
			capacityStr := luggageCapacity.String
			ride.LuggageCapacity = &capacityStr
		}

		rides = append(rides, &ride)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through rides: %w", err)
	}

	return rides, nil
}

// CancelRide cancels a ride (changes status to cancelled)
func (s *RideService) CancelRide(ctx context.Context, userID, rideID uuid.UUID) error {
	// First check if the ride exists and belongs to the user
	ride, err := s.GetRide(ctx, rideID)
	if err != nil {
		return err
	}

	if ride.HostID != userID {
		return errors.New("unauthorized: only the host can cancel this ride")
	}

	if ride.Status != "scheduled" {
		return fmt.Errorf("ride cannot be cancelled from status: %s", ride.Status)
	}

	// Update the ride status
	query := `
		UPDATE rides
		SET status = 'cancelled', updated_at = NOW()
		WHERE ride_id = $1
	`

	_, err = s.dbManager.GetPrimary().ExecContext(ctx, query, rideID)
	if err != nil {
		return fmt.Errorf("error cancelling ride: %w", err)
	}

	// Note: In a real system, we would also notify passengers, handle refunds, etc.

	return nil
}