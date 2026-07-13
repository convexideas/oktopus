package cli

import (
	"fmt"
	"path/filepath"
	"text/tabwriter"

	"github.com/convexideas/oktopus/internal/execution"
	"github.com/spf13/cobra"
)

func newWorkspaceCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workspace",
		Short:   "Manage workspaces",
		Aliases: []string{"ws"},
	}
	cmd.AddCommand(
		newWorkspaceCreateCmd(app),
		newWorkspaceListCmd(app),
		newWorkspaceShowCmd(app),
	)
	return cmd
}

func newWorkspaceCreateCmd(app *App) *cobra.Command {
	var sourceFlag string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

			_, err := app.Store.GetWorkspaceByName(ctx, name)
			if err == nil {
				return fmt.Errorf("workspace %q already exists", name)
			}

			ws := &execution.Workspace{Name: name}
			if err := app.Store.CreateWorkspace(ctx, ws); err != nil {
				return err
			}

			if sourceFlag != "" {
				abs, _ := filepath.Abs(sourceFlag)
				ref := &execution.WorkspaceRef{
					WorkspaceID: ws.ID,
					Kind:        "source",
					Ref:         abs,
				}
				if err := app.Store.AddWorkspaceRef(ctx, ref); err != nil {
					return err
				}
			}

			cmd.Printf("created workspace %q (%s)\n", name, ws.ID[:8])
			return nil
		},
	}
	cmd.Flags().StringVar(&sourceFlag, "source", "", "Source path")
	return cmd
}

func newWorkspaceListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List all workspaces",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaces, err := app.Store.ListWorkspaces(cmd.Context())
			if err != nil {
				return err
			}
			if len(workspaces) == 0 {
				cmd.Println("no workspaces defined")
				return nil
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tCREATED")
			for _, ws := range workspaces {
				fmt.Fprintf(w, "%s\t%s\n", ws.Name, ws.CreatedAt)
			}
			w.Flush()
			return nil
		},
	}
}

func newWorkspaceShowCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show workspace details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			name := args[0]

			ws, err := app.Store.GetWorkspaceByName(ctx, name)
			if err != nil {
				return fmt.Errorf("workspace %q not found", name)
			}

			cmd.Printf("name:    %s\n", ws.Name)
			cmd.Printf("id:      %s\n", ws.ID)
			cmd.Printf("created: %s\n", ws.CreatedAt)

			refs, err := app.Store.ListWorkspaceRefs(ctx, ws.ID)
			if err != nil {
				return err
			}
			if len(refs) > 0 {
				cmd.Println()
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
				fmt.Fprintln(w, "KIND\tREF")
				for _, r := range refs {
					fmt.Fprintf(w, "%s\t%s\n", r.Kind, r.Ref)
				}
				w.Flush()
			}
			return nil
		},
	}
}
