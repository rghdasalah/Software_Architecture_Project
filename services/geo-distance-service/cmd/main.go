package main

import (
	"log"
	"net"

	"geo-distance-service/internal/config"
	"geo-distance-service/internal/transport"
	pb "geo-distance-service/proto"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterGeoDistanceServiceServer(server, transport.NewGeoServer(cfg))

	log.Printf("GeoDistanceService listening on port %s with radius %.2fkm", cfg.GRPCPort, cfg.RadiusKM)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
