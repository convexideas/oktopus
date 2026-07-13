package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/convexideas/oktopus/internal/execution"
	"github.com/convexideas/oktopus/internal/identity"
	"github.com/convexideas/oktopus/internal/memory"
	"github.com/convexideas/oktopus/internal/policy"
	"github.com/convexideas/oktopus/internal/registry"
	"github.com/spf13/cobra"
)

func newRunCmd(app *App) *cobra.Command {
	var personaFlag string
	var workspaceFlag string
	var runtimeFlag string
	var taskFlag string

	cmd := &cobra.Command{
		Use:   "run [agent] [-- agent-args...]",
		Short: "Launch agent in a tracked session",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var agentName string
			var agentArgs []string
			if len(args) > 0 {
				agentName = args[0]
				if len(args) > 1 {
					agentArgs = args[1:]
				}
			}

			// --- Resolution ---

			harnessName := agentName
			if runtimeFlag != "" {
				harnessName = runtimeFlag
			}

			// Profile
			profile, _ := app.Store.GetProfile(ctx, localUserID)
			var prefs identity.Preferences
			var hcfg identity.HarnessConfig
			if profile != nil {
				prefs = profile.Preferences
				hcfg = profile.HarnessConfig
			}
			if harnessName == "" && prefs.DefaultRuntime != "" {
				harnessName = prefs.DefaultRuntime
			}

			// Persona
			var personaSpec *registry.PersonaSpec
			var personaLabel string
			if personaFlag != "" {
				cap, err := app.Store.FindCapabilityByName(ctx, "persona", personaFlag)
				if err != nil {
					return fmt.Errorf("persona %q not registered", personaFlag)
				}
				var spec registry.PersonaSpec
				if err := json.Unmarshal([]byte(cap.ManifestJSON), &spec); err != nil {
					return fmt.Errorf("parsing persona manifest: %w", err)
				}
				personaSpec = &spec
				personaLabel = cap.Name

				if harnessName == "" && len(spec.Native) == 1 {
					for name := range spec.Native {
						harnessName = name
					}
				}
			}

			if harnessName == "" {
				return fmt.Errorf("specify an agent or --runtime")
			}

			h, err := app.Harness.Get(harnessName)
			if err != nil {
				return err
			}

			// Workspace
			workspacePath, _ := os.Getwd()
			if workspaceFlag != "" {
				ws, err := app.Store.GetWorkspaceByName(ctx, workspaceFlag)
				if err != nil {
					return fmt.Errorf("workspace %q not found", workspaceFlag)
				}
				refs, _ := app.Store.ListWorkspaceRefs(ctx, ws.ID)
				for _, r := range refs {
					if r.Kind == "source" {
						workspacePath = r.Ref
						break
					}
				}
			}

			// --- Sandbox + Assembly ---

			// Memory injection: load recent summaries for context
			if workspaceFlag != "" {
				if ws, err := app.Store.GetWorkspaceByName(ctx, workspaceFlag); err == nil {
					summaries, _ := app.Store.ListSummariesByWorkspace(ctx, ws.ID, 5)
					if len(summaries) > 0 && personaSpec != nil {
						var memCtx string
						for i := len(summaries) - 1; i >= 0; i-- {
							memCtx += fmt.Sprintf("- [%s] %s\n", summaries[i].CapturedAt, summaries[i].Content)
						}
						personaSpec.SystemPrompt += "\n\n## Context from prior sessions\n\n" + memCtx
					}
				}
			}

			provider := &execution.DirectorySandbox{}
			sb, err := provider.Create(workspacePath)
			if err != nil {
				return fmt.Errorf("creating sandbox: %w", err)
			}
			defer provider.Destroy(sb)

			layout := execution.LayoutFor(harnessName)
			if err := execution.Assemble(sb, layout, &prefs, &hcfg, harnessName, personaSpec); err != nil {
				return fmt.Errorf("assembling sandbox: %w", err)
			}

			// --- Harness config ---

			cfg := execution.HarnessConfig{
				Workspace: sb.Dir,
				Args:      agentArgs,
				Env:       make(map[string]string),
				Model:     resolveModel(harnessName, &prefs, &hcfg),
			}

			cfg.ExcludeTools = prefs.ToolsExclude

			if taskFlag != "" {
				cfg.Task = taskFlag
			}

			if personaSpec != nil {
				if native, ok := personaSpec.Native[harnessName]; ok {
					cfg.Args = append(native.Flags, cfg.Args...)
				}
			}

			if app.Config.ProxyAddr != "" {
				cfg.ProxyAddr = app.Config.ProxyAddr
			}
			if app.Config.LogDir != "" {
				os.MkdirAll(app.Config.LogDir, 0o755)
				ts := time.Now().Format("2006-01-02T15-04-05")
				cfg.LogPath = filepath.Join(app.Config.LogDir, fmt.Sprintf("%s-%s.log", harnessName, ts))
			}

			// --- Policy ---

			hook := policy.Noop{}
			decision, reason, err := hook.OnSessionStart(ctx, cfg)
			if err != nil {
				return fmt.Errorf("policy error: %w", err)
			}
			if decision == policy.Deny {
				return fmt.Errorf("denied by policy: %s", reason)
			}

			// --- Launch ---

			sess, err := h.Start(ctx, cfg)
			if err != nil {
				return fmt.Errorf("starting %s: %w", harnessName, err)
			}

			dbSess := &execution.Session{
				ID:        sess.ID(),
				Agent:     harnessName,
				Workspace: workspacePath,
				Title:     harnessName,
			}
			app.Store.CreateSession(ctx, dbSess)

			label := harnessName
			if personaLabel != "" {
				label += ", persona=" + personaLabel
			}
			app.Log.Info("session started", "id", sess.ID()[:8], "agent", label, "sandbox", sb.Dir)

			sigCtx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
			defer cancel()
			go func() {
				<-sigCtx.Done()
				sess.Stop()
			}()

			waitErr := sess.Wait()
			status := "completed"
			if waitErr != nil {
				status = "failed"
			}

			app.Store.CompleteSession(ctx, sess.ID(), status)
			hook.OnSessionEnd(ctx, sess.ID(), policy.Result{ExitCode: sess.ExitCode(), Status: status})

			// Memory capture: save output as episode
			if output := sess.Output(); output != "" {
				wsID := ""
				if workspaceFlag != "" {
					if ws, err := app.Store.GetWorkspaceByName(ctx, workspaceFlag); err == nil {
						wsID = ws.ID
					}
				}
				// Save raw output
				ep := &memory.Episode{
					SessionID:   sess.ID(),
					WorkspaceID: wsID,
					Source:      "stdout",
					Content:     output,
				}
				app.Store.SaveEpisode(ctx, ep)

				// Generate and save summary
				if wsID != "" {
					summary := summarize(ctx, app, harnessName, output)
					if summary != output { // only save if LLM actually produced something different
						sumEp := &memory.Episode{
							SessionID:   sess.ID(),
							WorkspaceID: wsID,
							Source:      "summary",
							Content:     summary,
						}
						app.Store.SaveEpisode(ctx, sumEp)
					}
				}
			}

			if sess.ExitCode() != 0 {
				return fmt.Errorf("agent exited with code %d", sess.ExitCode())
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&personaFlag, "persona", "p", "", "Persona to use")
	cmd.Flags().StringVarP(&workspaceFlag, "workspace", "w", "", "Workspace to run in")
	cmd.Flags().StringVarP(&runtimeFlag, "runtime", "r", "", "Override harness runtime")
	cmd.Flags().StringVarP(&taskFlag, "task", "t", "", "Non-interactive task prompt")

	return cmd
}

func resolveModel(harnessName string, prefs *identity.Preferences, hcfg *identity.HarnessConfig) string {
	// Preferences default model (portable)
	if prefs != nil && prefs.DefaultModel != "" {
		return prefs.DefaultModel
	}
	return ""
}

// summarize calls the harness to produce a comprehensive summary of session output.
// Uses the same runtime that just ran. If it fails, falls back to keeping raw output.
func summarize(ctx context.Context, app *App, harnessName, output string) string {
	if len(output) < 50 {
		return output // trivial output, no summarization needed
	}

	// Cap input to avoid exceeding context windows, but be generous
	input := output
	if len(input) > 16000 {
		input = input[:16000]
	}

	h, err := app.Harness.Get(harnessName)
	if err != nil {
		return output
	}

	prompt := `Produce a detailed summary of this session. Include:
- What was accomplished
- Key decisions made
- Files created or modified
- Issues identified
- Open questions or next steps

Be thorough — this summary is the persistent memory of this session for future context.

Session output:
---
` + input + `
---

Summary:`

	cfg := execution.HarnessConfig{
		Workspace: os.TempDir(),
		Task:      prompt,
		Internal:  true,
		Env:       make(map[string]string),
	}

	// Use configured model
	profile, _ := app.Store.GetProfile(ctx, localUserID)
	if profile != nil && profile.Preferences.DefaultModel != "" {
		cfg.Model = profile.Preferences.DefaultModel
	}

	sess, err := h.Start(ctx, cfg)
	if err != nil {
		return output // fallback: keep raw
	}

	sess.Wait()
	result := strings.TrimSpace(sess.Output())
	if result == "" {
		return output // fallback: keep raw
	}
	return result
}

