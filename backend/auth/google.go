package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

// GoogleProvider implements Provider for Google OAuth2.
type GoogleProvider struct {
	config     oauth2.Config
	httpClient *http.Client
}

// NewGoogleProvider creates a Google OAuth2 provider.
func NewGoogleProvider(clientID, clientSecret, redirectURL string) *GoogleProvider {
	return &GoogleProvider{
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *GoogleProvider) AuthCodeURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *GoogleProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.httpClient)
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google token exchange: %w", err)
	}
	return token, nil
}

func (p *GoogleProvider) FetchUser(ctx context.Context, token *oauth2.Token) (UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return UserInfo{}, fmt.Errorf("creating google userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return UserInfo{}, fmt.Errorf("fetching google userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UserInfo{}, fmt.Errorf("google userinfo returned status %d", resp.StatusCode)
	}

	var profile struct {
		ID      json.Number `json:"id"`
		Email   string      `json:"email"`
		Name    string      `json:"name"`
		Picture string      `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return UserInfo{}, fmt.Errorf("decoding google userinfo: %w", err)
	}

	return UserInfo{
		Email:       profile.Email,
		DisplayName: profile.Name,
		AvatarURL:   profile.Picture,
		ProviderID:  profile.ID.String(),
	}, nil
}
