package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

func newClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Manage OAuth2 clients",
	}
	cmd.AddCommand(
		newClientListCmd(),
		newClientCreateCmd(),
		newClientRotateSecretCmd(),
		newClientDeleteCmd(),
	)
	return cmd
}

// ── client list ───────────────────────────────────────────────────────────────

func newClientListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registered OAuth clients",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlogin login' first")
			}

			req, _ := http.NewRequest("GET", cfg.IssuerURL+"/v1/clients", nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var clients []map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
				return fmt.Errorf("decode error: %w", err)
			}
			if len(clients) == 0 {
				fmt.Println("No registered clients.")
				return nil
			}
			fmt.Printf("%-24s  %-20s  %-10s  %s\n", "CLIENT_ID", "NAME", "TYPE", "CREATED")
			fmt.Println(strings.Repeat("─", 80))
			for _, c := range clients {
				clientType := "confidential"
				if p, _ := c["is_public"].(bool); p {
					clientType = "public"
				}
				fmt.Printf("%-24s  %-20s  %-10s  %s\n",
					stringField(c, "client_id"),
					stringField(c, "name"),
					clientType,
					stringField(c, "created_at"),
				)
			}
			return nil
		},
	}
}

// ── client create ─────────────────────────────────────────────────────────────

func newClientCreateCmd() *cobra.Command {
	var (
		name             string
		redirectURIs     []string
		scopes           []string
		isPublic         bool
		subjectType      string
		sectorIdentifier string
		description      string
		homepageURL      string
		logoURL          string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Register a new OAuth client",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if len(redirectURIs) == 0 {
				return fmt.Errorf("--redirect-uri is required (can be specified multiple times)")
			}
		if len(scopes) == 0 {
			scopes = []string{"openid"}
		}
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlogin login' first")
			}

			payload := map[string]interface{}{
				"name":          name,
				"redirect_uris": redirectURIs,
				"scopes":        scopes,
				"is_public":     isPublic,
			}
			if subjectType != "" {
				payload["subject_type"] = subjectType
			}
			if sectorIdentifier != "" {
				payload["sector_identifier"] = sectorIdentifier
			}
			if description != "" {
				payload["description"] = description
			}
			if homepageURL != "" {
				payload["homepage_url"] = homepageURL
			}
			if logoURL != "" {
				payload["logo_url"] = logoURL
			}

			bodyBytes, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("encode error: %w", err)
			}

			req, _ := http.NewRequest("POST", cfg.IssuerURL+"/v1/clients",
				bytes.NewReader(bodyBytes))
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("decode error: %w", err)
			}
			if errCode := stringField(result, "error"); errCode != "" {
				return fmt.Errorf("server error: %s", errCode)
			}

			fmt.Printf("OAuth client registered\n")
			fmt.Printf("  Client ID: %s\n", stringField(result, "client_id"))
			fmt.Printf("  Name:      %s\n", name)
			if secret := stringField(result, "client_secret"); secret != "" {
				fmt.Printf("  Secret:    %s\n", secret)
				fmt.Printf("  Note:      %s\n", stringField(result, "note"))
			} else {
				fmt.Printf("  Type:      public (no secret — use PKCE)\n")
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "Application name (required)")
	cmd.Flags().StringArrayVarP(&redirectURIs, "redirect-uri", "r", nil, "Redirect URI (can repeat)")
	cmd.Flags().StringArrayVarP(&scopes, "scope", "s", nil, "Scope (can repeat; default: openid)")
	cmd.Flags().BoolVar(&isPublic, "public", false, "Register as a public client (PKCE only, no secret)")
	cmd.Flags().StringVar(&subjectType, "subject-type", "", "Subject type: public (default) or pairwise")
	cmd.Flags().StringVar(&sectorIdentifier, "sector-identifier", "", "Pairwise sector identifier hostname")
	cmd.Flags().StringVar(&description, "description", "", "Short description shown on the consent screen")
	cmd.Flags().StringVar(&homepageURL, "homepage-url", "", "Client homepage URL shown on the consent screen")
	cmd.Flags().StringVar(&logoURL, "logo-url", "", "Absolute URL of client logo for the consent screen")
	return cmd
}

// ── client rotate-secret ──────────────────────────────────────────────────────

func newClientRotateSecretCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rotate-secret <client-id>",
		Short: "Rotate the client secret",
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

			req, _ := http.NewRequest("POST", cfg.IssuerURL+"/v1/clients/"+clientID+"/rotate-secret", nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return fmt.Errorf("decode error: %w", err)
			}
			if errCode := stringField(result, "error"); errCode != "" {
				return fmt.Errorf("server error: %s", errCode)
			}
			fmt.Printf("Secret rotated for client %s\n", clientID)
			fmt.Printf("  New secret: %s\n", stringField(result, "client_secret"))
			fmt.Printf("  Note:       %s\n", stringField(result, "note"))
			return nil
		},
	}
}

// ── client delete ─────────────────────────────────────────────────────────────

func newClientDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Disable an OAuth client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlog login' first")
			}

			req, _ := http.NewRequest("DELETE", cfg.IssuerURL+"/v1/clients/"+id, nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNoContent {
				fmt.Printf("Client %s disabled.\n", id)
				return nil
			}
			var result map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&result)
			return fmt.Errorf("failed (status %d): %s", resp.StatusCode, stringField(result, "error"))
		},
	}
}
