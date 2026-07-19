package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/convexideas/oktopus/internal/credential"
	"github.com/convexideas/oktopus/internal/runtime"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// initConfig is the YAML template written by ok init.
const initConfigTemplate = `# ok configuration — edit freely, or re-run 'ok init'
# Docs: https://github.com/convexideas/oktopus/docs/design/config-schema.yaml

api_providers:
  %s:
    protocol: %s
    base_url: %s
    credential: %s

sandbox_providers:
  local:
    type: auto  # seatbelt on macOS, bwrap on Linux

defaults:
  api_provider: %s
  sandbox_provider: local
  harness: %s
  model: %s
`

type apiChoice struct {
	Name     string
	Protocol string
	BaseURL  string
}

var apiChoices = []apiChoice{
	{Name: "anthropic", Protocol: "anthropic", BaseURL: "https://api.anthropic.com"},
	{Name: "openai", Protocol: "openai", BaseURL: "https://api.openai.com/v1"},
	{Name: "openrouter", Protocol: "openai", BaseURL: "https://openrouter.ai/api/v1"},
	{Name: "ollama", Protocol: "openai", BaseURL: "http://localhost:11434/v1"},
}

var harnessChoices = []string{"pi", "claude", "codex", "kiro"}

func newInitCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Interactive setup — configure providers, credentials, and defaults",
		Long: `Guided setup that creates your ~/.ok/config.yaml and stores
credentials in the system keychain. Run again to reconfigure.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)
			home := runtime.Home()

			cmd.Println("\nWelcome to ok.")
			cmd.Println()

			// 1. API provider
			cmd.Println("? API provider:")
			for i, c := range apiChoices {
				cmd.Printf("  [%d] %s\n", i+1, c.Name)
			}
			cmd.Print("  [5] Custom endpoint\n\n")

			api := promptChoice(reader, cmd, "> ", 1, 5)
			var chosen apiChoice
			var credName string
			if api <= len(apiChoices) {
				chosen = apiChoices[api-1]
				credName = chosen.Name
			} else {
				// Custom
				cmd.Print("? Protocol (anthropic/openai): ")
				proto := promptLine(reader)
				cmd.Print("? Base URL: ")
				baseURL := promptLine(reader)
				cmd.Print("? Credential name: ")
				credName = promptLine(reader)
				chosen = apiChoice{Name: credName, Protocol: proto, BaseURL: baseURL}
			}

			// 2. Credential
			if chosen.Name != "ollama" {
				cmd.Printf("\n? Paste your %s API key: ", chosen.Name)
				key, err := term.ReadPassword(int(os.Stdin.Fd()))
				cmd.Println()
				if err != nil {
					return fmt.Errorf("reading key: %w", err)
				}
				if len(key) > 0 {
					if err := credential.Store(credName, string(key)); err != nil {
						return fmt.Errorf("storing credential: %w", err)
					}
					cmd.Printf("  ✓ Stored in system keychain as ok:%s\n", credName)
				}
			}

			// 3. Default model
			cmd.Print("\n? Default model (leave blank for provider default): ")
			model := promptLine(reader)

			// 4. Default harness
			cmd.Println("\n? Default harness:")
			for i, h := range harnessChoices {
				cmd.Printf("  [%d] %s\n", i+1, h)
			}
			hIdx := promptChoice(reader, cmd, "> ", 1, len(harnessChoices))
			harness := harnessChoices[hIdx-1]

			// 5. Write config
			if err := os.MkdirAll(home, 0o755); err != nil {
				return fmt.Errorf("creating config dir: %w", err)
			}

			cfgPath := filepath.Join(home, "config.yaml")
			content := fmt.Sprintf(initConfigTemplate,
				chosen.Name, chosen.Protocol, chosen.BaseURL, credName,
				chosen.Name, harness, model,
			)

			if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
				return fmt.Errorf("writing config: %w", err)
			}

			// 6. Show sandbox detection
			sbType, sbWarning := runtime.DetectDefaultType()
			sbStatus := fmt.Sprintf("local/%s", sbType)
			if sbWarning != "" {
				sbStatus += fmt.Sprintf(" (%s)", sbWarning)
			}

			cmd.Printf("\n  ✓ Wrote %s\n", cfgPath)
			cmd.Printf("  ✓ Default sandbox: %s\n", sbStatus)
			cmd.Printf("\n  Ready. Run: ok run\n\n")

			return nil
		},
	}
	return cmd
}

func promptLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func promptChoice(reader *bufio.Reader, cmd *cobra.Command, prompt string, min, max int) int {
	for {
		cmd.Print(prompt)
		line := promptLine(reader)
		var n int
		if _, err := fmt.Sscanf(line, "%d", &n); err == nil && n >= min && n <= max {
			return n
		}
		cmd.Printf("  please enter %d-%d\n", min, max)
	}
}
