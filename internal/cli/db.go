package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/convexideas/oktopus/internal/db"
	"github.com/spf13/cobra"
)

func newDBCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Manage local Oktopus database",
	}
	cmd.AddCommand(newDBInitCommand(opts))
	cmd.AddCommand(newDBMigrateCommand(opts))
	return cmd
}

// db init: creates the database (directory + file) if absent and applies all
// migrations. Safe to run repeatedly; reports when already up to date.
func newDBInitCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create the SQLite database and apply all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := db.Open(opts.DBPath)
			if err != nil {
				return err
			}
			defer conn.Close()

			applied, err := db.Migrate(context.Background(), conn)
			if err != nil {
				return err
			}
			if len(applied) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "ok: database already initialized and up to date: %s\n", opts.DBPath)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ok: initialized %s and applied %d migration(s)\n", opts.DBPath, len(applied))
			for _, migration := range applied {
				fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", migration)
			}
			return nil
		},
	}
}

// db migrate: applies pending migrations to an already-initialized database.
// Errors if the database file does not exist, directing the user to run init.
func newDBMigrateCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Apply pending migrations to an existing database",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(opts.DBPath); os.IsNotExist(err) {
				return fmt.Errorf("database not found at %s: run 'oktopus db init' first", opts.DBPath)
			} else if err != nil {
				return fmt.Errorf("stat database %s: %w", opts.DBPath, err)
			}

			conn, err := db.Open(opts.DBPath)
			if err != nil {
				return err
			}
			defer conn.Close()

			applied, err := db.Migrate(context.Background(), conn)
			if err != nil {
				return err
			}
			if len(applied) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "ok: database up to date: %s\n", opts.DBPath)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ok: applied %d migration(s) to %s\n", len(applied), opts.DBPath)
			for _, migration := range applied {
				fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", migration)
			}
			return nil
		},
	}
}
