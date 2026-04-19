package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

func newInviteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invite",
		Short: "Manage registration invite codes (invite admins only)",
	}
	cmd.AddCommand(
		newInviteCreateCmd(),
		newInviteListCmd(),
		newInviteDeleteCmd(),
	)
	return cmd
}

// ── invite create ─────────────────────────────────────────────────────────────

func newInviteCreateCmd() *cobra.Command {
	var (
		note         string
		expiresInDays int
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new invite code",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlogin login' first")
			}

			payload := map[string]interface{}{"note": note}
			if expiresInDays > 0 {
				payload["expires_in_days"] = expiresInDays
			}
			bodyBytes, _ := json.Marshal(payload)

			req, _ := http.NewRequest("POST", cfg.IssuerURL+"/v1/invites",
				strings.NewReader(string(bodyBytes)))
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)
			req.Header.Set("Content-Type", "application/json")

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

			fmt.Printf("Invite code created\n")
			fmt.Printf("  Code: %s\n", stringField(result, "code"))
			if url := stringField(result, "registration_url"); url != "" {
				fmt.Printf("  URL:  %s\n", url)
			}
			if note != "" {
				fmt.Printf("  Note: %s\n", note)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&note, "note", "m", "", "Optional note describing who this code is for")
	cmd.Flags().IntVarP(&expiresInDays, "expires-in-days", "e", 0, "Expire after N days (0 = no expiry)")
	return cmd
}

// ── invite list ───────────────────────────────────────────────────────────────

func newInviteListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all invite codes",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlogin login' first")
			}

			req, _ := http.NewRequest("GET", cfg.IssuerURL+"/v1/invites", nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			var invites []map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&invites); err != nil {
				return fmt.Errorf("decode error: %w", err)
			}
			if len(invites) == 0 {
				fmt.Println("No invite codes.")
				return nil
			}
			fmt.Printf("%-34s  %-20s  %-22s  %s\n", "CODE", "NOTE", "EXPIRES", "USED_BY")
			fmt.Println(strings.Repeat("─", 95))
			for _, inv := range invites {
				code := stringField(inv, "code")
				note := stringField(inv, "note")
				expires := stringField(inv, "expires_at")
				if expires == "" {
					expires = "never"
				}
				usedBy := stringField(inv, "used_by")
				if usedBy == "" {
					usedBy = "—"
				}
				fmt.Printf("%-34s  %-20s  %-22s  %s\n", code, note, expires, usedBy)
			}
			return nil
		},
	}
}

// ── invite delete ─────────────────────────────────────────────────────────────

func newInviteDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <code>",
		Short: "Hard-delete an unused invite code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			code := args[0]
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlogin login' first")
			}

			req, _ := http.NewRequest("DELETE", cfg.IssuerURL+"/v1/invites/"+code, nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNoContent {
				fmt.Printf("Invite code %s deleted.\n", code)
				return nil
			}
			var result map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&result)
			return fmt.Errorf("delete failed (status %d): %s", resp.StatusCode, stringField(result, "error"))
		},
	}
}
