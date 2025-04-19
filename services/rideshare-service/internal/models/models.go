package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user in the system
type UserRole string

// RideStatus represents the status of a ride
type RideStatus string

// RequestStatus represents the status of a ride request
type RequestStatus string

const (
	RoleRider UserRole = "rider"
	RoleDriver UserRole = "driver"
	RoleAdmin  UserRole = "admin"

	StatusScheduled   RideStatus = "scheduled"
	StatusInProgress  RideStatus = "in_progress"
	StatusCompleted   RideStatus = "completed"
	StatusCancelled   RideStatus = "cancelled"

	RequestPending   RequestStatus = "pending"
	RequestAccepted  RequestStatus = "accepted"
	RequestRejected  RequestStatus = "rejected"
	RequestCancelled RequestStatus = "cancelled"
)

// User represents a user in the system
type User struct {
	ID              uuid.UUID  `json:"id" db:"user_id"`
	Email           string     `json:"email" db:"email"`
	PasswordHash    string     `json:"-" db:"password_hash"`
	FirstName       string     `json:"firstName" db:"first_name"`
	LastName        string     `json:"lastName" db:"last_name"`
	PhoneNumber     string     `json:"phoneNumber" db:"phone_number"`
	Role            string     `json:"role" db:"role"`
	ProfilePicture  *string    `json:"profilePicture,omitempty" db:"profile_picture_url"`
	DateOfBirth     time.Time  `json:"dateOfBirth" db:"date_of_birth"`
	Bio             *string    `json:"bio,omitempty" db:"bio"`
	AverageRating   float64    `json:"averageRating" db:"average_rating"`
	IsVerified      bool       `json:"isVerified" db:"is_verified"`
	IsActive        bool       `json:"isActive" db:"is_active"`
	CreatedAt       time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time  `json:"updatedAt" db:"updated_at"`
	LastLoginAt     *time.Time `json:"lastLoginAt,omitempty" db:"last_login_at"`
}

// Vehicle represents a vehicle in the system
type Vehicle struct {
	ID          uuid.UUID `json:"vehicleId" db:"vehicle_id"`
	UserID      uuid.UUID `json:"userId" db:"user_id"`
	Make        string    `json:"make" db:"make"`
	Model       string    `json:"model" db:"model"`
	Year        int       `json:"year" db:"year"`
	Color       string    `json:"color" db:"color"`
	LicensePlate string   `json:"licensePlate" db:"license_plate"`
	Capacity    int       `json:"capacity" db:"capacity"`
	IsActive    bool      `json:"isActive" db:"is_active"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// Ride represents a ride in the system
type Ride struct {
	ID                  uuid.UUID  `json:"id" db:"ride_id"`
	HostID              uuid.UUID  `json:"hostId" db:"host_id"`
	VehicleID           uuid.UUID  `json:"vehicleId" db:"vehicle_id"`
	OriginAddress       string     `json:"originAddress" db:"origin_address"`
	OriginLatitude      float64    `json:"originLatitude" db:"origin_latitude"`
	OriginLongitude     float64    `json:"originLongitude" db:"origin_longitude"`
	DestinationAddress  string     `json:"destinationAddress" db:"destination_address"`
	DestinationLatitude float64    `json:"destinationLatitude" db:"destination_latitude"`
	DestinationLongitude float64   `json:"destinationLongitude" db:"destination_longitude"`
	DepartureTime       time.Time  `json:"departureTime" db:"departure_time"`
	EstimatedArrivalTime time.Time `json:"estimatedArrivalTime" db:"estimated_arrival_time"`
	MaxPassengers       int        `json:"maxPassengers" db:"max_passengers"`
	AvailableSeats      int        `json:"availableSeats" db:"available_seats"`
	PricePerSeat        float64    `json:"pricePerSeat" db:"price_per_seat"`
	RoutePolyline       *string    `json:"routePolyline,omitempty" db:"route_polyline"`
	Status              string     `json:"status" db:"status"`
	Description         *string    `json:"description,omitempty" db:"description"`
	LuggageCapacity     *string    `json:"luggageCapacity,omitempty" db:"luggage_capacity"`
	IsPetsAllowed       bool       `json:"isPetsAllowed" db:"is_pets_allowed"`
	IsSmokingAllowed    bool       `json:"isSmokingAllowed" db:"is_smoking_allowed"`
	CreatedAt           time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time  `json:"updatedAt" db:"updated_at"`
}

// RideRequest represents a request to join a ride
type RideRequest struct {
	ID               uuid.UUID  `json:"requestId" db:"request_id"`
	RideID           uuid.UUID  `json:"rideId" db:"ride_id"`
	RiderID          uuid.UUID  `json:"riderId" db:"rider_id"`
	PickupAddress    string     `json:"pickupAddress" db:"pickup_address"`
	PickupLatitude   float64    `json:"pickupLatitude" db:"pickup_latitude"`
	PickupLongitude  float64    `json:"pickupLongitude" db:"pickup_longitude"`
	DropoffAddress   *string    `json:"dropoffAddress,omitempty" db:"dropoff_address"`
	DropoffLatitude  *float64   `json:"dropoffLatitude,omitempty" db:"dropoff_latitude"`
	DropoffLongitude *float64   `json:"dropoffLongitude,omitempty" db:"dropoff_longitude"`
	Status           string     `json:"status" db:"status"`
	SeatsRequested   int        `json:"seatsRequested" db:"seats_requested"`
	DistanceAdded    *float64   `json:"distanceAdded,omitempty" db:"distance_added_meters"`
	Message          *string    `json:"message,omitempty" db:"message"`
	CreatedAt        time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time  `json:"updatedAt" db:"updated_at"`
}

// RidePassenger represents a confirmed passenger on a ride
type RidePassenger struct {
	RideID       uuid.UUID  `json:"rideId" db:"ride_id"`
	UserID       uuid.UUID  `json:"userId" db:"user_id"`
	RequestID    uuid.UUID  `json:"requestId" db:"request_id"`
	SeatsTaken   int        `json:"seatsTaken" db:"seats_taken"`
	PickupTime   *time.Time `json:"pickupTime,omitempty" db:"pickup_time"`
	DropoffTime  *time.Time `json:"dropoffTime,omitempty" db:"dropoff_time"`
	PaymentStatus bool      `json:"paymentStatus" db:"payment_status"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
}

