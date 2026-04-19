package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear stored tokens locally (does not call the server — use 'anonlogin grants revoke' to revoke server-side)",
		RunE: func(_ *cobra.Command, _ []string) error {
			deleteTokens()
			fmt.Println("Logged out. Stored tokens removed.")
			return nil
		},
	}
}
