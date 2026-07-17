package claude

import "github.com/convexideas/oktopus/internal/runtime"

func NewAdapter() *runtime.ProcessAdapter {
	return runtime.NewProcessAdapter("claude-code", "claude")
}
