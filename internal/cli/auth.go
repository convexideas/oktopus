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
	cmd.AddCommand(newAuthListCmd(app))
	cmd.AddCommand(newAuthRemoveCmd(app))
	cmd.AddCommand(newAuthStatusCmd(app))

	return cmd
}

func newAuthAddCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "add <provider>",
		Short: "Store an API key in the system keychain",
		Long:  "Providers: anthropic, openai, ollama",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]

			fmt.Fprintf(cmd.OutOrStdout(), "Paste your %s API key: ", provider)
			key, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(cmd.OutOrStdout()) // newline after hidden input
			if err != nil {
				return fmt.Errorf("reading key: %w", err)
			}

			if len(key) == 0 {
				return fmt.Errorf("empty key")
			}

			if err := credential.Store(provider, string(key)); err != nil {
				return fmt.Errorf("storing credential: %w", err)
			}

			cmd.Printf("✓ Stored in system keychain as ok:%s\n", provider)
			return nil
		},
	}
}

func newAuthListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List stored credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			providers := credential.List()
			if len(providers) == 0 {
				cmd.Println("no credentials stored (use: ok auth add <provider>)")
				return nil
			}
			for _, p := range providers {
				cmd.Printf("  ✓ %s\n", p)
			}
			return nil
		},
	}
}

func newAuthRemoveCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <provider>",
		Short: "Remove a stored credential",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			if err := credential.Delete(provider); err != nil {
				return fmt.Errorf("removing credential: %w", err)
			}
			cmd.Printf("✓ Removed ok:%s from keychain\n", provider)
			return nil
		},
	}
}

func newAuthStatusCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show credential resolution status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, p := range []string{"anthropic", "openai", "ollama"} {
				key, err := credential.Resolve(p)
				if err != nil {
					cmd.Printf("  ✗ %s — not configured\n", p)
				} else {
					// Show first 8 chars + masked
					masked := key[:min(8, len(key))] + "..."
					cmd.Printf("  ✓ %s — %s\n", p, masked)
				}
			}
			return nil
		},
	}
}
