package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/convexideas/oktopus/internal/profile"
	"github.com/convexideas/oktopus/internal/store"
	"github.com/spf13/cobra"
)

func newProfileCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage Oktopus profiles",
	}
	cmd.AddCommand(newProfileInitCommand(opts))
	cmd.AddCommand(newProfileListCommand(opts))
	return cmd
}

func newProfileInitCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init local",
		Short: "Create or update the local profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "local" {
				return fmt.Errorf("unsupported profile %q; only local is implemented", args[0])
			}
			st, err := store.Open(context.Background(), opts.DatabaseURL)
			if err != nil {
				return err
			}
			defer st.Close()
			if _, err := st.Migrate(cmd.Context()); err != nil {
				return err
			}

			p, err := profile.InitLocal(cmd.Context(), st.DB, profile.LocalOptions{})
			if err != nil {
				return err
			}
			if opts.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(p)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ok: profile %s\n", p.Name)
			return nil
		},
	}
	return cmd
}

func newProfileListCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := store.Open(context.Background(), opts.DatabaseURL)
			if err != nil {
				return err
			}
			defer st.Close()
			if _, err := st.Migrate(cmd.Context()); err != nil {
				return err
			}

			profiles, err := profile.List(cmd.Context(), st.DB)
			if err != nil {
				return err
			}
			if opts.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(profiles)
			}
			for _, p := range profiles {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", p.Name)
			}
			return nil
		},
	}
}
