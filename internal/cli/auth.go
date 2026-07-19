package cli

import (
	"fmt"
	"os"

	"github.com/convexideas/oktopus/internal/credential"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newAuthCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage provider credentials",
	}

	cmd.AddCommand(newAuthAddCmd(app))
	cmd.AddCommand(newAuthRemoveCmd(app))
	cmd.AddCommand(newAuthStatusCmd(app))

	return cmd
}

func newAuthAddCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "add <credential-name>",
		Short: "Store a secret in the system keychain",
		Long:  "Credential name should match the 'credential' field in your config.yaml provider entries.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			fmt.Fprintf(cmd.OutOrStdout(), "Paste secret for %q: ", name)
			key, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(cmd.OutOrStdout())
			if err != nil {
				return fmt.Errorf("reading secret: %w", err)
			}

			if len(key) == 0 {
				return fmt.Errorf("empty secret")
			}

			if err := credential.Store(name, string(key)); err != nil {
				return fmt.Errorf("storing credential: %w", err)
			}

			cmd.Printf("✓ Stored in system keychain as ok:%s\n", name)
			return nil
		},
	}
}

func newAuthRemoveCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <credential-name>",
		Short: "Remove a stored credential",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := credential.Delete(name); err != nil {
				return fmt.Errorf("removing credential: %w", err)
			}
			cmd.Printf("✓ Removed ok:%s from keychain\n", name)
			return nil
		},
	}
}

func newAuthStatusCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show credential resolution status for configured providers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(app.Config.APIProviders) == 0 {
				cmd.Println("no api_providers configured in config.yaml")
				return nil
			}
			for name, p := range app.Config.APIProviders {
				if p.Credential == "" {
					cmd.Printf("  - %s — no credential configured\n", name)
					continue
				}
				if credential.Exists(p.Credential) {
					cmd.Printf("  ✓ %s (credential: %s)\n", name, p.Credential)
				} else {
					cmd.Printf("  ✗ %s (credential: %s — not found)\n", name, p.Credential)
				}
			}
			return nil
		},
	}
}
