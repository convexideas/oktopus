package importer

import (
	"encoding/json"
	"os"
	"strings"
)

// NewPi returns a Pi harness importer.
// Only the session parser is Pi-specific — everything else uses Base machinery.
func NewPi() Harness {
	b := &Base{
		name:       "pi",
		homeDir:    ".pi/agent",
		configFile: "settings.json",
		fieldMap: FieldMap{
			Model:    "defaultModel",
			Provider: "defaultProvider",
			Packages: "packages",
		},
		sessionDir: "sessions",
		sessionExt: ".jsonl",
	}
	b.sessionParser = parsePiSession
	return b
}

// parsePiSession extracts assistant text content from a Pi JSONL file.
func parsePiSession(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var parts []string
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		var msg struct {
			Type    string `json:"type"`
			Message struct {
				Role    string `json:"role"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &msg) != nil {
			continue
		}
		if msg.Type != "message" || msg.Message.Role != "assistant" {
			continue
		}
		for _, c := range msg.Message.Content {
			if c.Type == "text" && c.Text != "" {
				parts = append(parts, c.Text)
			}
		}
	}

	return strings.Join(parts, "\n\n"), nil
}
