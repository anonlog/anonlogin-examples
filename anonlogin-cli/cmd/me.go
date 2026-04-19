package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

func newMeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "me",
		Short: "Show the identity and scopes of the current credential",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlogin login' first")
			}

			req, _ := http.NewRequest("GET", cfg.IssuerURL+"/v1/me", nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusUnauthorized {
				ts, err = refreshTokens(cfg, ts.RefreshToken)
				if err != nil {
					return fmt.Errorf("token refresh failed: %w", err)
				}
				if err := saveTokens(ts); err != nil {
					return err
				}
				req2, _ := http.NewRequest("GET", cfg.IssuerURL+"/v1/me", nil)
				req2.Header.Set("Authorization", "Bearer "+ts.AccessToken)
				resp, err = http.DefaultClient.Do(req2)
				if err != nil {
					return fmt.Errorf("request failed: %w", err)
				}
				defer resp.Body.Close()
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("decode error: %w", err)
			}
			if errMsg := stringField(result, "error"); errMsg != "" {
				return fmt.Errorf("server error: %s", errMsg)
			}

			fmt.Printf("Account:     %s\n", stringField(result, "account_id"))
			fmt.Printf("Auth method: %s\n", stringField(result, "auth_method"))
			if scopesRaw, ok := result["scopes"].([]interface{}); ok {
				strs := make([]string, len(scopesRaw))
				for i, s := range scopesRaw {
					strs[i], _ = s.(string)
				}
				if len(strs) > 0 {
					fmt.Printf("Scopes:      %s\n", strings.Join(strs, " "))
				} else {
					fmt.Printf("Scopes:      (none)\n")
				}
			}
			return nil
		},
	}
}
