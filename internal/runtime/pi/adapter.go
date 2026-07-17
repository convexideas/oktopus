// Package pi provides the Pi harness adapter with full config translation.
// Task mode uses --mode json for structured capture. Interactive uses native file reading.
package pi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/convexideas/oktopus/internal/runtime"
)

// toolMap maps generic tool names to Pi-specific names.
var toolMap = map[string]string{
	"web_browser": "computer_use",
	"shell":       "bash",
	"file_write":  "file_editor",
	"file_read":   "file_reader",
}

// NewAdapter returns a Pi harness adapter with Pi-specific arg building and native capture.
func NewAdapter() *runtime.ProcessAdapter {
	a := runtime.NewProcessAdapter("pi", "pi")
	a.ArgsBuilder = buildArgs
	a.ConversationReader = readConversation
	return a
}

func buildArgs(cfg runtime.HarnessConfig) []string {
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

	// Non-interactive: use --mode json for structured capture
	if cfg.Task != "" {
		args = append(args, "--mode", "json", "--print", cfg.Task)
	}

	args = append(args, cfg.Args...)
	return args
}

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

// readConversation parses Pi's structured output.
// For task mode: parses JSON stream from captured stdout.
// For interactive mode: reads session JSONL from sandbox HOME.
func readConversation(sandboxHome string) (*runtime.Conversation, error) {
	// Try reading from Pi's session directory in the sandbox HOME
	sessDir := filepath.Join(sandboxHome, ".pi", "agent", "sessions")
	if _, err := os.Stat(sessDir); err != nil {
		return nil, nil
	}

	// Find most recent session file
	latest := findLatestSession(sessDir)
	if latest == "" {
		return nil, nil
	}

	return parseSessionJSONL(latest)
}

// ParseJSONOutput parses Pi's --mode json output stream into a Conversation.
// Called by the adapter post-session when task mode was used.
func ParseJSONOutput(output string) *runtime.Conversation {
	conv := &runtime.Conversation{}
	var rawParts []string

	for _, line := range strings.Split(output, "\n") {
		if line == "" || line[0] != '{' {
			continue
		}

		var entry jsonEntry
		if json.Unmarshal([]byte(line), &entry) != nil {
			continue
		}

		switch entry.Type {
		case "message_end":
			role := entry.Message.Role
			if role == "custom" || role == "" {
				continue
			}
			var text string
			for _, c := range entry.Message.Content {
				if c.Type == "text" && c.Text != "" {
					text += c.Text
				}
			}
			if text != "" {
				conv.Messages = append(conv.Messages, runtime.Message{Role: role, Content: text})
				rawParts = append(rawParts, fmt.Sprintf("[%s]: %s", role, text))
			}
			// Capture model and cost from assistant message_end
			if role == "assistant" && entry.Message.Model != "" {
				conv.Model = entry.Message.Model
			}
		}
	}

	conv.Raw = strings.Join(rawParts, "\n\n")
	return conv
}

func findLatestSession(sessDir string) string {
	entries, _ := os.ReadDir(sessDir)
	var latest string
	var latestName string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		wsDir := filepath.Join(sessDir, entry.Name())
		files, _ := os.ReadDir(wsDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".jsonl") && f.Name() > latestName {
				latestName = f.Name()
				latest = filepath.Join(wsDir, f.Name())
			}
		}
	}
	return latest
}

func parseSessionJSONL(path string) (*runtime.Conversation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading session file: %w", err)
	}

	conv := &runtime.Conversation{}
	var rawParts []string

	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		var entry jsonEntry
		if json.Unmarshal([]byte(line), &entry) != nil {
			continue
		}

		switch entry.Type {
		case "model_change":
			conv.Model = entry.ModelID
		case "message":
			role := entry.Message.Role
			if role == "custom" || role == "" {
				continue
			}
			var text string
			for _, c := range entry.Message.Content {
				if c.Type == "text" && c.Text != "" {
					text += c.Text + "\n"
				}
			}
			if text != "" {
				text = strings.TrimSpace(text)
				conv.Messages = append(conv.Messages, runtime.Message{Role: role, Content: text})
				rawParts = append(rawParts, fmt.Sprintf("[%s]: %s", role, text))
			}
		}
	}

	conv.Raw = strings.Join(rawParts, "\n\n")
	return conv, nil
}

// JSON types for Pi output parsing
type jsonEntry struct {
	Type    string      `json:"type"`
	ModelID string      `json:"modelId"`
	Message jsonMessage `json:"message"`
}

type jsonMessage struct {
	Role    string        `json:"role"`
	Content []jsonContent `json:"content"`
	Model   string        `json:"model"`
}

type jsonContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Keep sort imported for potential use
var _ = sort.Strings
