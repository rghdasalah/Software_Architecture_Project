package transport

import (
	"context"
	"geo-distance-service/internal/config"
	"geo-distance-service/internal/service"
	pb "geo-distance-service/proto"
)

type GeoServer struct {
	pb.UnimplementedGeoDistanceServiceServer
	Config *config.Config
}

func NewGeoServer(cfg *config.Config) *GeoServer {
	return &GeoServer{Config: cfg}
}

func (s *GeoServer) FilterRidesByProximity(ctx context.Context, req *pb.FilterRequest) (*pb.FilterResponse, error) {
	filtered := service.FilterRidesByProximity(s.Config, req)
	return &pb.FilterResponse{Rides: filtered}, nil
}
