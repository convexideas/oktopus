package cli

import (
	"fmt"
	"os"

	"github.com/convexideas/oktopus/internal/store"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

type Options struct {
	DatabaseURL  string
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

	cmd.PersistentFlags().StringVar(&opts.DatabaseURL, "database-url", defaultDatabaseURL(), "state store URL (sqlite://path for local MVP)")
	cmd.PersistentFlags().StringVar(&opts.DatabaseURL, "db", defaultDatabaseURL(), "SQLite database path or URL (deprecated; use --database-url)")
	_ = cmd.PersistentFlags().MarkDeprecated("db", "use --database-url")
	cmd.PersistentFlags().StringVar(&opts.RegistryPath, "registry", defaultRegistryPath(), "capability registry path")
	cmd.PersistentFlags().StringVar(&opts.RunsDir, "runs-dir", defaultRunsDir(), "run artifacts directory")
	cmd.PersistentFlags().BoolVar(&opts.JSON, "json", false, "emit JSON output")
	cmd.PersistentFlags().BoolVar(&opts.Verbose, "verbose", false, "enable verbose output")

	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newDBCommand(opts))
	cmd.AddCommand(newProfileCommand(opts))
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

func defaultDatabaseURL() string {
	if v := os.Getenv("OKTOPUS_DATABASE_URL"); v != "" {
		return v
	}
	if v := os.Getenv("OKTOPUS_DB"); v != "" {
		return v
	}
	return store.DefaultURL()
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
