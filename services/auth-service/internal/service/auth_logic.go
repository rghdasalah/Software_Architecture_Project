package service

import (
	"auth-service/internal/config"
	"auth-service/internal/db"
	"auth-service/internal/token"
	"database/sql"
	"errors"
	"net/http"
)

// GoogleUser represents user info returned from Google OAuth
type GoogleUser struct {
	Name        string
	Email       string
	Gender      string
	Batch       string
	PhoneNumber string
}

// AuthService wraps DB dependency
type AuthService struct {
	DB *sql.DB
}

// LoginOrRegister handles login or registration via Google OAuth
func (a *AuthService) LoginOrRegister(googleUser *GoogleUser, w http.ResponseWriter) (*db.User, string, error) {
	// Step 1: Check if user already exists
	existingUser, err := db.GetUserByEmail(a.DB, googleUser.Email)
	if err != nil {
		return nil, "", err
	}

	// Step 2: Register new user if not found
	if existingUser == nil {
		newUser := &db.User{
			Name:        googleUser.Name,
			Email:       googleUser.Email,
			Gender:      googleUser.Gender,
			Batch:       googleUser.Batch,
			PhoneNumber: googleUser.PhoneNumber,
			Role:        "user",
		}

		err := db.CreateUser(a.DB, newUser)
		if err != nil {
			return nil, "", err
		}

		existingUser = newUser
	}

	// Step 3: Generate Access Token
	accessToken, err := token.GenerateAccessToken(existingUser)
	if err != nil {
		return nil, "", errors.New("failed to generate access token")
	}

	// Step 4: Generate Refresh Token
	refreshToken, err := token.GenerateRefreshToken(existingUser)
	if err != nil {
		return nil, "", errors.New("failed to generate refresh token")
	}

	// Step 5: Set Refresh Token as HttpOnly Secure Cookie
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/v1/auth/refresh",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   int(config.AppConfig.RefreshTokenExpiry.Seconds()),
	}

	http.SetCookie(w, cookie)

	return existingUser, accessToken, nil
}

/**Core business logic like:

User exists? → log them in

New user? → create in DB

Generate JWT**/
