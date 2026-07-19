package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/convexideas/oktopus/internal/credential"
	"github.com/convexideas/oktopus/internal/portal"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const portalCredentialKey = "portal-token"

func newLoginCmd(app *App) *cobra.Command {
	var tokenFlag string

	cmd := &cobra.Command{
		Use:   "login [server-url]",
		Short: "Authenticate with a remote portal",
		Long:  "Connect this CLI to a team/org portal for sync, collaboration, and policy.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			reader := bufio.NewReader(os.Stdin)

			serverURL := "https://portal.oktopus.dev"
			if len(args) > 0 {
				serverURL = args[0]
			}

			var auth *portal.AuthState
			var err error

			if tokenFlag != "" {
				// Token-based login (for CI, automation)
				auth, err = app.Portal.LoginWithToken(ctx, serverURL, tokenFlag)
			} else {
				// Interactive login
				cmd.Print("? Email: ")
				email := strings.TrimSpace(promptLine(reader))

				cmd.Print("? Password: ")
				pass, _ := term.ReadPassword(int(os.Stdin.Fd()))
				cmd.Println()

				auth, err = app.Portal.Login(ctx, serverURL, email, string(pass))
			}

			if err != nil {
				if err == portal.ErrOffline {
					return fmt.Errorf("portal not available — working offline")
				}
				return fmt.Errorf("login failed: %w", err)
			}

			// Store token in keychain for session persistence
			// ponytail: actual token comes from portal response. For now, store server URL as marker.
			credential.Store(portalCredentialKey, serverURL)

			cmd.Printf("✓ Logged in as %s (%s)\n", auth.Email, auth.OrgName)
			cmd.Printf("  server: %s\n", auth.ServerURL)
			return nil
		},
	}

	cmd.Flags().StringVar(&tokenFlag, "token", "", "API token (for CI/automation)")
	return cmd
}

func newLogoutCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Disconnect from the remote portal",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if err := app.Portal.Logout(ctx); err != nil && err != portal.ErrOffline {
				return err
			}
			credential.Delete(portalCredentialKey)
			cmd.Println("✓ Logged out")
			return nil
		},
	}
}

func newWhoAmICmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current portal authentication state",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			auth, err := app.Portal.WhoAmI(ctx)
			if err != nil {
				if err == portal.ErrOffline {
					cmd.Println("not connected — working offline")
					cmd.Println("  run 'ok login' to connect to a portal")
					return nil
				}
				return err
			}
			cmd.Printf("user:   %s\n", auth.Email)
			cmd.Printf("org:    %s\n", auth.OrgName)
			cmd.Printf("server: %s\n", auth.ServerURL)
			if !auth.ExpiresAt.IsZero() {
				cmd.Printf("expires: %s\n", auth.ExpiresAt.Format("2006-01-02 15:04"))
			}
			return nil
		},
	}
}

func newSyncCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync local state with remote portal",
	}
	cmd.AddCommand(
		newSyncStatusCmd(app),
		newSyncPushCmd(app),
		newSyncPullCmd(app),
	)
	return cmd
}

func newSyncStatusCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show what's pending between local and remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			status, err := app.Portal.Status(ctx)
			if err != nil {
				if err == portal.ErrOffline {
					cmd.Println("not connected — all data is local only")
					return nil
				}
				return err
			}
			cmd.Printf("sessions pending push: %d\n", status.SessionsPending)
			cmd.Printf("memory pending push:   %d\n", status.MemoryPending)
			cmd.Printf("remote pending pull:   %d\n", status.RemotePending)
			if !status.LastSyncedAt.IsZero() {
				cmd.Printf("last synced:           %s\n", status.LastSyncedAt.Format("2006-01-02 15:04"))
			}
			return nil
		},
	}
}

func newSyncPushCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Push local sessions and memory to portal",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if err := app.Portal.Push(ctx); err != nil {
				if err == portal.ErrOffline {
					return fmt.Errorf("not connected — run 'ok login' first")
				}
				return err
			}
			cmd.Println("✓ pushed")
			return nil
		},
	}
}

func newSyncPullCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: "Pull remote sessions, memory, and policy to local",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if err := app.Portal.Pull(ctx); err != nil {
				if err == portal.ErrOffline {
					return fmt.Errorf("not connected — run 'ok login' first")
				}
				return err
			}
			cmd.Println("✓ pulled")
			return nil
		},
	}
}

