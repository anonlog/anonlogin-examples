package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Sign in to anonlog.in via the device authorization flow (RFC 8628)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			return loginDevice(cfg)
		},
	}
}

// loginDevice implements the RFC 8628 device authorization flow.
func loginDevice(cfg *CLIConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// Step 1: Request device and user codes.
	form := url.Values{
		"client_id": {cfg.ClientID},
		"scope":     {"openid offline_access api:read api:write"},
	}
	resp, err := http.PostForm(cfg.IssuerURL+"/device/code", form)
	if err != nil {
		return fmt.Errorf("device code request: %w", err)
	}
	defer resp.Body.Close()

	var dc struct {
		DeviceCode              string `json:"device_code"`
		UserCode                string `json:"user_code"`
		VerificationURI         string `json:"verification_uri"`
		VerificationURIComplete string `json:"verification_uri_complete"`
		ExpiresIn               int    `json:"expires_in"`
		Interval                int    `json:"interval"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dc); err != nil {
		return fmt.Errorf("decode device code response: %w", err)
	}
	if dc.DeviceCode == "" {
		return fmt.Errorf("server returned no device code")
	}

	interval := dc.Interval
	if interval <= 0 {
		interval = 5
	}

	fmt.Printf("\n  Activate your device:\n\n")
	fmt.Printf("  URL:  %s\n", dc.VerificationURI)
	fmt.Printf("  Code: %s\n\n", dc.UserCode)
	fmt.Printf("  Or go directly to:\n  %s\n\n", dc.VerificationURIComplete)
	fmt.Printf("  Waiting for activation (Ctrl-C to cancel)...\n\n")

	// Step 2: Poll /device/token until approved or expired.
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("login timed out")
		case <-ticker.C:
			ts, err := pollDeviceToken(cfg, dc.DeviceCode)
			if err != nil {
				if strings.Contains(err.Error(), "authorization_pending") {
					continue
				}
				return err
			}
			ts.IssuerURL = cfg.IssuerURL
			if err := saveTokens(ts); err != nil {
				return fmt.Errorf("save tokens: %w", err)
			}
			fmt.Println("  ✓ Authenticated successfully.")
			fmt.Printf("  Scope: %s\n", ts.Scope)
			return nil
		}
	}
}

func pollDeviceToken(cfg *CLIConfig, deviceCode string) (*TokenStore, error) {
	form := url.Values{
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		"device_code": {deviceCode},
		"client_id":   {cfg.ClientID},
	}
	resp, err := http.PostForm(cfg.IssuerURL+"/device/token", form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if errCode, ok := body["error"].(string); ok {
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
	return ts, nil
}

func stringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
