// Package pi provides the Pi harness adapter with full config translation.
package pi

import (
	"strings"

	"github.com/convexideas/oktopus/internal/execution"
)

// toolMap maps generic tool names to Pi-specific names.
var toolMap = map[string]string{
	"web_browser": "computer_use",
	"shell":       "bash",
	"file_write":  "file_editor",
	"file_read":   "file_reader",
}

// NewAdapter returns a Pi harness adapter with Pi-specific arg building.
func NewAdapter() *execution.ProcessAdapter {
	a := execution.NewProcessAdapter("pi", "pi")
	a.ArgsBuilder = buildArgs
	return a
}

func buildArgs(cfg execution.HarnessConfig) []string {
	var args []string

	if cfg.Model != "" {
		args = append(args, "--model", cfg.Model)
	}

	if cfg.SystemPrompt != "" {
		if cfg.SystemPromptMode == "replace" {
			args = append(args, "--system-prompt", cfg.SystemPrompt)
		} else {
			args = append(args, "--append-system-prompt", cfg.SystemPrompt)
		}
	}

	if len(cfg.Tools) > 0 {
		args = append(args, "--tools", strings.Join(cfg.Tools, ","))
	}
	if len(cfg.ExcludeTools) > 0 {
		mapped := mapToolNames(cfg.ExcludeTools)
		args = append(args, "--exclude-tools", strings.Join(mapped, ","))
	}

	// Non-interactive mode
	if cfg.Task != "" {
		args = append(args, "--print", cfg.Task)
	}

	args = append(args, cfg.Args...)
	return args
}

// mapToolNames translates generic tool names to Pi-specific names.
// Unknown names pass through unchanged.
func mapToolNames(names []string) []string {
	mapped := make([]string, len(names))
	for i, name := range names {
		if piName, ok := toolMap[name]; ok {
			mapped[i] = piName
		} else {
			mapped[i] = name
		}
	}
	return mapped
}
