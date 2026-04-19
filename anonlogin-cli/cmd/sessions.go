package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newSessionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage active web sessions",
	}
	cmd.AddCommand(
		newSessionsListCmd(),
		newSessionsRevokeCmd(),
	)
	return cmd
}

// ── sessions list ─────────────────────────────────────────────────────────────

func newSessionsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active web sessions for your account",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlogin login' first")
			}

			req, _ := http.NewRequest("GET", cfg.IssuerURL+"/v1/sessions", nil)
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
				req2, _ := http.NewRequest("GET", cfg.IssuerURL+"/v1/sessions", nil)
				req2.Header.Set("Authorization", "Bearer "+ts.AccessToken)
				resp, err = http.DefaultClient.Do(req2)
				if err != nil {
					return fmt.Errorf("request failed: %w", err)
				}
				defer resp.Body.Close()
			}

			var sessions []map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
				return fmt.Errorf("decode error: %w", err)
			}
			if len(sessions) == 0 {
				fmt.Println("No active sessions.")
				return nil
			}

			fmt.Printf("%-26s  %-16s  %-16s  %-20s  %s\n", "ID", "AMR", "IP", "CREATED", "EXPIRES")
			fmt.Println(strings.Repeat("─", 95))
			for _, s := range sessions {
				id := stringField(s, "id")
				ip := stringField(s, "ip")
				created := formatTime(stringField(s, "created_at"))
				expires := formatTime(stringField(s, "expires_at"))

				amrRaw, _ := s["amr"].([]interface{})
				amrStrs := make([]string, len(amrRaw))
				for i, a := range amrRaw {
					amrStrs[i], _ = a.(string)
				}
				amr := strings.Join(amrStrs, "+")

				fmt.Printf("%-26s  %-16s  %-16s  %-20s  %s\n", id, amr, ip, created, expires)
			}
			return nil
		},
	}
}

// ── sessions revoke ───────────────────────────────────────────────────────────

func newSessionsRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <id>",
		Short: "Revoke a web session; that browser will be logged out on its next request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlogin login' first")
			}

			req, _ := http.NewRequest("DELETE", cfg.IssuerURL+"/v1/sessions/"+sessionID, nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNoContent {
				fmt.Printf("Session %s revoked.\n", sessionID)
				return nil
			}
			var result map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&result)
			return fmt.Errorf("revoke failed (status %d): %s", resp.StatusCode, stringField(result, "error"))
		},
	}
}

func formatTime(s string) string {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Local().Format("2006-01-02 15:04:05")
	}
	return s
}
