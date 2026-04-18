package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Print current identity (issuer, subject, scopes, expiry)",
		RunE: func(_ *cobra.Command, _ []string) error {
			ts, err := loadTokens()
			if err != nil {
				return fmt.Errorf("not logged in: %w", err)
			}

			claims, err := parseJWTClaims(ts.AccessToken)
			if err != nil {
				return fmt.Errorf("invalid access token: %w", err)
			}

			fmt.Printf("Issuer:  %s\n", stringField(claims, "iss"))
			fmt.Printf("Subject: %s\n", stringField(claims, "sub"))
			fmt.Printf("Scope:   %s\n", ts.Scope)

			if exp, ok := claims["exp"].(float64); ok {
				t := time.Unix(int64(exp), 0)
				fmt.Printf("Expires: %s", t.Format(time.RFC3339))
				if time.Now().After(t) {
					fmt.Print(" (EXPIRED)")
				}
				fmt.Println()
			}
			return nil
		},
	}
}

// parseJWTClaims decodes the payload of a JWT without verifying its signature.
// Verification is handled server-side; the CLI only needs to read the claims.
func parseJWTClaims(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("not a JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}
	return claims, nil
}
