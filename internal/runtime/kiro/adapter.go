package kiro

import "github.com/convexideas/oktopus/internal/runtime"

func NewAdapter() *runtime.ProcessAdapter {
	return runtime.NewProcessAdapter("kiro", "kiro")
}
