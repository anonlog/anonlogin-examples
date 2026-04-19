package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTokenCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "token",
		Short: "Manage the current access token",
	}

	c.AddCommand(
		&cobra.Command{
			Use:   "print",
			Short: "Print the current access token",
			RunE: func(_ *cobra.Command, _ []string) error {
				ts, err := loadTokens()
				if err != nil {
					return fmt.Errorf("not logged in: %w", err)
				}
				fmt.Println(ts.AccessToken)
				return nil
			},
		},
		&cobra.Command{
			Use:   "refresh",
			Short: "Force a token refresh using the stored refresh token",
			RunE: func(_ *cobra.Command, _ []string) error {
				ts, err := loadTokens()
				if err != nil {
					return fmt.Errorf("not logged in: %w", err)
				}
				if ts.RefreshToken == "" {
					return fmt.Errorf("no refresh token stored (re-login with: anonlogin login)")
				}
				cfg, err := loadConfig()
				if err != nil {
					return err
				}
				newTS, err := refreshTokens(cfg, ts.RefreshToken)
				if err != nil {
					return fmt.Errorf("token refresh failed: %w", err)
				}
				newTS.IssuerURL = cfg.IssuerURL
				if newTS.Scope == "" {
					newTS.Scope = ts.Scope
				}
				if err := saveTokens(newTS); err != nil {
					return fmt.Errorf("save tokens: %w", err)
				}
				fmt.Println("Token refreshed successfully.")
				return nil
			},
		},
	)
	return c
}
