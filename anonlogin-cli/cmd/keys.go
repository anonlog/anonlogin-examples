package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func newKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage signing keys",
	}
	cmd.AddCommand(newKeysRotateCmd())
	return cmd
}

func newKeysRotateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rotate",
		Short: "Rotate the RSA signing key used for JWT tokens",
		Long: `Generates a new RSA-2048 signing key on the server.

New tokens will be signed with the new key immediately.
The old public key remains in the JWKS endpoint so existing
tokens can continue to be verified until they expire.

Run this command periodically (e.g. every 90 days) or after
a suspected key compromise.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			ts, err := loadTokens()
			if err != nil || ts.AccessToken == "" {
				return fmt.Errorf("not logged in; run 'anonlogin login' first")
			}

			req, _ := http.NewRequest("POST", cfg.IssuerURL+"/v1/keys/rotate", nil)
			req.Header.Set("Authorization", "Bearer "+ts.AccessToken)
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
				return fmt.Errorf("server error: %s — %s", errCode, stringField(result, "detail"))
			}
			fmt.Printf("Signing key rotated\n")
			fmt.Printf("  New key ID: %s\n", stringField(result, "new_key_id"))
			fmt.Printf("  Note:       %s\n", stringField(result, "message"))
			return nil
		},
	}
}
