package db

import "search-service/proto"

// RideRepo defines an interface for ride data access
type RideRepo interface {
	GetAllAvailableRides() ([]*proto.Ride, error)
}

type MockRideRepo struct{}

// NewMockRideRepo returns a new instance of MockRideRepo
func NewMockRideRepo() RideRepo {
	return &MockRideRepo{}
}

// GetAllAvailableRides returns a mocked list of rides
func (r *MockRideRepo) GetAllAvailableRides() ([]*proto.Ride, error) {
	rides := []*proto.Ride{
		{
			RideId:     "ride_1",
			StartPoint: &proto.Point{Lat: 30.0444, Lng: 31.2357}, // Cairo
			EndPoint:   &proto.Point{Lat: 30.0333, Lng: 31.2333}, // Nearby
		},
		{
			RideId:     "ride_2",
			StartPoint: &proto.Point{Lat: 29.9792, Lng: 31.1342}, // Giza
			EndPoint:   &proto.Point{Lat: 30.0500, Lng: 31.2333}, // Downtown
		},
		{
			RideId:     "ride_3",
			StartPoint: &proto.Point{Lat: 31.2001, Lng: 29.9187}, // Alexandria
			EndPoint:   &proto.Point{Lat: 30.0444, Lng: 31.2357}, // Cairo
		},
	}

	return rides, nil
}

/**
package db

import (
	"context"
	"fmt"

	pb "search-service/proto"

	"github.com/jackc/pgx/v5"
)

type RideRepository struct {
	Conn *pgx.Conn
}

func NewRideRepository(dsn string) (*RideRepository, error) {
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	return &RideRepository{Conn: conn}, nil
}

func (r *RideRepository) GetAllRides(ctx context.Context) ([]*pb.Ride, error) {
	rows, err := r.Conn.Query(ctx, "SELECT ride_id, start_lat, start_lng, end_lat, end_lng FROM rides")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rides []*pb.Ride
	for rows.Next() {
		var rideID string
		var startLat, startLng, endLat, endLng float64

		if err := rows.Scan(&rideID, &startLat, &startLng, &endLat, &endLng); err != nil {
			return nil, err
		}

		rides = append(rides, &pb.Ride{
			RideId:     rideID,
			StartPoint: &pb.Point{Lat: startLat, Lng: startLng},
			EndPoint:   &pb.Point{Lat: endLat, Lng: endLng},
		})
	}
	return rides, nil
}
**/
