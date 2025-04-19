package transport

import (
	"auth-service/internal/oauth" //added for google
	"auth-service/internal/service"
	"auth-service/internal/token"
	"encoding/json"

	//"log"
	"net/http"
)

// AuthHandler wraps the AuthService
type AuthHandler struct {
	Service *service.AuthService
}

// GET /api/v1/auth/google/callback
func (h *AuthHandler) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	accessToken, err := oauth.ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Token exchange failed", http.StatusInternalServerError)
		return
	}

	googleUser, err := oauth.GetGoogleUser(accessToken)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	user, jwtToken, err := h.Service.LoginOrRegister(googleUser, w)
	if err != nil {
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": jwtToken,
		"user":         user,
	})
}

// GET /api/v1/auth/refresh
func (h *AuthHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Missing refresh token", http.StatusUnauthorized)
		return
	}

	claims, err := token.VerifyToken(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Generate new access token
	accessToken, err := token.GenerateAccessToken(token.UserFromClaims(*claims))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"access_token": accessToken})
}

// GET /api/v1/auth/profile
func (h *AuthHandler) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": claims.UserID,
		"email":   claims.Email,
		"role":    claims.Role,
	})
}

/**HTTP Handlers:

/auth/google → Redirect to Google

/auth/google/callback → Process code

/auth/verify → Optional JWT check route**/
