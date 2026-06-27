package cli

import (
	"encoding/json"
	"fmt"

	"github.com/convexideas/oktopus/internal/registry"
	"github.com/convexideas/oktopus/internal/store"
	"github.com/spf13/cobra"
)

func newRegistryCommand(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Inspect and validate capability registry",
	}
	cmd.AddCommand(newRegistryValidateCommand(opts))
	cmd.AddCommand(newRegistryListCommand(opts))
	cmd.AddCommand(newRegistryShowCommand(opts))
	return cmd
}

func newRegistryValidateCommand(opts *Options) *cobra.Command {
	var noIndex bool
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate local capability manifests and update the capability index",
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := registry.Load(opts.RegistryPath)
			if err != nil {
				return err
			}

			indexed := 0
			if !noIndex {
				st, err := store.Open(cmd.Context(), opts.DatabaseURL)
				if err != nil {
					return fmt.Errorf("open index store (use --no-index to validate files only): %w", err)
				}
				defer st.Close()
				if _, err := st.Migrate(cmd.Context()); err != nil {
					return fmt.Errorf("migrate index store: %w", err)
				}
				indexed, err = registry.SyncCapabilities(cmd.Context(), st.DB, reg)
				if err != nil {
					return err
				}
			}

			if opts.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"ok":           true,
					"capabilities": len(reg.Capabilities),
					"indexed":      indexed,
				})
			}
			if noIndex {
				fmt.Fprintf(cmd.OutOrStdout(), "ok: %d capabilities (files only, index skipped)\n", len(reg.Capabilities))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "ok: %d capabilities (%d indexed to %s)\n", len(reg.Capabilities), indexed, opts.DatabaseURL)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&noIndex, "no-index", false, "validate manifest files only; do not write the capability index")
	return cmd
}

func newRegistryListCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := registry.Load(opts.RegistryPath)
			if err != nil {
				return err
			}
			if opts.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(reg.Capabilities)
			}
			for _, cap := range reg.Capabilities {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", cap.ID(), cap.Path)
			}
			return nil
		},
	}
}

func newRegistryShowCommand(opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "show <kind:name[@version]|name>",
		Short: "Show capability manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := registry.Load(opts.RegistryPath)
			if err != nil {
				return err
			}
			cap, ok := reg.Find(args[0])
			if !ok {
				return fmt.Errorf("capability not found: %s", args[0])
			}
			body, err := cap.JSON()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(body))
			return nil
		},
	}
}
