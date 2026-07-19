package cli

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"text/tabwriter"

	"github.com/convexideas/oktopus/internal/registry"
	"github.com/spf13/cobra"
)

func newCapabilitiesCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "capabilities",
		Short:   "Manage capabilities registry",
		Aliases: []string{"caps"},
	}
	cmd.AddCommand(
		newCapabilitiesListCmd(app),
		newCapabilitiesAddCmd(app),
	)
	return cmd
}

func newCapabilitiesListCmd(app *App) *cobra.Command {
	var kindFlag string

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List registered capabilities",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var caps []registry.Capability
			var err error
			if kindFlag != "" {
				caps, err = app.Capabilities.ListCapabilitiesByKind(cmd.Context(), kindFlag)
			} else {
				caps, err = app.Capabilities.ListCapabilities(cmd.Context())
			}
			if err != nil {
				return err
			}
			if len(caps) == 0 {
				cmd.Println("no capabilities registered")
				return nil
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
			fmt.Fprintln(w, "KIND\tNAME\tVERSION\tDESCRIPTION")
			for _, c := range caps {
				desc := c.Description
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Kind, c.Name, c.Version, desc)
			}
			w.Flush()
			return nil
		},
	}
	cmd.Flags().StringVar(&kindFlag, "kind", "", "Filter by capability kind")
	return cmd
}

func newCapabilitiesAddCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "add <manifest.yaml>",
		Short: "Register a capability from a manifest file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			path := args[0]

			meta, spec, err := registry.LoadPersonaFile(path)
			if err != nil {
				return err
			}

			specJSON, err := json.Marshal(spec)
			if err != nil {
				return fmt.Errorf("encoding manifest: %w", err)
			}

			_, err = app.Capabilities.FindCapability(ctx, meta.Kind, meta.Name, meta.Version)
			if err == nil {
				return fmt.Errorf("%s/%s@%s already registered", meta.Kind, meta.Name, meta.Version)
			}

			absPath, _ := filepath.Abs(path)
			hash := fmt.Sprintf("%x", sha256.Sum256(specJSON))

			cap := &registry.Capability{
				Kind:         meta.Kind,
				Name:         meta.Name,
				Version:      meta.Version,
				Description:  meta.Description,
				Author:       meta.Author,
				SourceType:   "local",
				Scope:        "user",
				ManifestJSON: string(specJSON),
				Hash:         hash,
				ManifestPath: absPath,
			}

			if err := app.Capabilities.Register(ctx, cap); err != nil {
				return err
			}

			cmd.Printf("registered %s/%s@%s\n", meta.Kind, meta.Name, meta.Version)
			return nil
		},
	}
}
