package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"

	"github.com/convexideas/oktopus/internal/config"
)

// SandboxType is the isolation mechanism — the only thing code knows.
type SandboxType string

const (
	SandboxSeatbelt  SandboxType = "seatbelt"  // macOS kernel sandbox
	SandboxBwrap     SandboxType = "bwrap"     // bubblewrap user namespaces
	SandboxContainer SandboxType = "container" // OCI container
	SandboxVM        SandboxType = "vm"        // Firecracker, etc.
	SandboxProcess   SandboxType = "process"   // bare process — degraded, no enforcement
)

// Enforced returns true if this sandbox type provides OS-level enforcement.
// Gateway requires an enforced sandbox to start.
func (t SandboxType) Enforced() bool {
	switch t {
	case SandboxSeatbelt, SandboxBwrap, SandboxContainer, SandboxVM:
		return true
	default:
		return false
	}
}

// Sandbox is a persistent, named execution environment within a workspace.
type Sandbox struct {
	Home      string
	Workspace string
	Name      string
	Type      SandboxType
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

// SandboxDriver creates and manages sandboxes.
type SandboxDriver interface {
	Name() string
	Type() SandboxType
	Create(workspace, name string, cfg map[string]any) (*Sandbox, error)
	Get(workspace, name string) (*Sandbox, error)
	List(workspace string) ([]string, error)
	Destroy(workspace, name string) error
}

// DetectDefaultType returns the best sandbox type available on this OS.
func DetectDefaultType() (SandboxType, string) {
	switch goruntime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("sandbox-exec"); err == nil {
			return SandboxSeatbelt, ""
		}
		return SandboxProcess, "sandbox-exec not found (unexpected on macOS)"
	case "linux":
		if _, err := exec.LookPath("bwrap"); err == nil {
			return SandboxBwrap, ""
		}
		return SandboxProcess, "bwrap not found — install: sudo apt install bubblewrap"
	default:
		return SandboxProcess, fmt.Sprintf("no OS sandbox available for %s", goruntime.GOOS)
	}
}

// ResolveSandboxDriver resolves from a config.SandboxProvider entry.
// The provider config is data (user-managed), type resolution is the only logic.
func ResolveSandboxDriver(sp *config.SandboxProvider) (SandboxDriver, string) {
	sandboxType := SandboxType(sp.Type)

	// "auto" means detect the best available for this OS
	if sp.Type == "" || sp.Type == "auto" {
		detected, warning := DetectDefaultType()
		sandboxType = detected
		if !sandboxType.Enforced() {
			return &LocalProcess{}, warning
		}
		return &LocalSandboxed{sandboxType: detected}, warning
	}

	switch sandboxType {
	case SandboxSeatbelt, SandboxBwrap:
		return &LocalSandboxed{sandboxType: sandboxType}, ""
	case SandboxContainer:
		// ponytail: container driver not yet implemented
		return &LocalProcess{}, "container driver not implemented — falling back to bare process"
	case SandboxVM:
		return &LocalProcess{}, "vm driver not implemented — falling back to bare process"
	default:
		return &LocalProcess{}, ""
	}
}

// LocalSandboxed wraps a local process with OS-level enforcement (seatbelt or bwrap).
// ponytail: today creates directory structure. Actual seatbelt/bwrap exec wrapping
// is wired in the harness launch path. This holds the sandbox config/profile.
type LocalSandboxed struct {
	sandboxType SandboxType
}

func (d *LocalSandboxed) Name() string      { return "local/" + string(d.sandboxType) }
func (d *LocalSandboxed) Type() SandboxType { return d.sandboxType }

func (d *LocalSandboxed) homePath(workspace, name string) string {
	return filepath.Join(Home(), workspace, "sandboxes", name, "home")
}

func (d *LocalSandboxed) Create(workspace, name string, cfg map[string]any) (*Sandbox, error) {
	home := d.homePath(workspace, name)
	if err := os.MkdirAll(home, 0o755); err != nil {
		return nil, fmt.Errorf("creating sandbox: %w", err)
	}
	return &Sandbox{Home: home, Workspace: workspace, Name: name, Type: d.sandboxType}, nil
}

func (d *LocalSandboxed) Get(workspace, name string) (*Sandbox, error) {
	home := d.homePath(workspace, name)
	if _, err := os.Stat(home); err != nil {
		return nil, fmt.Errorf("sandbox %s:%s not found", workspace, name)
	}
	return &Sandbox{Home: home, Workspace: workspace, Name: name, Type: d.sandboxType}, nil
}

func (d *LocalSandboxed) List(workspace string) ([]string, error) {
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

func (d *LocalSandboxed) Destroy(workspace, name string) error {
	return os.RemoveAll(filepath.Join(Home(), workspace, "sandboxes", name))
}

// LocalProcess is the degraded fallback: mkdir + HOME redirect. No isolation.
type LocalProcess struct{}

func (d *LocalProcess) Name() string      { return "local/process" }
func (d *LocalProcess) Type() SandboxType { return SandboxProcess }

func (d *LocalProcess) homePath(workspace, name string) string {
	return filepath.Join(Home(), workspace, "sandboxes", name, "home")
}

func (d *LocalProcess) Create(workspace, name string, cfg map[string]any) (*Sandbox, error) {
	home := d.homePath(workspace, name)
	if err := os.MkdirAll(home, 0o755); err != nil {
		return nil, fmt.Errorf("creating sandbox: %w", err)
	}
	return &Sandbox{Home: home, Workspace: workspace, Name: name, Type: SandboxProcess}, nil
}

func (d *LocalProcess) Get(workspace, name string) (*Sandbox, error) {
	home := d.homePath(workspace, name)
	if _, err := os.Stat(home); err != nil {
		return nil, fmt.Errorf("sandbox %s:%s not found", workspace, name)
	}
	return &Sandbox{Home: home, Workspace: workspace, Name: name, Type: SandboxProcess}, nil
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
	return os.RemoveAll(filepath.Join(Home(), workspace, "sandboxes", name))
}
