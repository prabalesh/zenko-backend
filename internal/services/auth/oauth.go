package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthService interface {
	GetAuthURL(ctx context.Context) (string, error)
	VerifyState(ctx context.Context, state string) error
	GetGoogleUserInfo(ctx context.Context, code string) (*GoogleUser, error)
}

type oauthService struct {
	oauthConfig *oauth2.Config
	redisClient *redis.Client
}

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func NewOAuthService(clientID, clientSecret, redirectURL string, redisClient *redis.Client) OAuthService {
	return &oauthService{
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		redisClient: redisClient,
	}
}

func (s *oauthService) GetAuthURL(ctx context.Context) (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	state := base64.URLEncoding.EncodeToString(b)

	// Store state in Redis for 10 minutes
	err = s.redisClient.Set(ctx, "oauth_state:"+state, true, 10*time.Minute).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store oauth state: %w", err)
	}

	url := s.oauthConfig.AuthCodeURL(state)
	return url, nil
}

func (s *oauthService) VerifyState(ctx context.Context, state string) error {
	key := "oauth_state:" + state
	exists, err := s.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to verify oauth state: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("invalid or expired oauth state")
	}

	// Delete state after verification
	s.redisClient.Del(ctx, key)
	return nil
}

func (s *oauthService) GetGoogleUserInfo(ctx context.Context, code string) (*GoogleUser, error) {
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %w", err)
	}

	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info, status: %d", resp.StatusCode)
	}

	var user GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &user, nil
}
