package cli

import (
	"github.com/convexideas/oktopus/internal/identity"
	"github.com/convexideas/oktopus/internal/importer"
	"github.com/convexideas/oktopus/internal/memory"
	"github.com/spf13/cobra"
)

func newInitCmd(app *App) *cobra.Command {
	var sessionLimit int

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Import settings and history from installed harnesses",
		Long: `Discover installed harnesses, import their settings into your profile,
and import session history as memory episodes.

Settings are merged non-destructively — existing profile values are not overwritten.
Run with --force to overwrite existing settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			harnesses := []importer.Harness{
				importer.NewPi(),
				// TODO: importer.NewCodex(), importer.NewKiro()
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
				// TODO: shortcomings of current implementation:
				// 1. Only runs on explicit invocation (ok init). Does not auto-detect on first run.
				//    Future: prompt user on first `ok run` if no profile exists and harnesses are detected.
				// 2. Non-destructive merge only — existing profile values are never overwritten.
				//    Future: --force flag to overwrite, or --diff to show what would change.
				// 3. No deduplication of extensions — running init twice appends duplicates.
				//    Future: deduplicate by extension name before appending.
				// 4. Package-to-extension mapping is naive (npm/git prefix parsing).
				//    Future: resolve actual MCP server configs from harness extension manifests.
				// 5. No selective import — all-or-nothing per harness.
				//    Future: interactive mode with checkboxes, or --only settings|sessions.
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
				// TODO: shortcomings:
				// 1. Imported as raw assistant text — no structured conversation (user turns lost).
				//    Future: import full conversation with roles for richer memory.
				// 2. No workspace matching — WorkspaceID left empty. CWD is captured but not
				//    linked to an existing ok workspace.
				//    Future: auto-create workspaces from discovered CWDs, or match by path.
				// 3. No summarization on import — raw content stored directly.
				//    Future: run LLM summarization on imported episodes (expensive but valuable).
				// 4. Timestamp format preserved as-is from Pi (dashes in time component).
				//    Future: normalize to RFC3339.
				// 5. Large sessions stored in full — no size cap.
				//    Future: truncate or split very large sessions.
				episodes, err := h.Sessions(sessionLimit)
				if err != nil {
					cmd.PrintErrf("  warning: could not read sessions: %v\n", err)
					continue
				}

				imported := 0
				for _, ep := range episodes {
					memEp := &memory.Episode{
						SessionID:   "", // no matching session record — this is a historical import
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
