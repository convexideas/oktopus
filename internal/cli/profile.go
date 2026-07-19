package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/convexideas/oktopus/internal/identity"
	"github.com/spf13/cobra"
)

const localUserID = "local"

func newProfileCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage your runtime profile",
	}
	cmd.AddCommand(
		newProfileShowCmd(app),
		newProfileSetCmd(app),
	)
	return cmd
}

func newProfileShowCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := app.Profiles.GetProfile(cmd.Context(), localUserID)
			if err != nil {
				cmd.Println("no profile configured yet (use 'ok profile set' to create one)")
				return nil
			}

			prefs := profile.Preferences
			if prefs.DefaultModel != "" {
				cmd.Printf("default_model:   %s\n", prefs.DefaultModel)
			}
			if prefs.DefaultRuntime != "" {
				cmd.Printf("default_runtime: %s\n", prefs.DefaultRuntime)
			}
			if len(prefs.ToolsExclude) > 0 {
				cmd.Printf("tools_exclude:   %s\n", strings.Join(prefs.ToolsExclude, ", "))
			}
			if len(prefs.Extensions) > 0 {
				cmd.Println("\nextensions:")
				for _, ext := range prefs.Extensions {
					cmd.Printf("  - %s (%s %s)\n", ext.Name, ext.Command, strings.Join(ext.Args, " "))
				}
			}

			hcfg := profile.HarnessConfig
			if hcfg.Pi != nil && hcfg.Pi.ApprovalMode != "" {
				cmd.Printf("\npi.approval_mode: %s\n", hcfg.Pi.ApprovalMode)
			}
			if hcfg.Codex != nil && hcfg.Codex.ApprovalMode != "" {
				cmd.Printf("\ncodex.approval_mode: %s\n", hcfg.Codex.ApprovalMode)
			}
			if hcfg.Kiro != nil && len(hcfg.Kiro.SteeringFiles) > 0 {
				cmd.Println("\nkiro.steering_files:")
				for _, f := range hcfg.Kiro.SteeringFiles {
					cmd.Printf("  - %s\n", f)
				}
			}

			return nil
		},
	}
}

func newProfileSetCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a profile value",
		Long: `Set a profile configuration value.

Preferences (portable):
  default_model             Default model for all harnesses
  default_runtime           Default runtime (pi, codex, kiro)
  tools_exclude             Tools to exclude (comma-separated generic names)
  extensions.add            Add an extension (JSON: {"name":"...","command":"...","args":[...]})
  extensions.remove         Remove an extension by name

Harness config (harness-specific):
  pi.approval_mode          Pi approval mode
  codex.approval_mode       Codex approval mode
  kiro.steering_files       Kiro steering file paths (comma-separated)`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			key, value := args[0], args[1]

			profile, err := app.Profiles.GetProfile(ctx, localUserID)
			if err != nil {
				profile = &identity.Profile{UserID: localUserID}
			}

			prefs := &profile.Preferences
			hcfg := &profile.HarnessConfig

			switch key {
			// Preferences
			case "default_model":
				prefs.DefaultModel = value
			case "default_runtime":
				prefs.DefaultRuntime = value
			case "tools_exclude":
				prefs.ToolsExclude = strings.Split(value, ",")
			case "extensions.add":
				var ext identity.Extension
				if err := json.Unmarshal([]byte(value), &ext); err != nil {
					return fmt.Errorf("invalid extension JSON: %w", err)
				}
				if ext.Name == "" {
					return fmt.Errorf("extension must have a name")
				}
				prefs.Extensions = append(prefs.Extensions, ext)
			case "extensions.remove":
				filtered := prefs.Extensions[:0]
				for _, ext := range prefs.Extensions {
					if ext.Name != value {
						filtered = append(filtered, ext)
					}
				}
				prefs.Extensions = filtered

			// Harness config
			case "pi.approval_mode":
				if hcfg.Pi == nil {
					hcfg.Pi = &identity.PiConfig{}
				}
				hcfg.Pi.ApprovalMode = value
			case "codex.approval_mode":
				if hcfg.Codex == nil {
					hcfg.Codex = &identity.CodexConfig{}
				}
				hcfg.Codex.ApprovalMode = value
			case "kiro.steering_files":
				if hcfg.Kiro == nil {
					hcfg.Kiro = &identity.KiroConfig{}
				}
				hcfg.Kiro.SteeringFiles = strings.Split(value, ",")

			default:
				return fmt.Errorf("unknown key %q", key)
			}

			if err := app.Profiles.UpsertProfile(ctx, profile); err != nil {
				return err
			}

			cmd.Printf("set %s = %s\n", key, value)
			return nil
		},
	}
}
