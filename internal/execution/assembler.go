package execution

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/convexideas/oktopus/internal/identity"
	"github.com/convexideas/oktopus/internal/registry"
)

// Layout maps abstract concepts to filesystem paths within the sandbox.
type Layout struct {
	SystemPromptPath   string
	SkillsDir          string
	ExtensionsPath     string
	ExtensionFormatter func([]identity.Extension) (string, error)
}

// Known layouts.
var (
	PiLayout = Layout{
		SystemPromptPath:   ".pi/SYSTEM.md",
		SkillsDir:          ".pi/skills",
		ExtensionsPath:     ".pi/extensions.json",
		ExtensionFormatter: formatPiExtensions,
	}

	CodexLayout = Layout{
		SystemPromptPath: "AGENTS.md",
		SkillsDir:        ".codex/skills",
	}

	KiroLayout = Layout{
		SystemPromptPath: ".kiro/steering/persona.md",
		SkillsDir:        ".kiro/steering",
	}
)

// LayoutFor returns the layout for a given harness name.
func LayoutFor(harnessName string) Layout {
	switch harnessName {
	case "pi":
		return PiLayout
	case "codex":
		return CodexLayout
	case "kiro":
		return KiroLayout
	default:
		return PiLayout
	}
}

// Assemble materializes preferences + persona into the sandbox.
func Assemble(sb *Sandbox, layout Layout, prefs *identity.Preferences, hcfg *identity.HarnessConfig, harnessName string, persona *registry.PersonaSpec) error {
	if persona == nil && prefs == nil {
		return nil
	}

	if persona != nil {
		if err := materializePersona(sb, layout, persona); err != nil {
			return err
		}
	}

	if prefs != nil || hcfg != nil {
		if err := materializeExtensions(sb, layout, prefs, hcfg, harnessName); err != nil {
			return err
		}
	}

	return nil
}

func materializePersona(sb *Sandbox, layout Layout, persona *registry.PersonaSpec) error {
	if persona.SystemPrompt != "" && layout.SystemPromptPath != "" {
		if err := sb.WriteFile(layout.SystemPromptPath, persona.SystemPrompt); err != nil {
			return fmt.Errorf("writing system prompt: %w", err)
		}
	}

	if len(persona.Skills) > 0 && layout.SkillsDir != "" {
		for _, skill := range persona.Skills {
			name := strings.ReplaceAll(strings.ToLower(skill), " ", "-")
			content := fmt.Sprintf("# Skill: %s\n", skill)
			filePath := path.Join(layout.SkillsDir, name+".md")
			if err := sb.WriteFile(filePath, content); err != nil {
				return fmt.Errorf("writing skill %s: %w", name, err)
			}
		}
	}

	for _, native := range persona.Native {
		for relPath, content := range native.Files {
			if err := sb.WriteFile(relPath, content); err != nil {
				return fmt.Errorf("writing native file %s: %w", relPath, err)
			}
		}
	}

	return nil
}

// materializeExtensions merges portable + harness-specific extensions and writes them.
func materializeExtensions(sb *Sandbox, layout Layout, prefs *identity.Preferences, hcfg *identity.HarnessConfig, harnessName string) error {
	if layout.ExtensionsPath == "" {
		return nil
	}

	// Collect portable extensions
	var all []identity.Extension
	if prefs != nil {
		all = append(all, prefs.Extensions...)
	}

	// Collect harness-specific extensions
	if hcfg != nil {
		switch harnessName {
		case "pi":
			if hcfg.Pi != nil {
				all = append(all, hcfg.Pi.Extensions...)
			}
		case "codex":
			if hcfg.Codex != nil {
				all = append(all, hcfg.Codex.Extensions...)
			}
		case "kiro":
			if hcfg.Kiro != nil {
				all = append(all, hcfg.Kiro.Extensions...)
			}
		}
	}

	if len(all) == 0 {
		return nil
	}

	formatter := layout.ExtensionFormatter
	if formatter == nil {
		formatter = formatPiExtensions
	}

	content, err := formatter(all)
	if err != nil {
		return fmt.Errorf("formatting extensions: %w", err)
	}

	return sb.WriteFile(layout.ExtensionsPath, content)
}

func formatPiExtensions(exts []identity.Extension) (string, error) {
	servers := make(map[string]any, len(exts))
	for _, ext := range exts {
		entry := map[string]any{"command": ext.Command}
		if len(ext.Args) > 0 {
			entry["args"] = ext.Args
		}
		if len(ext.Env) > 0 {
			entry["env"] = ext.Env
		}
		servers[ext.Name] = entry
	}

	data, err := json.MarshalIndent(map[string]any{"mcpServers": servers}, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
