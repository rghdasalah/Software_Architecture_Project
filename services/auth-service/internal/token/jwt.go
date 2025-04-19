package token

import (
	"auth-service/internal/config"
	"auth-service/internal/db"
	"time"

	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims defines the JWT payload structure
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a signed access token
func GenerateAccessToken(user *db.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.AppConfig.AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

// GenerateRefreshToken creates a signed refresh token
func GenerateRefreshToken(user *db.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.AppConfig.RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

// VerifyToken parses and validates a JWT string
func VerifyToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// 🔁 Helper to convert claims to db.User (for refresh flow)
func UserFromClaims(c Claims) *db.User {
	return &db.User{
		ID:    c.UserID,
		Email: c.Email,
		Role:  c.Role,
	}
}

//Generates and validates JWT tokens
