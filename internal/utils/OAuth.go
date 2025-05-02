package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"

	"github.com/Olegsandrik/Exponenta/config"
)

func convertToken(t *ExToken) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  t.AccessToken,
		TokenType:    t.TokenType,
		Expiry:       t.Expiry,
		RefreshToken: t.RefreshToken,
	}
}

type ExToken struct {
	AccessToken string `json:"access_token"`

	TokenType string `json:"token_type,omitempty"`

	RefreshToken string `json:"refresh_token,omitempty"`

	Expiry time.Time `json:"expires,omitempty"`

	ExpiresIn int `json:"expires_in,omitempty"`

	UserID int `json:"user_id,omitempty"`
}

func ExchangeToken(
	ctx context.Context, code string, deviceID string, state string, config *config.Config,
) (*oauth2.Token, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code_verifier", config.OauthCodeVer)
	data.Set("code", code)
	data.Set("client_id", config.OauthAppID)
	data.Set("device_id", deviceID)
	data.Set("redirect_uri", "http://localhost")
	data.Set("state", state)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://id.vk.com/oauth2/auth",
		bytes.NewBufferString(data.Encode()),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tokenResp ExToken
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	OauthToken := convertToken(&tokenResp)

	return OauthToken, nil
}
