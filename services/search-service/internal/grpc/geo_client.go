package grpc

import (
	"context"
	"log"
	"time"

	"search-service/internal/config"
	pb "search-service/proto"

	"google.golang.org/grpc"
)

type GeoClient struct {
	client pb.GeoDistanceServiceClient
}

func NewGeoClient(cfg *config.Config) *GeoClient {
	conn, err := grpc.Dial(cfg.GeoGRPCAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to geo-distance-service: %v", err)
	}

	client := pb.NewGeoDistanceServiceClient(conn)
	return &GeoClient{client: client}
}

func (g *GeoClient) FilterRides(ctx context.Context, point *pb.Point, rides []*pb.Ride, matchType pb.MatchType) ([]*pb.Ride, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.FilterRequest{
		Point:     point,
		Rides:     rides,
		MatchType: matchType,
	}

	resp, err := g.client.FilterRidesByProximity(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Rides, nil
}
