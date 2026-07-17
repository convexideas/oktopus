package runtime

import (
	"fmt"
	"os"
	"path/filepath"
)

// Sandbox is a persistent, named execution environment within a workspace.
// The harness runs with HOME redirected here. Project files stay at their real path.
type Sandbox struct {
	Home      string // ~/.ok/<workspace>/sandboxes/<name>/home/
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
// ponytail: HOME redirect only for now. Upgrade: bwrap/seatbelt command wrapping.
func (s *Sandbox) Env() map[string]string {
	return map[string]string{
		"HOME": s.Home,
	}
}

// SandboxProvider creates and manages sandboxes.
// Implementations: local (mkdir), bwrap, seatbelt, docker, cloud (future).
type SandboxProvider interface {
	Create(workspace, name string) (*Sandbox, error)
	Get(workspace, name string) (*Sandbox, error)
	List(workspace string) ([]string, error)
	Destroy(workspace, name string) error
}

// LocalProvider stores sandboxes under ~/.ok/<workspace>/sandboxes/<name>/home/.
type LocalProvider struct {
	root string
}

func NewLocalProvider() *LocalProvider {
	home, _ := os.UserHomeDir()
	return &LocalProvider{root: filepath.Join(home, ".ok")}
}

func (p *LocalProvider) homePath(workspace, name string) string {
	return filepath.Join(p.root, workspace, "sandboxes", name, "home")
}

func (p *LocalProvider) Create(workspace, name string) (*Sandbox, error) {
	home := p.homePath(workspace, name)
	if err := os.MkdirAll(home, 0o755); err != nil {
		return nil, fmt.Errorf("creating sandbox: %w", err)
	}
	return &Sandbox{Home: home, Workspace: workspace, Name: name}, nil
}

func (p *LocalProvider) Get(workspace, name string) (*Sandbox, error) {
	home := p.homePath(workspace, name)
	if _, err := os.Stat(home); err != nil {
		return nil, fmt.Errorf("sandbox %s:%s not found", workspace, name)
	}
	return &Sandbox{Home: home, Workspace: workspace, Name: name}, nil
}

func (p *LocalProvider) List(workspace string) ([]string, error) {
	dir := filepath.Join(p.root, workspace, "sandboxes")
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

func (p *LocalProvider) Destroy(workspace, name string) error {
	dir := filepath.Join(p.root, workspace, "sandboxes", name)
	return os.RemoveAll(dir)
}
