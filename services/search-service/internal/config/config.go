package config

import (
	"os"
)

type Config struct {
	GeoGRPCAddress string
	RedisAddress   string
}

func LoadConfig() (*Config, error) {
	geoAddr := os.Getenv("GEO_GRPC_ADDRESS")
	if geoAddr == "" {
		geoAddr = "geo-distance-service:50051"
		//return nil, fmt.Errorf("GEO_GRPC_ADDRESS not set")
	}

	redisAddr := os.Getenv("REDIS_ADDRESS")
	if redisAddr == "" {
		redisAddr = "redis:6379"
		//return nil, fmt.Errorf("REDIS_ADDRESS not set")
	}

	return &Config{
		GeoGRPCAddress: geoAddr,
		RedisAddress:   redisAddr,
	}, nil
}
