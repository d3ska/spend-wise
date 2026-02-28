package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	oauthgithub "golang.org/x/oauth2/github"
)

const (
	githubUserURL   = "https://api.github.com/user"
	githubEmailsURL = "https://api.github.com/user/emails"
)

// GitHubProvider implements Provider for GitHub OAuth2.
type GitHubProvider struct {
	config     oauth2.Config
	httpClient *http.Client
}

// NewGitHubProvider creates a GitHub OAuth2 provider.
func NewGitHubProvider(clientID, clientSecret, redirectURL string) *GitHubProvider {
	return &GitHubProvider{
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"user:email"},
			Endpoint:     oauthgithub.Endpoint,
		},
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *GitHubProvider) AuthCodeURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *GitHubProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.httpClient)
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("github token exchange: %w", err)
	}
	return token, nil
}

func (p *GitHubProvider) FetchUser(ctx context.Context, token *oauth2.Token) (UserInfo, error) {
	profile, err := p.fetchProfile(ctx, token)
	if err != nil {
		return UserInfo{}, err
	}

	// GitHub may not return a public email — fall back to /user/emails.
	if profile.Email == "" {
		email, err := p.fetchPrimaryEmail(ctx, token)
		if err != nil {
			return UserInfo{}, err
		}
		profile.Email = email
	}

	return UserInfo{
		Email:       profile.Email,
		DisplayName: profile.Name,
		AvatarURL:   profile.AvatarURL,
		ProviderID:  strconv.FormatInt(profile.ID, 10),
	}, nil
}

type githubProfile struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

func (p *GitHubProvider) fetchProfile(ctx context.Context, token *oauth2.Token) (githubProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubUserURL, nil)
	if err != nil {
		return githubProfile{}, fmt.Errorf("creating github user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return githubProfile{}, fmt.Errorf("fetching github user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return githubProfile{}, fmt.Errorf("github user API returned status %d", resp.StatusCode)
	}

	var profile githubProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return githubProfile{}, fmt.Errorf("decoding github user: %w", err)
	}
	return profile, nil
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (p *GitHubProvider) fetchPrimaryEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubEmailsURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating github emails request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching github emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github emails API returned status %d", resp.StatusCode)
	}

	var emails []githubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("decoding github emails: %w", err)
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	return "", fmt.Errorf("no primary verified email found on GitHub account")
}
