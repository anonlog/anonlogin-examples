package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose the anonlog.in server (discovery, JWKS, token, clock skew)",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			return runDoctor(cfg)
		},
	}
}

func runDoctor(cfg *CLIConfig) error {
	issuer := strings.TrimRight(cfg.IssuerURL, "/")
	fmt.Printf("Checking anonlog.in at %s\n\n", issuer)

	var allOK = true
	check := func(name string, fn func() error) {
		fmt.Printf("  %-40s", name+"...")
		if err := fn(); err != nil {
			fmt.Printf("FAIL (%v)\n", err)
			allOK = false
		} else {
			fmt.Println("OK")
		}
	}

	var discoveryDoc map[string]interface{}

	check("Discovery metadata", func() error {
		resp, err := httpGet(issuer + "/.well-known/openid-configuration")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return json.NewDecoder(resp.Body).Decode(&discoveryDoc)
	})

	check("JWKS endpoint", func() error {
		jwksURI, _ := discoveryDoc["jwks_uri"].(string)
		if jwksURI == "" {
			jwksURI = issuer + "/jwks.json"
		}
		resp, err := httpGet(jwksURI)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		var jwks map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
			return err
		}
		keys, _ := jwks["keys"].([]interface{})
		if len(keys) == 0 {
			return fmt.Errorf("no keys in JWKS")
		}
		return nil
	})

	check("Stored token validity", func() error {
		ts, err := loadTokens()
		if err != nil {
			return fmt.Errorf("not logged in")
		}
		claims, err := parseJWTClaims(ts.AccessToken)
		if err != nil {
			return fmt.Errorf("invalid access token: %w", err)
		}
		exp, ok := claims["exp"].(float64)
		if !ok {
			return fmt.Errorf("no exp claim in token")
		}
		expTime := time.Unix(int64(exp), 0)
		if time.Now().After(expTime) {
			return fmt.Errorf("access token expired at %s", expTime.Format(time.RFC3339))
		}
		remaining := time.Until(expTime).Round(time.Second)
		fmt.Printf("OK (expires in %s)", remaining)
		return nil
	})

	check("Clock skew", func() error {
		before := time.Now()
		resp, err := httpGet(issuer + "/jwks.json")
		if err != nil {
			return err
		}
		after := time.Now()
		resp.Body.Close()

		dateHeader := resp.Header.Get("Date")
		if dateHeader == "" {
			return nil // Can't check without Date header
		}
		serverTime, err := http.ParseTime(dateHeader)
		if err != nil {
			return err
		}
		clientTime := before.Add(after.Sub(before) / 2) // midpoint
		skew := clientTime.Sub(serverTime)
		if skew < 0 {
			skew = -skew
		}
		if skew > 5*time.Minute {
			return fmt.Errorf("clock skew too large: %s (TOTP will fail if > 30s)", skew)
		}
		fmt.Printf("OK (skew: %s)", skew.Round(time.Millisecond))
		return nil
	})

	fmt.Println()
	if allOK {
		fmt.Println("All checks passed.")
	} else {
		fmt.Println("Some checks failed. Run 'anonlog login' if not authenticated.")
	}
	return nil
}

func httpGet(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	return client.Get(url)
}
