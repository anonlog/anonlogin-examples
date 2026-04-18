// Package cmd defines the anonlogin CLI command tree.
package cmd

import "github.com/spf13/cobra"

// Root returns the top-level cobra.Command for the anonlogin CLI.
func Root() *cobra.Command {
	root := &cobra.Command{
		Use:   "anonlogin",
		Short: "anonlog.in identity CLI",
		Long: `anonlogin – command-line client for anonlog.in

Authenticate, manage tokens, API keys, and OAuth clients from your terminal.`,
		SilenceUsage: true,
	}

	root.AddCommand(
		newLoginCmd(),
		newLogoutCmd(),
		newWhoamiCmd(),
		newTokenCmd(),
		newDoctorCmd(),
		newConfigCmd(),
		newAPIKeyCmd(),
		newAuthLogCmd(),
		newClientCmd(),
		newKeysCmd(),
		newInviteCmd(),
		newAppsCmd(),
		newGrantsCmd(),
	)
	return root
}
