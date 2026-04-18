package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Revoke the current CLI session and clear stored tokens",
		RunE: func(_ *cobra.Command, _ []string) error {
			deleteTokens()
			fmt.Println("Logged out. Stored tokens removed.")
			return nil
		},
	}
}
