package config

import (
	"log"
	//"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Port               string
	DBUrl              string
	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string
}

var AppConfig Config

func LoadConfig() {
	// Load from .env
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	AppConfig = Config{
		Port:               viper.GetString("PORT"),
		DBUrl:              viper.GetString("DATABASE_URL"),
		JWTSecret:          viper.GetString("JWT_SECRET"),
		AccessTokenExpiry:  viper.GetDuration("ACCESS_TOKEN_EXPIRY"),
		RefreshTokenExpiry: viper.GetDuration("REFRESH_TOKEN_EXPIRY"),
		GoogleClientID:     viper.GetString("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: viper.GetString("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURI:  viper.GetString("GOOGLE_REDIRECT_URI"),
	}
}
