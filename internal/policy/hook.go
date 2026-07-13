package policy

import (
	"context"

	"github.com/convexideas/oktopus/internal/execution"
)

// Decision represents the outcome of a policy evaluation.
type Decision int

const (
	Allow Decision = iota
	Deny
	Ask
)

// Result captures the outcome of a completed session.
type Result struct {
	ExitCode int
	Status   string
}

// Hook is the policy interception interface.
type Hook interface {
	OnSessionStart(ctx context.Context, cfg execution.HarnessConfig) (Decision, string, error)
	OnSessionEnd(ctx context.Context, sessionID string, result Result) error
}

// Noop is a policy hook that allows everything.
type Noop struct{}

func (Noop) OnSessionStart(_ context.Context, _ execution.HarnessConfig) (Decision, string, error) {
	return Allow, "", nil
}

func (Noop) OnSessionEnd(_ context.Context, _ string, _ Result) error {
	return nil
}
