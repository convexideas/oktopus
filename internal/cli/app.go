package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/convexideas/oktopus/internal/config"
	"github.com/convexideas/oktopus/internal/gateway"
	"github.com/convexideas/oktopus/internal/identity"
	"github.com/convexideas/oktopus/internal/memory"
	"github.com/convexideas/oktopus/internal/portal"
	"github.com/convexideas/oktopus/internal/registry"
	"github.com/convexideas/oktopus/internal/runtime"
	"github.com/convexideas/oktopus/internal/runtime/claude"
	"github.com/convexideas/oktopus/internal/runtime/codex"
	"github.com/convexideas/oktopus/internal/runtime/kiro"
	"github.com/convexideas/oktopus/internal/runtime/pi"
	"github.com/convexideas/oktopus/internal/store/sqlite"
)

// App holds resolved dependencies for all CLI commands.
// Fields are interfaces — backed by sqlite locally, portal remotely.
type App struct {
	Config       *config.Config
	Sessions     runtime.SessionStore
	Workspaces   runtime.WorkspaceStore
	Captures     gateway.CaptureStore
	Memory       memory.Store
	Profiles     identity.Store
	Capabilities registry.Store
	Portal       portal.Client
	Harness      *runtime.HarnessRegistry
	Log          *slog.Logger

	closer func() // cleanup function
}

// NewApp initializes the application with SQLite backing all stores.
func NewApp() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	store, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("opening store: %w", err)
	}

	harnesses := runtime.NewHarnessRegistry()
	harnesses.Register(pi.NewAdapter())
	harnesses.Register(codex.NewAdapter())
	harnesses.Register(kiro.NewAdapter())
	harnesses.Register(claude.NewAdapter())

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	return &App{
		Config:       cfg,
		Sessions:     store,
		Workspaces:   store,
		Captures:     store,
		Memory:       store,
		Profiles:     store,
		Capabilities: store,
		Portal:       portal.NewLocalClient(cfg.HomeDir),
		Harness:      harnesses,
		Log:          log,
		closer:       func() { store.Close() },
	}, nil
}

// Close cleans up resources.
func (a *App) Close() {
	if a.closer != nil {
		a.closer()
	}
}
