package authService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/17HIERARCH70/SocialManager/internal/config"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// AuthService handles authentication and token management
type AuthService struct {
	psql        *pgxpool.Pool
	oauthConfig *oauth2.Config
	Log         *slog.Logger
}

// NewAuthService creates a new AuthService
func NewAuthService(psql *pgxpool.Pool, cfg *config.Config, log *slog.Logger) (*AuthService, error) {
	oauthConfig, err := loadOAuthConfig(cfg.OAuth2.CredentialPath)
	if err != nil {
		return nil, err
	}

	return &AuthService{
		psql:        psql,
		oauthConfig: oauthConfig,
		Log:         log,
	}, nil
}

// loadOAuthConfig loads the OAuth configuration from a file
func loadOAuthConfig(credentialsFile string) (*oauth2.Config, error) {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, err
	}

	return google.ConfigFromJSON(b, gmail.GmailReadonlyScope, gmail.GmailModifyScope, "https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email")
}

// OAuthConfig returns the OAuth configuration
func (s *AuthService) OAuthConfig() *oauth2.Config {
	return s.oauthConfig
}

// ExchangeToken exchanges the authorization code for an access token
func (s *AuthService) ExchangeToken(code string) (*oauth2.Token, error) {
	token, err := s.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

	if !s.hasRequiredScopes(token) {
		return nil, fmt.Errorf("insufficient token scopes")
	}

	userID, err := s.GetUserIDByToken(token)
	if err != nil {
		return nil, err
	}

	if token.Expiry.Before(time.Now()) {
		token, err = s.RefreshToken(userID, token)
		if err != nil {
			return nil, err
		}
	}

	return token, nil
}

// hasRequiredScopes checks if the token has the required scopes
func (s *AuthService) hasRequiredScopes(token *oauth2.Token) bool {
	client := s.oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v1/tokeninfo?access_token=" + token.AccessToken)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var tokenInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return false
	}

	scopes, ok := tokenInfo["scope"].(string)
	if !ok {
		return false
	}

	requiredScopes := s.oauthConfig.Scopes
	for _, scope := range requiredScopes {
		if !strings.Contains(scopes, scope) {
			return false
		}
	}
	return true
}

// FetchUserInfo fetches the user info from the token
func (s *AuthService) FetchUserInfo(token *oauth2.Token) (map[string]interface{}, error) {
	client := s.oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}

// SaveUserToDB saves the user info to the database
func (s *AuthService) SaveUserToDB(userInfo map[string]interface{}) (int, error) {
	tx, err := s.psql.Begin(context.Background())
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(context.Background())

	query := `INSERT INTO users (google_id, email) VALUES ($1, $2)
              ON CONFLICT (google_id) DO NOTHING RETURNING id`
	var userID int
	err = tx.QueryRow(context.Background(), query, userInfo["sub"], userInfo["email"]).Scan(&userID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			var id int
			err = tx.QueryRow(context.Background(), "SELECT id FROM users WHERE google_id=$1", userInfo["sub"]).Scan(&id)
			if err != nil {
				return 0, err
			}
			userID = id
		} else {
			return 0, err
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		return 0, err
	}

	return userID, nil
}

// SaveGoogleTokensToDB saves the Google tokens to the database
func (s *AuthService) SaveGoogleTokensToDB(userID int, token *oauth2.Token) error {
	ctx := context.Background()
	query := `
        INSERT INTO google_tokens (user_id, access_token, refresh_token, expires_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id)
        DO UPDATE SET access_token = EXCLUDED.access_token, refresh_token = EXCLUDED.refresh_token, expires_at = EXCLUDED.expires_at`
	_, err := s.psql.Exec(ctx, query, userID, token.AccessToken, token.RefreshToken, token.Expiry)
	if err != nil {
		s.Log.Error("Failed to save Google tokens to DB", "error", err)
	}
	return err
}

// SaveTokensToDB saves the JWT tokens to the database
func (s *AuthService) SaveTokensToDB(userID int, accessToken, refreshToken string, expiresAt time.Time) error {
	ctx := context.Background()
	query := `INSERT INTO tokens (user_id, access_token, refresh_token, expires_at) 
	          VALUES ($1, $2, $3, $4)
	          ON CONFLICT (user_id) 
	          DO UPDATE SET access_token = EXCLUDED.access_token, refresh_token = EXCLUDED.refresh_token, expires_at = EXCLUDED.expires_at`
	_, err := s.psql.Exec(ctx, query, userID, accessToken, refreshToken, expiresAt)
	return err
}

// FetchGoogleTokenByUserID retrieves the Google token from the database by user ID
func (s *AuthService) FetchGoogleTokenByUserID(userID int) (*oauth2.Token, error) {
	row := s.psql.QueryRow(context.Background(), "SELECT access_token, refresh_token, expires_at FROM google_tokens WHERE user_id=$1", userID)

	var accessToken, refreshToken string
	var expiresAt time.Time
	err := row.Scan(&accessToken, &refreshToken, &expiresAt)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiresAt,
		TokenType:    "Bearer",
	}, nil
}

// RefreshToken refreshes the access token using the refresh token
func (s *AuthService) RefreshToken(userID int, token *oauth2.Token) (*oauth2.Token, error) {
	if token.RefreshToken == "" {
		s.Log.Warn("No refresh token available")
		return nil, errors.New("no refresh token available")
	}

	newToken, err := s.oauthConfig.TokenSource(context.Background(), token).Token()
	if err != nil {
		s.Log.Error("Failed to refresh token", "error", err)
		return nil, err
	}

	s.Log.Info("New token obtained", slog.String("access_token", newToken.AccessToken), slog.String("refresh_token", newToken.RefreshToken), slog.Time("expires_at", newToken.Expiry))

	err = s.SaveGoogleTokensToDB(userID, newToken)
	if err != nil {
		s.Log.Error("Failed to save new tokens to DB", "error", err)
		return nil, err
	}

	return newToken, nil
}

// GetUserIDByToken retrieves the user ID associated with the token
func (s *AuthService) GetUserIDByToken(token *oauth2.Token) (int, error) {
	userInfo, err := s.FetchUserInfo(token)
	if err != nil {
		return 0, err
	}
	return s.SaveUserToDB(userInfo)
}
