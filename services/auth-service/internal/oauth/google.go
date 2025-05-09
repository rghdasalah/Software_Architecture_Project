package oauth

import (
	"auth-service/internal/config"
	"auth-service/internal/service"

	//"context"
	"encoding/json"
	//"fmt"
	"net/http"
	"net/url"
	"strings"
)

type GoogleUserResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	// You can expand this struct with more fields if needed
}

// ExchangeCodeForToken exchanges the authorization code for an access token
func ExchangeCodeForToken(code string) (string, error) {
	endpoint := "https://oauth2.googleapis.com/token"

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", config.AppConfig.GoogleClientID)
	data.Set("client_secret", config.AppConfig.GoogleClientSecret)
	data.Set("redirect_uri", config.AppConfig.GoogleRedirectURI)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", err
	}

	return res.AccessToken, nil
}

// GetGoogleUser fetches the user's profile info from Google
func GetGoogleUser(accessToken string) (*service.GoogleUser, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gRes GoogleUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&gRes); err != nil {
		return nil, err
	}

	return &service.GoogleUser{
		Name:  gRes.Name,
		Email: gRes.Email,
	}, nil
}

//Handles OAuth2 flow: exchange code for token, fetch profile
