package runtime

import (
	"fmt"
	"os"
	"path/filepath"
)

// Sandbox is a persistent, named execution environment within a workspace.
type Sandbox struct {
	Home      string
	Workspace string
	Name      string
}

// WriteFile writes content relative to sandbox Home.
func (s *Sandbox) WriteFile(relPath, content string) error {
	absPath := filepath.Join(s.Home, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(absPath), err)
	}
	return os.WriteFile(absPath, []byte(content), 0o644)
}

// Env returns environment overrides for the harness process.
func (s *Sandbox) Env() map[string]string {
	return map[string]string{
		"HOME": s.Home,
	}
}

// SandboxDriver creates and manages sandboxes for a specific provider+type combination.
type SandboxDriver interface {
	Name() string
	Create(workspace, name string, config map[string]any) (*Sandbox, error)
	Get(workspace, name string) (*Sandbox, error)
	List(workspace string) ([]string, error)
	Destroy(workspace, name string) error
}

// ResolveSandboxDriver returns the appropriate driver for the given sandbox config.
// Resolution: explicit config > workspace default > "local" provider + "process" type.
func ResolveSandboxDriver(cfg SandboxConfig) SandboxDriver {
	// ponytail: only local/process exists today. Add docker, modal, etc. here.
	switch cfg.Provider + "/" + cfg.Type {
	default:
		return &LocalProcess{}
	}
}

// LocalProcess is the simplest sandbox driver: mkdir + HOME redirect.
// Provider: local. Type: process. No real isolation.
type LocalProcess struct{}

func (d *LocalProcess) Name() string { return "local/process" }

func (d *LocalProcess) homePath(workspace, name string) string {
	return filepath.Join(Home(), workspace, "sandboxes", name, "home")
}

func (d *LocalProcess) Create(workspace, name string, config map[string]any) (*Sandbox, error) {
	home := d.homePath(workspace, name)
	if err := os.MkdirAll(home, 0o755); err != nil {
		return nil, fmt.Errorf("creating sandbox: %w", err)
	}
	return &Sandbox{Home: home, Workspace: workspace, Name: name}, nil
}

func (d *LocalProcess) Get(workspace, name string) (*Sandbox, error) {
	home := d.homePath(workspace, name)
	if _, err := os.Stat(home); err != nil {
		return nil, fmt.Errorf("sandbox %s:%s not found", workspace, name)
	}
	return &Sandbox{Home: home, Workspace: workspace, Name: name}, nil
}

func (d *LocalProcess) List(workspace string) ([]string, error) {
	dir := filepath.Join(Home(), workspace, "sandboxes")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func (d *LocalProcess) Destroy(workspace, name string) error {
	dir := filepath.Join(Home(), workspace, "sandboxes", name)
	return os.RemoveAll(dir)
}
