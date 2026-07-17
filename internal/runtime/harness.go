package runtime

import "context"

// Harness is the adapter interface for agent runtimes.
type Harness interface {
	Name() string
	BinaryName() string
	Start(ctx context.Context, cfg HarnessConfig) (HarnessSession, error)
	ReadConversation(sandboxHome string) (*Conversation, error)
}

// HarnessSession represents a running agent process.
type HarnessSession interface {
	ID() string
	Wait() error
	Stop() error
	ExitCode() int
	Output() string
}

// HarnessConfig holds the configuration for starting a harness.
type HarnessConfig struct {
	Workspace        string
	Env              map[string]string
	ProxyAddr        string
	Args             []string
	LogPath          string
	SystemPrompt     string
	SystemPromptMode string
	Model            string
	Tools            []string
	ExcludeTools     []string
	Task             string
	Files            map[string]string
	Internal         bool
}

// Conversation is the structured capture from a session.
type Conversation struct {
	Messages []Message
	Model    string
	Raw      string // full text for storage/summarization
}

// Message is a single conversation turn.
type Message struct {
	Role    string // "user", "assistant", "system", "tool"
	Content string
}
