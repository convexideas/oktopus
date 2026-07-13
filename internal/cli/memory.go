package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newMemoryCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "memory",
		Short:   "View captured memories",
		Aliases: []string{"mem"},
	}
	cmd.AddCommand(newMemoryListCmd(app))
	return cmd
}

func newMemoryListCmd(app *App) *cobra.Command {
	var workspaceFlag string
	var limitFlag int

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List recent episodes",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var wsID string
			if workspaceFlag != "" {
				ws, err := app.Store.GetWorkspaceByName(ctx, workspaceFlag)
				if err != nil {
					return fmt.Errorf("workspace %q not found", workspaceFlag)
				}
				wsID = ws.ID
			}

			var episodes []struct {
				ID         string
				SessionID  string
				Source     string
				Content    string
				CapturedAt string
			}

			if wsID != "" {
				eps, err := app.Store.ListEpisodesByWorkspace(ctx, wsID, limitFlag)
				if err != nil {
					return err
				}
				for _, ep := range eps {
					episodes = append(episodes, struct {
						ID, SessionID, Source, Content, CapturedAt string
					}{ep.ID, ep.SessionID, ep.Source, ep.Content, ep.CapturedAt})
				}
			} else {
				eps, err := app.Store.ListRecentEpisodes(ctx, limitFlag)
				if err != nil {
					return err
				}
				for _, ep := range eps {
					episodes = append(episodes, struct {
						ID, SessionID, Source, Content, CapturedAt string
					}{ep.ID, ep.SessionID, ep.Source, ep.Content, ep.CapturedAt})
				}
			}

			if len(episodes) == 0 {
				cmd.Println("no episodes captured yet")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tSESSION\tSOURCE\tCAPTURED\tPREVIEW")
			for _, ep := range episodes {
				preview := ep.Content
				if len(preview) > 60 {
					preview = preview[:57] + "..."
				}
				// Strip newlines from preview
				for i, c := range preview {
					if c == '\n' || c == '\r' {
						preview = preview[:i] + "..."
						break
					}
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", ep.ID[:8], ep.SessionID[:8], ep.Source, ep.CapturedAt, preview)
			}
			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVarP(&workspaceFlag, "workspace", "w", "", "Filter by workspace")
	cmd.Flags().IntVarP(&limitFlag, "limit", "n", 10, "Number of episodes to show")
	return cmd
}
