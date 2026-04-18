package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

func newAppsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apps",
		Short: "Manage connected apps (OAuth consent grants)",
	}
	cmd.AddCommand(
		newAppsListCmd(),
		newAppsRevokeCmd(),
	)
	return cmd
}

// ── apps list ─────────────────────────────────────────────────────────────────

func newAppsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List apps you have approved",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlog login' first")
			}

			req, _ := http.NewRequest("GET", cfg.IssuerURL+"/v1/consent-grants", nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			var apps []map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
				return fmt.Errorf("decode error: %w", err)
			}
			if len(apps) == 0 {
				fmt.Println("No connected apps.")
				return nil
			}
			fmt.Printf("%-26s  %-22s  %-30s  %s\n", "CLIENT_ID", "NAME", "SCOPES", "GRANTED_AT")
			fmt.Println(strings.Repeat("─", 100))
			for _, a := range apps {
				scopesRaw, _ := a["scopes"].([]interface{})
				scopeStrs := make([]string, len(scopesRaw))
				for i, s := range scopesRaw {
					scopeStrs[i], _ = s.(string)
				}
				fmt.Printf("%-26s  %-22s  %-30s  %s\n",
					stringField(a, "client_id"),
					stringField(a, "client_name"),
					strings.Join(scopeStrs, " "),
					stringField(a, "granted_at"),
				)
			}
			return nil
		},
	}
}

// ── apps revoke ───────────────────────────────────────────────────────────────

func newAppsRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <client_id>",
		Short: "Revoke an app's consent grant; all its tokens are invalidated immediately",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientID := args[0]
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlog login' first")
			}

			req, _ := http.NewRequest("DELETE", cfg.IssuerURL+"/v1/consent-grants/"+clientID, nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNoContent {
				fmt.Printf("Access for %s revoked. All tokens issued to this app are now invalid.\n", clientID)
				return nil
			}
			var result map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&result)
			return fmt.Errorf("revoke failed (status %d): %s", resp.StatusCode, stringField(result, "error"))
		},
	}
}
