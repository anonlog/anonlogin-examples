package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

func newGrantsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grants",
		Short: "Manage active OAuth refresh-token grants",
	}
	cmd.AddCommand(
		newGrantsListCmd(),
		newGrantsRevokeCmd(),
	)
	return cmd
}

// ── grants list ───────────────────────────────────────────────────────────────

func newGrantsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active OAuth grants (one per CLI/app login session)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlog login' first")
			}

			req, _ := http.NewRequest("GET", cfg.IssuerURL+"/v1/grants", nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			var grants []map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&grants); err != nil {
				return fmt.Errorf("decode error: %w", err)
			}
			if len(grants) == 0 {
				fmt.Println("No active grants.")
				return nil
			}
			fmt.Printf("%-26s  %-24s  %-30s  %s\n", "REQUEST_ID", "CLIENT_ID", "SCOPES", "CREATED_AT")
			fmt.Println(strings.Repeat("─", 100))
			for _, g := range grants {
				scopesRaw, _ := g["scopes"].([]interface{})
				scopeStrs := make([]string, len(scopesRaw))
				for i, s := range scopesRaw {
					scopeStrs[i], _ = s.(string)
				}
				fmt.Printf("%-26s  %-24s  %-30s  %s\n",
					stringField(g, "request_id"),
					stringField(g, "client_id"),
					strings.Join(scopeStrs, " "),
					stringField(g, "created_at"),
				)
			}
			return nil
		},
	}
}

// ── grants revoke ─────────────────────────────────────────────────────────────

func newGrantsRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <request_id>",
		Short: "Revoke a grant and its entire refresh-token family",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			requestID := args[0]
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlog login' first")
			}

			req, _ := http.NewRequest("DELETE", cfg.IssuerURL+"/v1/grants/"+requestID, nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNoContent {
				fmt.Printf("Grant %s revoked.\n", requestID)
				return nil
			}
			var result map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&result)
			return fmt.Errorf("revoke failed (status %d): %s", resp.StatusCode, stringField(result, "error"))
		},
	}
}
