package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newAPIKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-key",
		Short: "Manage personal API keys",
	}
	cmd.AddCommand(
		newAPIKeyListCmd(),
		newAPIKeyCreateCmd(),
		newAPIKeyRevokeCmd(),
	)
	return cmd
}

// ── api-key list ──────────────────────────────────────────────────────────────

func newAPIKeyListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlogin login' first")
			}

			req, _ := http.NewRequest("GET", cfg.IssuerURL+"/v1/api-keys", nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusUnauthorized {
				// Try to refresh first.
				ts, err = refreshTokens(cfg, ts.RefreshToken)
				if err != nil {
					return fmt.Errorf("token refresh failed: %w", err)
				}
				if err := saveTokens(ts); err != nil {
					return err
				}
				req2, _ := http.NewRequest("GET", cfg.IssuerURL+"/v1/api-keys", nil)
				req2.Header.Set("Authorization", "Bearer "+ts.AccessToken)
				resp, err = http.DefaultClient.Do(req2)
				if err != nil {
					return fmt.Errorf("request failed: %w", err)
				}
				defer resp.Body.Close()
			}

			var keys []map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
				return fmt.Errorf("decode error: %w", err)
			}
			if len(keys) == 0 {
				fmt.Println("No active API keys.")
				return nil
			}
			fmt.Printf("%-26s  %-14s  %-20s  %-24s  %-20s  %s\n", "ID", "PREFIX", "NAME", "SCOPES", "CREATED", "LAST USED")
			fmt.Println(strings.Repeat("─", 115))
			for _, k := range keys {
				id := stringField(k, "id")
				prefix := stringField(k, "prefix")
				name := stringField(k, "name")
				created := stringField(k, "created_at")
				lastUsed := stringField(k, "last_used_at")
				if lastUsed == "" {
					lastUsed = "—"
				}
				scopesRaw, _ := k["scopes"].([]interface{})
				scopeStrs := make([]string, len(scopesRaw))
				for i, s := range scopesRaw {
					scopeStrs[i], _ = s.(string)
				}
				scopes := strings.Join(scopeStrs, " ")
				fmt.Printf("%-26s  %-14s  %-20s  %-24s  %-20s  %s\n", id, prefix, name, scopes, created, lastUsed)
			}
			return nil
		},
	}
}

// ── api-key create ────────────────────────────────────────────────────────────

func newAPIKeyCreateCmd() *cobra.Command {
	var name, scope string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlogin login' first")
			}

			body := url.Values{
				"name":  {name},
				"scope": {scope},
			}
			req, _ := http.NewRequest("POST", cfg.IssuerURL+"/v1/api-keys",
				strings.NewReader(body.Encode()))
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("decode error: %w", err)
			}
			if errCode := stringField(result, "error"); errCode != "" {
				return fmt.Errorf("server error: %s", errCode)
			}

			key := stringField(result, "key")
			id := stringField(result, "id")
			fmt.Printf("API key created\n")
			fmt.Printf("  ID:     %s\n", id)
			fmt.Printf("  Name:   %s\n", name)
			fmt.Printf("  Key:    %s\n", key)
			fmt.Printf("  Note:   %s\n", stringField(result, "note"))
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Key name (required)")
	cmd.Flags().StringVarP(&scope, "scope", "s", "api:read", `Space-separated scopes (e.g. "api:read api:write")`)
	return cmd
}

// ── api-key revoke ────────────────────────────────────────────────────────────

func newAPIKeyRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <id>",
		Short: "Revoke an API key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyID := args[0]
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlog login' first")
			}

			req, _ := http.NewRequest("DELETE", cfg.IssuerURL+"/v1/api-keys/"+keyID, nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNoContent {
				fmt.Printf("API key %s revoked.\n", keyID)
				return nil
			}
			var result map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&result)
			return fmt.Errorf("revoke failed (status %d): %s", resp.StatusCode, stringField(result, "error"))
		},
	}
}

// ── auth-log ──────────────────────────────────────────────────────────────────

func newAuthLogCmd() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "auth-log",
		Short: "Show authentication audit log",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlog login' first")
			}

			endpoint := fmt.Sprintf("%s/v1/auth-events?limit=%d", cfg.IssuerURL, limit)
			req, _ := http.NewRequest("GET", endpoint, nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			var events []map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
				return fmt.Errorf("decode error: %w", err)
			}
			if len(events) == 0 {
				fmt.Println("No auth events found.")
				return nil
			}

			fmt.Printf("%-20s  %-6s  %-18s  %-16s  %s\n", "TIME", "RESULT", "METHOD", "IP", "CLIENT")
			fmt.Println(strings.Repeat("─", 85))
			for _, e := range events {
				ts := stringField(e, "created_at")
				if t, err := time.Parse(time.RFC3339, ts); err == nil {
					ts = t.Local().Format("2006-01-02 15:04:05")
				}
				result := "ok"
				if ok, _ := e["success"].(bool); !ok {
					result = "fail"
				}
				client := stringField(e, "client_id")
				if client == "" {
					client = "—"
				}
				fmt.Printf("%-20s  %-6s  %-18s  %-16s  %s\n",
					ts,
					result,
					stringField(e, "auth_method"),
					stringField(e, "ip"),
					client,
				)
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Number of events to show (max 500)")
	return cmd
}
