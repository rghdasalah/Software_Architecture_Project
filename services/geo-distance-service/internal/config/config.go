package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	GRPCPort string
	RadiusKM float64
}

func Load() *Config {
	return &Config{
		GRPCPort: getEnv("GRPC_PORT", "50051"),
		RadiusKM: getEnvAsFloat("RADIUS_KM", 5.0),
	}
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Printf("Env %s not found. Using default: %s", key, defaultVal)
		return defaultVal
	}
	return val
}

func getEnvAsFloat(key string, defaultVal float64) float64 {
	valStr := os.Getenv(key)
	if valStr == "" {
		log.Printf("Env %s not found. Using default: %f", key, defaultVal)
		return defaultVal
	}
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		log.Printf("Invalid float for %s: %v. Using default: %f", key, err, defaultVal)
		return defaultVal
	}
	return val
}
