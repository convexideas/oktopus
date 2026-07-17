package codex

import "github.com/convexideas/oktopus/internal/runtime"

func NewAdapter() *runtime.ProcessAdapter {
	return runtime.NewProcessAdapter("codex", "codex")
}
