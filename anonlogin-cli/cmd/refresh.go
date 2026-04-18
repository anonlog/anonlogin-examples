package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// refreshTokens exchanges a refresh token for new access and refresh tokens.
func refreshTokens(cfg *CLIConfig, refreshToken string) (*TokenStore, error) {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {cfg.ClientID},
	}
	resp, err := http.PostForm(cfg.IssuerURL+"/token", form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if errCode, ok := body["error"].(string); ok {
		desc := stringField(body, "error_description")
		if desc != "" {
			return nil, fmt.Errorf("%s: %s", errCode, desc)
		}
		return nil, fmt.Errorf("%s", errCode)
	}

	ts := &TokenStore{
		AccessToken:  stringField(body, "access_token"),
		RefreshToken: stringField(body, "refresh_token"),
		Scope:        stringField(body, "scope"),
	}
	if v, ok := body["expires_in"].(float64); ok {
		ts.ExpiresIn = int(v)
	}
	if ts.AccessToken == "" {
		return nil, fmt.Errorf("server returned no access token")
	}
	return ts, nil
}