// Rating represents a rating for a user
type Rating struct {
	ID        uuid.UUID `json:"id" db:"rating_id"`
	RideID    uuid.UUID `json:"rideId" db:"ride_id"`
	RaterID   uuid.UUID `json:"raterId" db:"rater_id"`
	RatedID   uuid.UUID `json:"ratedId" db:"rated_id"`
	Rating    float64   `json:"rating" db:"rating"`
	Comment   *string   `json:"comment,omitempty" db:"comment"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// NearbyRideResult represents a ride that is near a location
type NearbyRideResult struct {
	Ride                 Ride    `json:"ride"`
	DistanceFromOrigin   float64 `json:"distanceFromOrigin"`
	DistanceFromDestination float64 `json:"distanceFromDestination"`
}

// UserAuthResponse represents a response to a successful authentication
type UserAuthResponse struct {
	Token  string    `json:"token"`
	User   User      `json:"user"`
}

// CreateVehicleRequest represents a request to create a vehicle
type CreateVehicleRequest struct {
	Make         string `json:"make"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	Color        string `json:"color"`
	LicensePlate string `json:"licensePlate"`
	Capacity     int    `json:"capacity"`
}

// CreateRideRequest represents a request to create a ride
type CreateRideRequest struct {
	VehicleID           uuid.UUID `json:"vehicleId"`
	OriginAddress       string    `json:"originAddress"`
	OriginLatitude      float64   `json:"originLatitude"`
	OriginLongitude     float64   `json:"originLongitude"`
	DestinationAddress  string    `json:"destinationAddress"`
	DestinationLatitude float64   `json:"destinationLatitude"`
	DestinationLongitude float64  `json:"destinationLongitude"`
	DepartureTime       string    `json:"departureTime"` // ISO8601 string
	EstimatedArrivalTime string   `json:"estimatedArrivalTime"` // ISO8601 string
	MaxPassengers       int       `json:"maxPassengers"`
	AvailableSeats      int       `json:"availableSeats"`
	PricePerSeat        float64   `json:"pricePerSeat"`
	Description         string    `json:"description,omitempty"`
	LuggageCapacity     string    `json:"luggageCapacity,omitempty"`
	IsPetsAllowed       bool      `json:"isPetsAllowed,omitempty"`
	IsSmokingAllowed    bool      `json:"isSmokingAllowed,omitempty"`
}