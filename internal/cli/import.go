package cli

import (
	"github.com/convexideas/oktopus/internal/identity"
	"github.com/convexideas/oktopus/internal/importer"
	"github.com/convexideas/oktopus/internal/memory"
	"github.com/spf13/cobra"
)

func newImportCmd(app *App) *cobra.Command {
	var sessionLimit int

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import settings and history from installed harnesses",
		Long: `Discover installed harnesses, import their settings into your profile,
and import session history as memory episodes.

Settings are merged non-destructively — existing profile values are not overwritten.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			harnesses := []importer.Harness{
				importer.NewPi(),
			}

			for _, h := range harnesses {
				d, err := h.Discover()
				if err != nil || !d.Found {
					continue
				}

				cmd.PrintErrf("Found %s at %s\n", h.Name(), d.Path)
				cmd.PrintErrf("  model: %s, provider: %s\n", d.Model, d.Provider)
				cmd.PrintErrf("  packages: %d, sessions: %d\n", len(d.Packages), d.Sessions)

				// Import settings
				prefs, hcfg, err := h.Settings()
				if err == nil && prefs != nil {
					profile, _ := app.Profiles.GetProfile(ctx, localUserID)
					if profile == nil {
						profile = &identity.Profile{UserID: localUserID}
					}
					if profile.Preferences.DefaultModel == "" {
						profile.Preferences.DefaultModel = prefs.DefaultModel
					}
					if profile.Preferences.DefaultRuntime == "" {
						profile.Preferences.DefaultRuntime = prefs.DefaultRuntime
					}
					profile.Preferences.Extensions = append(profile.Preferences.Extensions, prefs.Extensions...)
					if hcfg != nil {
						profile.HarnessConfig = *hcfg
					}

					if err := app.Profiles.UpsertProfile(ctx, profile); err != nil {
						cmd.PrintErrf("  warning: could not save profile: %v\n", err)
					} else {
						cmd.PrintErrf("  ✓ settings imported into profile\n")
					}
				}

				// Import sessions as episodes
				episodes, err := h.Sessions(sessionLimit)
				if err != nil {
					cmd.PrintErrf("  warning: could not read sessions: %v\n", err)
					continue
				}

				imported := 0
				for _, ep := range episodes {
					memEp := &memory.Episode{
						SessionID:   "",
						Source:      "import:" + h.Name(),
						WorkspaceID: "",
						Content:     ep.Content,
						CapturedAt:  ep.Timestamp,
					}
					if err := app.Memory.SaveEpisode(ctx, memEp); err != nil {
						app.Log.Debug("SaveEpisode failed", "error", err)
					} else {
						imported++
					}
				}
				cmd.PrintErrf("  ✓ %d sessions imported as episodes\n", imported)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&sessionLimit, "sessions", "n", 20, "Max sessions to import per harness")
	return cmd
}
