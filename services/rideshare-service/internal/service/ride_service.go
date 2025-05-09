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

// RequestToJoinRide creates a new request to join a ride
func (s *RideService) RequestToJoinRide(ctx context.Context, rideID uuid.UUID, riderID uuid.UUID, req *models.JoinRideRequest) (*models.RideRequest, error) {
	// Validate the ride exists and has enough seats
	ride, err := s.GetRide(ctx, rideID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ride: %w", err)
	}

	// Check if ride is valid for join requests
	if ride.Status != "scheduled" {
		return nil, errors.New("cannot join a ride that is not scheduled")
	}

	if ride.AvailableSeats < req.SeatsRequested {
		return nil, errors.New("not enough available seats")
	}

	// Check if user is already in this ride
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM ride_passengers WHERE ride_id = $1 AND user_id = $2)"
	if err := s.dbManager.GetPrimary().QueryRowContext(ctx, checkQuery, rideID, riderID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("failed to check existing passenger: %w", err)
	}
	if exists {
		return nil, errors.New("you are already a passenger in this ride")
	}

	// Check for existing pending requests
	checkReqQuery := "SELECT EXISTS(SELECT 1 FROM ride_requests WHERE ride_id = $1 AND rider_id = $2 AND status = 'pending')"
	if err := s.dbManager.GetPrimary().QueryRowContext(ctx, checkReqQuery, rideID, riderID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("failed to check existing requests: %w", err)
	}
	if exists {
		return nil, errors.New("you already have a pending request for this ride")
	}

	// Create the request with a new UUID
	requestID := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO ride_requests (
			request_id, ride_id, rider_id, pickup_address, pickup_latitude, pickup_longitude,
			seats_requested, message, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING request_id, ride_id, rider_id, pickup_address, pickup_latitude, 
		    pickup_longitude, status, seats_requested, message, created_at, updated_at
	`

	var rideReq models.RideRequest
	var message sql.NullString
	if req.Message != "" {
		message.String = req.Message
		message.Valid = true
	}

	err = s.dbManager.GetPrimary().QueryRowContext(
		ctx, query, 
		requestID, rideID, riderID, req.PickupAddress, req.PickupLatitude, req.PickupLongitude,
		req.SeatsRequested, message, "pending", now, now,
	).Scan(
		&rideReq.ID, &rideReq.RideID, &rideReq.RiderID, &rideReq.PickupAddress,
		&rideReq.PickupLatitude, &rideReq.PickupLongitude, &rideReq.Status,
		&rideReq.SeatsRequested, &message, &rideReq.CreatedAt, &rideReq.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create ride request: %w", err)
	}

	if message.Valid {
		rideReq.Message = &message.String
	}

	return &rideReq, nil
}

// GetRequestsByRide retrieves all requests for a specific ride
func (s *RideService) GetRequestsByRide(ctx context.Context, rideID uuid.UUID, userID uuid.UUID) ([]*models.RideRequest, error) {
	// Verify the user is the host of this ride
	var isHost bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM rides WHERE ride_id = $1 AND host_id = $2)"
	if err := s.dbManager.GetReplica().QueryRowContext(ctx, checkQuery, rideID, userID).Scan(&isHost); err != nil {
		return nil, fmt.Errorf("failed to check ride host: %w", err)
	}

	if !isHost {
		return nil, errors.New("only the ride host can view requests")
	}

	// Get all requests for this ride
	query := `
		SELECT rr.request_id, rr.ride_id, rr.rider_id, rr.pickup_address, 
		       rr.pickup_latitude, rr.pickup_longitude, rr.status, rr.seats_requested, 
		       rr.message, rr.created_at, rr.updated_at,
		       u.first_name, u.last_name, u.average_rating
		FROM ride_requests rr
		JOIN users u ON rr.rider_id = u.user_id
		WHERE rr.ride_id = $1
		ORDER BY rr.created_at DESC
	`

	rows, err := s.dbManager.GetReplica().QueryContext(ctx, query, rideID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ride requests: %w", err)
	}
	defer rows.Close()

	var requests []*models.RideRequest
	for rows.Next() {
		var req models.RideRequest
		var message sql.NullString
		var firstName, lastName string
		var rating float64

		err := rows.Scan(
			&req.ID, &req.RideID, &req.RiderID, &req.PickupAddress,
			&req.PickupLatitude, &req.PickupLongitude, &req.Status, &req.SeatsRequested,
			&message, &req.CreatedAt, &req.UpdatedAt,
			&firstName, &lastName, &rating,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ride request: %w", err)
		}

		if message.Valid {
			req.Message = &message.String
		}

		requests = append(requests, &req)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ride requests: %w", err)
	}

	return requests, nil
}

// GetRequestsByUser retrieves all ride requests made by a user
func (s *RideService) GetRequestsByUser(ctx context.Context, userID uuid.UUID) ([]*models.RideRequest, error) {
	query := `
		SELECT rr.request_id, rr.ride_id, rr.rider_id, rr.pickup_address, 
		       rr.pickup_latitude, rr.pickup_longitude, rr.status, rr.seats_requested, 
		       rr.message, rr.created_at, rr.updated_at,
		       r.origin_address, r.destination_address, r.departure_time,
		       r.price_per_seat
		FROM ride_requests rr
		JOIN rides r ON rr.ride_id = r.ride_id
		WHERE rr.rider_id = $1
		ORDER BY rr.created_at DESC
	`

	rows, err := s.dbManager.GetReplica().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ride requests: %w", err)
	}
	defer rows.Close()

	var requests []*models.RideRequest
	for rows.Next() {
		var req models.RideRequest
		var message sql.NullString
		var originAddr, destAddr string
		var departureTime time.Time
		var pricePerSeat float64

		err := rows.Scan(
			&req.ID, &req.RideID, &req.RiderID, &req.PickupAddress,
			&req.PickupLatitude, &req.PickupLongitude, &req.Status, &req.SeatsRequested,
			&message, &req.CreatedAt, &req.UpdatedAt,
			&originAddr, &destAddr, &departureTime, &pricePerSeat,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user ride request: %w", err)
		}

		if message.Valid {
			req.Message = &message.String
		}

		requests = append(requests, &req)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user ride requests: %w", err)
	}

	return requests, nil
}

// UpdateRequestStatus changes a request's status (accept/reject)
func (s *RideService) UpdateRequestStatus(ctx context.Context, requestID, rideID, hostID uuid.UUID, status string) error {
	// Validate the status
	if status != "accepted" && status != "rejected" && status != "cancelled" {
		return errors.New("invalid status - must be 'accepted', 'rejected', or 'cancelled'")
	}

	// Start a transaction
	tx, err := s.dbManager.GetPrimary().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Verify the user is the host if accepting/rejecting, or the rider if cancelling
	var isAuthorized bool
	var checkQuery string
	if status == "cancelled" {
		checkQuery = "SELECT EXISTS(SELECT 1 FROM ride_requests WHERE request_id = $1 AND ride_id = $2 AND rider_id = $3)"
		if err := tx.QueryRowContext(ctx, checkQuery, requestID, rideID, hostID).Scan(&isAuthorized); err != nil {
			return fmt.Errorf("failed to check request: %w", err)
		}
	} else {
		checkQuery = "SELECT EXISTS(SELECT 1 FROM rides WHERE ride_id = $1 AND host_id = $2)"
		if err := tx.QueryRowContext(ctx, checkQuery, rideID, hostID).Scan(&isAuthorized); err != nil {
			return fmt.Errorf("failed to check ride host: %w", err)
		}
	}

	if !isAuthorized {
		return errors.New("you are not authorized to perform this action")
	}

	// Get the request details
	var req models.RideRequest
	reqQuery := "SELECT rider_id, seats_requested, status FROM ride_requests WHERE request_id = $1 AND ride_id = $2"
	if err := tx.QueryRowContext(ctx, reqQuery, requestID, rideID).Scan(&req.RiderID, &req.SeatsRequested, &req.Status); err != nil {
		return fmt.Errorf("failed to get request details: %w", err)
	}

	// Check if request is already in the requested status
	if req.Status == status {
		return fmt.Errorf("request is already %s", status)
	}

	// Check if request can be updated (must be pending)
	if req.Status != "pending" {
		return errors.New("only pending requests can be updated")
	}

	// Update the request status
	updateQuery := "UPDATE ride_requests SET status = $1, updated_at = $2 WHERE request_id = $3 AND ride_id = $4"
	now := time.Now()
	_, err = tx.ExecContext(ctx, updateQuery, status, now, requestID, rideID)
	if err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}

	// If accepting, also update available seats and add to ride_passengers
	if status == "accepted" {
		// Check if ride has enough seats
		var availableSeats int
		seatsQuery := "SELECT available_seats FROM rides WHERE ride_id = $1 FOR UPDATE"
		if err := tx.QueryRowContext(ctx, seatsQuery, rideID).Scan(&availableSeats); err != nil {
			return fmt.Errorf("failed to get available seats: %w", err)
		}

		if availableSeats < req.SeatsRequested {
			return errors.New("not enough available seats")
		}

		// Update available seats
		updateSeatsQuery := "UPDATE rides SET available_seats = available_seats - $1 WHERE ride_id = $2"
		_, err = tx.ExecContext(ctx, updateSeatsQuery, req.SeatsRequested, rideID)
		if err != nil {
			return fmt.Errorf("failed to update available seats: %w", err)
		}

		// Add to ride_passengers
		insertPassengerQuery := `
			INSERT INTO ride_passengers (
				ride_id, user_id, request_id, seats_taken, payment_status, created_at
			) VALUES ($1, $2, $3, $4, false, $5)
		`
		_, err = tx.ExecContext(ctx, insertPassengerQuery, rideID, req.RiderID, requestID, req.SeatsRequested, now)
		if err != nil {
			return fmt.Errorf("failed to add passenger: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetRidePassengers retrieves all passengers for a specific ride
func (s *RideService) GetRidePassengers(ctx context.Context, rideID uuid.UUID, userID uuid.UUID) ([]*models.RidePassenger, error) {
	// Verify the user is either the host or a passenger
	var isAuthorized bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM rides WHERE ride_id = $1 AND host_id = $2
			UNION
			SELECT 1 FROM ride_passengers WHERE ride_id = $1 AND user_id = $2
		)
	`
	if err := s.dbManager.GetReplica().QueryRowContext(ctx, checkQuery, rideID, userID).Scan(&isAuthorized); err != nil {
		return nil, fmt.Errorf("failed to check authorization: %w", err)
	}

	if !isAuthorized {
		return nil, errors.New("only the ride host or passengers can view passenger details")
	}

	// Get all passengers
	query := `
		SELECT 
			rp.ride_id, rp.user_id, rp.request_id, rp.seats_taken, 
			rp.pickup_time, rp.dropoff_time, rp.payment_status, rp.created_at,
			rr.pickup_address, rr.pickup_latitude, rr.pickup_longitude,
			u.first_name, u.last_name, u.profile_picture_url, u.average_rating
		FROM ride_passengers rp
		JOIN ride_requests rr ON rp.request_id = rr.request_id
		JOIN users u ON rp.user_id = u.user_id
		WHERE rp.ride_id = $1
		ORDER BY rp.created_at
	`

	rows, err := s.dbManager.GetReplica().QueryContext(ctx, query, rideID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ride passengers: %w", err)
	}
	defer rows.Close()

	var passengers []*models.RidePassenger
	for rows.Next() {
		var passenger models.RidePassenger
		var pickupTime, dropoffTime sql.NullTime
		var firstName, lastName string
		var profilePicture sql.NullString
		var rating float64

		err := rows.Scan(
			&passenger.RideID, &passenger.UserID, &passenger.RequestID, &passenger.SeatsTaken,
			&pickupTime, &dropoffTime, &passenger.PaymentStatus, &passenger.CreatedAt,
			&passenger.PickupAddress, &passenger.PickupLatitude, &passenger.PickupLongitude,
			&firstName, &lastName, &profilePicture, &rating,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan passenger: %w", err)
		}

		if pickupTime.Valid {
			passenger.PickupTime = &pickupTime.Time
		}
		if dropoffTime.Valid {
			passenger.DropoffTime = &dropoffTime.Time
		}

		passengers = append(passengers, &passenger)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating passengers: %w", err)
	}

	return passengers, nil
}

