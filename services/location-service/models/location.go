package models

import "time"

// LocationUpdate represents a location update for a ride
type LocationUpdate struct {
    RideID    int       `json:"rideID" bson:"ride_id"`
    Latitude  float64   `json:"latitude" bson:"latitude"`
    Longitude float64   `json:"longitude" bson:"longitude"`
    Timestamp time.Time `json:"timestamp" bson:"timestamp"`
    ExpiresAt time.Time `json:"expiresAt" bson:"expires_at"`
}