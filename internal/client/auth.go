package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// authProvider manages authentication tokens.
type authProvider struct {
	mu sync.RWMutex

	// API token auth (preferred)
	apiToken string

	// OAuth2 auth
	username     string
	password     string
	clientID     string
	clientSecret string

	// Token state
	accessToken  string
	refreshToken string
	expiresAt    time.Time

	endpoint string
	client   *http.Client
}

func newAuthProvider(cfg *Config) *authProvider {
	return &authProvider{
		apiToken:     cfg.APIToken,
		username:     cfg.Username,
		password:     cfg.Password,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		endpoint:     strings.TrimRight(cfg.APIEndpoint, "/"),
		client:       &http.Client{Timeout: 30 * time.Second},
	}
}

// getAuthHeader returns the Authorization header value.
func (a *authProvider) getAuthHeader() (string, error) {
	if a.apiToken != "" {
		return "Bearer " + a.apiToken, nil
	}
	return a.getOAuthToken()
}

// getOAuthToken returns a valid OAuth2 access token, refreshing if needed.
func (a *authProvider) getOAuthToken() (string, error) {
	a.mu.RLock()
	if a.accessToken != "" && time.Now().Before(a.expiresAt.Add(-30*time.Second)) {
		token := "Bearer " + a.accessToken
		a.mu.RUnlock()
		return token, nil
	}
	a.mu.RUnlock()

	a.mu.Lock()
	defer a.mu.Unlock()

	// Double-check after acquiring write lock
	if a.accessToken != "" && time.Now().Before(a.expiresAt.Add(-30*time.Second)) {
		return "Bearer " + a.accessToken, nil
	}

	// Try refresh token first
	if a.refreshToken != "" {
		if err := a.doRefresh(); err == nil {
			return "Bearer " + a.accessToken, nil
		}
	}

	// Fall back to password grant
	return a.doPasswordGrant()
}

func (a *authProvider) doPasswordGrant() (string, error) {
	data := url.Values{
		"grant_type":    {"password"},
		"username":      {a.username},
		"password":      {a.password},
		"client_id":     {a.clientID},
		"client_secret": {a.clientSecret},
	}

	resp, err := a.client.PostForm(a.endpoint+"/api/oauth", data)
	if err != nil {
		return "", fmt.Errorf("oauth2 token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading oauth2 response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("oauth2 token request returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp OAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing oauth2 response: %w", err)
	}

	a.accessToken = tokenResp.AccessToken
	a.refreshToken = tokenResp.RefreshToken
	a.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return "Bearer " + a.accessToken, nil
}

func (a *authProvider) doRefresh() error {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {a.refreshToken},
		"client_id":     {a.clientID},
		"client_secret": {a.clientSecret},
	}

	resp, err := a.client.PostForm(a.endpoint+"/api/oauth", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh token request returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var tokenResp OAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return err
	}

	a.accessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		a.refreshToken = tokenResp.RefreshToken
	}
	a.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return nil
}
