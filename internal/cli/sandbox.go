package cli

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"text/tabwriter"

	"github.com/convexideas/oktopus/internal/runtime"
	"github.com/spf13/cobra"
)

func newSandboxCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sandbox",
		Short:   "Manage sandboxes",
		Aliases: []string{"sb"},
	}
	cmd.AddCommand(
		newSandboxCreateCmd(app),
		newSandboxExecCmd(app),
		newSandboxListCmd(app),
	)
	return cmd
}

func newSandboxCreateCmd(app *App) *cobra.Command {
	var personaFlag string

	cmd := &cobra.Command{
		Use:   "create <workspace:name>",
		Short: "Create a named sandbox",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, name, err := parseWorkspaceSandbox(args[0])
			if err != nil {
				return err
			}

			provider := runtime.NewLocalProvider()
			sb, err := provider.Create(ws, name)
			if err != nil {
				return err
			}

			// TODO: materialize persona + profile into sandbox
			_ = personaFlag

			cmd.Printf("created sandbox %s:%s (%s)\n", ws, name, sb.Home)
			return nil
		},
	}

	cmd.Flags().StringVarP(&personaFlag, "persona", "p", "", "Persona to configure the sandbox with")
	return cmd
}

func newSandboxExecCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "exec <workspace:name>",
		Short: "Enter a sandbox (shell with HOME redirected)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, name, err := parseWorkspaceSandbox(args[0])
			if err != nil {
				return err
			}

			provider := runtime.NewLocalProvider()
			sb, err := provider.Get(ws, name)
			if err != nil {
				return fmt.Errorf("sandbox %s:%s not found (create it with: ok sandbox create %s)", ws, name, args[0])
			}

			// Resolve workspace source for cwd
			ctx := cmd.Context()
			cwd, _ := os.Getwd()
			wsEntity, err := app.Store.GetWorkspaceByName(ctx, ws)
			if err == nil {
				refs, _ := app.Store.ListWorkspaceRefs(ctx, wsEntity.ID)
				for _, r := range refs {
					if r.Kind == "source" {
						cwd = r.Ref
						break
					}
				}
			}

			// Launch shell with HOME redirected
			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "/bin/sh"
			}

			app.Log.Info("entering sandbox", "sandbox", fmt.Sprintf("%s:%s", ws, name), "home", sb.Home)

			shellCmd := exec.Command(shell)
			shellCmd.Dir = cwd
			shellCmd.Stdin = os.Stdin
			shellCmd.Stdout = os.Stdout
			shellCmd.Stderr = os.Stderr
			shellCmd.Env = append(os.Environ(),
				"HOME="+sb.Home,
				"OK_SANDBOX="+ws+":"+name,
			)

			if err := shellCmd.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					os.Exit(exitErr.ExitCode())
				}
				return err
			}
			return nil
		},
	}
}

func newSandboxListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "list [workspace]",
		Short:   "List sandboxes",
		Aliases: []string{"ls"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var ws string
			if len(args) > 0 {
				ws = args[0]
			} else {
				// List all workspaces' sandboxes
				workspaces, _ := app.Store.ListWorkspaces(cmd.Context())
				if len(workspaces) == 0 {
					cmd.Println("no workspaces defined")
					return nil
				}

				provider := runtime.NewLocalProvider()
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
				fmt.Fprintln(w, "WORKSPACE\tSANDBOX")
				for _, workspace := range workspaces {
					names, _ := provider.List(workspace.Name)
					for _, name := range names {
						fmt.Fprintf(w, "%s\t%s\n", workspace.Name, name)
					}
				}
				w.Flush()
				return nil
			}

			provider := runtime.NewLocalProvider()
			names, _ := provider.List(ws)
			if len(names) == 0 {
				cmd.Printf("no sandboxes for workspace %q\n", ws)
				return nil
			}

			for _, name := range names {
				cmd.Printf("%s:%s\n", ws, name)
			}
			return nil
		},
	}
}

func parseWorkspaceSandbox(arg string) (workspace, name string, err error) {
	for i, c := range arg {
		if c == ':' {
			return arg[:i], arg[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("expected format workspace:sandbox (got %q)", arg)
}

// Ensure syscall is used
var _ = syscall.SIGTERM
