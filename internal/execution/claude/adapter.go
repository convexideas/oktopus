package claude

import "github.com/convexideas/oktopus/internal/execution"

func NewAdapter() *execution.ProcessAdapter {
	return execution.NewProcessAdapter("claude-code", "claude")
}
