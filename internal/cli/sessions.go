package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newSessionsCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "sessions",
		Short:   "List recent sessions",
		Aliases: []string{"sess"},
		RunE: func(cmd *cobra.Command, args []string) error {
			sessions, err := app.Store.ListSessions(cmd.Context(), 20)
			if err != nil {
				return err
			}
			if len(sessions) == 0 {
				cmd.Println("no sessions recorded yet")
				return nil
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tAGENT\tSTATUS\tCREATED")
			for _, s := range sessions {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.ID[:8], s.Agent, s.Status, s.CreatedAt)
			}
			w.Flush()
			return nil
		},
	}
}
