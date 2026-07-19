package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newSessionsCmd(app *App) *cobra.Command {
	var allFlag bool

	cmd := &cobra.Command{
		Use:     "sessions [workspace:sandbox]",
		Short:   "List sessions",
		Aliases: []string{"sess"},
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// ponytail: scoped listing requires store query by workspace/sandbox.
			// For now: list all (same as --all), but with sandbox field shown.
			// Upgrade: filter by workspace:sandbox when store supports it.
			_ = allFlag

			sessions, err := app.Sessions.ListSessions(cmd.Context(), 20)
			if err != nil {
				return err
			}
			if len(sessions) == 0 {
				cmd.Println("no sessions recorded yet")
				return nil
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tAGENT\tWORKSPACE\tSANDBOX\tSTATUS\tCREATED")
			for _, s := range sessions {
				ws := s.Workspace
				if ws == "" {
					ws = "-"
				}
				sb := s.Sandbox
				if sb == "" {
					sb = "-"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", s.ID[:8], s.Agent, ws, sb, s.Status, s.CreatedAt)
			}
			w.Flush()
			return nil
		},
	}

	cmd.Flags().BoolVar(&allFlag, "all", false, "Show sessions across all workspaces")
	return cmd
}
