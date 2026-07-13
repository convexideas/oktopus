package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const Version = "0.1.0"

// Execute runs the CLI.
func Execute(ctx context.Context) int {
	app, err := NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	defer app.Close()

	root := &cobra.Command{
		Use:   "ok",
		Short: "Oktopus — a governed operating system for agentic work",
	}

	root.AddCommand(
		newVersionCmd(),
		newRunCmd(app),
		newSessionsCmd(app),
		newCapabilitiesCmd(app),
		newWorkspaceCmd(app),
		newProfileCmd(app),
		newMemoryCmd(app),
	)

	root.SetContext(ctx)
	if err := root.Execute(); err != nil {
		return 1
	}
	return 0
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(Version)
		},
	}
}