// SearchRides searches for rides based on from/to city and departure time
func (s *RideService) SearchRides(ctx context.Context, fromCity, toCity string, departureAfter time.Time) ([]*models.Ride, error) {
	query := `
		SELECT ride_id, host_id, vehicle_id, origin_address, origin_latitude, 
			origin_longitude, destination_address, destination_latitude, destination_longitude,
			departure_time, estimated_arrival_time, max_passengers, available_seats, 
			price_per_seat, status, description, luggage_capacity, is_pets_allowed, 
			is_smoking_allowed, created_at, updated_at
		FROM rides
		WHERE 
			origin_address ILIKE $1
			AND destination_address ILIKE $2
			AND departure_time > $3
			AND status = 'scheduled'
			AND available_seats > 0
		ORDER BY departure_time ASC
		LIMIT 50
	`

	// Use ILIKE for case-insensitive search with wildcards
	fromCityPattern := "%" + fromCity + "%"
	toCityPattern := "%" + toCity + "%"

	rows, err := s.dbManager.GetReplica().QueryContext(
		ctx,
		query,
		fromCityPattern,
		toCityPattern,
		departureAfter,
	)
	if err != nil {
		return nil, fmt.Errorf("error searching rides: %w", err)
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

// GetUserRides gets all rides for a specific user
func (s *RideService) GetUserRides(ctx context.Context, userID uuid.UUID) ([]*models.Ride, error) {
	query := `
		SELECT r.ride_id, r.host_id, r.vehicle_id, r.origin_address, r.origin_latitude, 
			r.origin_longitude, r.destination_address, r.destination_latitude, r.destination_longitude,
			r.departure_time, r.estimated_arrival_time, r.max_passengers, r.available_seats, 
			r.price_per_seat, r.status, r.description, r.luggage_capacity, r.is_pets_allowed, 
			r.is_smoking_allowed, r.created_at, r.updated_at
		FROM rides r
		WHERE r.host_id = $1 
		UNION
		SELECT r.ride_id, r.host_id, r.vehicle_id, r.origin_address, r.origin_latitude, 
			r.origin_longitude, r.destination_address, r.destination_latitude, r.destination_longitude,
			r.departure_time, r.estimated_arrival_time, r.max_passengers, r.available_seats, 
			r.price_per_seat, r.status, r.description, r.luggage_capacity, r.is_pets_allowed, 
			r.is_smoking_allowed, r.created_at, r.updated_at
		FROM rides r
		JOIN ride_passengers rp ON r.ride_id = rp.ride_id
		WHERE rp.user_id = $1
		ORDER BY departure_time DESC
	`

	rows, err := s.dbManager.GetReplica().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching user rides: %w", err)
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