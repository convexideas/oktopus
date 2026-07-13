package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/convexideas/oktopus/internal/config"
	"github.com/convexideas/oktopus/internal/execution"
	"github.com/convexideas/oktopus/internal/execution/claude"
	"github.com/convexideas/oktopus/internal/execution/codex"
	"github.com/convexideas/oktopus/internal/execution/kiro"
	"github.com/convexideas/oktopus/internal/execution/pi"
	"github.com/convexideas/oktopus/internal/store/sqlite"
)

// App holds resolved dependencies for all CLI commands.
type App struct {
	Config  *config.Config
	Store   *sqlite.Store
	Harness *execution.HarnessRegistry
	Log     *slog.Logger
}

// NewApp initializes the application.
func NewApp() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	store, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("opening store: %w", err)
	}

	harnesses := execution.NewHarnessRegistry()
	harnesses.Register(pi.NewAdapter())
	harnesses.Register(codex.NewAdapter())
	harnesses.Register(kiro.NewAdapter())
	harnesses.Register(claude.NewAdapter())

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	return &App{
		Config:  cfg,
		Store:   store,
		Harness: harnesses,
		Log:     log,
	}, nil
}

// Close cleans up resources.
func (a *App) Close() {
	if a.Store != nil {
		a.Store.Close()
	}
}
