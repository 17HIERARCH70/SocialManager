package authHandlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/17HIERARCH70/SocialManager/internal/lib/jwt"
	"github.com/17HIERARCH70/SocialManager/internal/services/authService"
	"golang.org/x/oauth2"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	authService *authService.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *authService.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// GoogleLoginHandler initiates Google OAuth login
// @Summary Google Login
// @Description Initiate Google OAuth login
// @Tags auth
// @Produce json
// @Success 200 {string} string "URL for Google login"
// @Router /auth/google_login [get]
func (h *AuthHandler) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := h.authService.OAuthConfig().AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallbackHandler handles Google OAuth callback
// @Summary Google OAuth Callback
// @Description Handle Google OAuth callback
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string "JWT Tokens"
// @Router /auth/google_callback [get]
func (h *AuthHandler) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state != "state-token" {
		http.Error(w, "State parameter doesn't match", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := h.authService.ExchangeToken(code)
	if err != nil {
		h.authService.Log.Error("Failed to exchange token", "error", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	userInfo, err := h.authService.FetchUserInfo(token)
	if err != nil {
		h.authService.Log.Error("Failed to fetch user info", "error", err)
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}

	userID, err := h.authService.SaveUserToDB(userInfo)
	if err != nil {
		h.authService.Log.Error("Failed to save user to DB", "error", err)
		http.Error(w, "Failed to save user to DB", http.StatusInternalServerError)
		return
	}

	err = h.authService.SaveGoogleTokensToDB(userID, token)
	if err != nil {
		h.authService.Log.Error("Failed to save Google tokens to DB", "error", err)
		http.Error(w, "Failed to save Google tokens to DB", http.StatusInternalServerError)
		return
	}

	authDetails := jwt.AuthDetails{
		AuthUuid: userInfo["sub"].(string),
		UserId:   uint64(userID),
	}

	jwtTokens, err := jwt.CreateTokenPair(authDetails)
	if err != nil {
		h.authService.Log.Error("Failed to create JWT tokens", "error", err)
		http.Error(w, "Failed to create JWT tokens", http.StatusInternalServerError)
		return
	}

	accessToken := jwtTokens["access_token"]
	refreshToken := jwtTokens["refresh_token"]
	expiresAt := time.Now().Add(15 * time.Minute)

	err = h.authService.SaveTokensToDB(userID, accessToken, refreshToken, expiresAt)
	if err != nil {
		h.authService.Log.Error("Failed to save JWT tokens to DB", "error", err)
		http.Error(w, "Failed to save JWT tokens to DB", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(jwtTokens)
	if err != nil {
		h.authService.Log.Error("Failed to encode response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
