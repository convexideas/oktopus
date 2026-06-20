package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

type Options struct {
	DBPath       string
	RegistryPath string
	RunsDir      string
	JSON         bool
	Verbose      bool
}

func NewRootCommand() *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "oktopus",
		Short: "Vendor-agnostic control plane for agent orchestration and governance",
		Long:  "Oktopus is a local-first, distributed-ready, vendor-agnostic control plane for agent orchestration and governance.",
	}

	cmd.PersistentFlags().StringVar(&opts.DBPath, "db", defaultDBPath(), "SQLite database path")
	cmd.PersistentFlags().StringVar(&opts.RegistryPath, "registry", defaultRegistryPath(), "capability registry path")
	cmd.PersistentFlags().StringVar(&opts.RunsDir, "runs-dir", defaultRunsDir(), "run artifacts directory")
	cmd.PersistentFlags().BoolVar(&opts.JSON, "json", false, "emit JSON output")
	cmd.PersistentFlags().BoolVar(&opts.Verbose, "verbose", false, "enable verbose output")

	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newDBCommand(opts))
	cmd.AddCommand(newRegistryCommand(opts))
	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print Oktopus version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version)
		},
	}
}

func defaultDBPath() string {
	if v := os.Getenv("OKTOPUS_DB"); v != "" {
		return v
	}
	return filepath.Join(".oktopus", "oktopus.db")
}

func defaultRegistryPath() string {
	if v := os.Getenv("OKTOPUS_REGISTRY"); v != "" {
		return v
	}
	return "registry"
}

func defaultRunsDir() string {
	if v := os.Getenv("OKTOPUS_RUNS_DIR"); v != "" {
		return v
	}
	return "runs"
}
