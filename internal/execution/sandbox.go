package execution

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Sandbox represents an active isolated execution environment.
// Dir IS the workspace root — harness runs from it directly.
type Sandbox struct {
	Dir string
}

// WriteFile writes content to a path relative to the sandbox root.
func (s *Sandbox) WriteFile(relPath, content string) error {
	absPath := filepath.Join(s.Dir, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(absPath), err)
	}
	return os.WriteFile(absPath, []byte(content), 0o644)
}

// SandboxProvider creates and manages sandboxes.
type SandboxProvider interface {
	Create(workspaceSource string) (*Sandbox, error)
	Destroy(s *Sandbox) error
}

// DirectorySandbox copies workspace source into a temp dir.
// ponytail: naive copy. Upgrade path: git worktree, overlayfs, bwrap, container.
type DirectorySandbox struct{}

func (p *DirectorySandbox) Create(workspaceSource string) (*Sandbox, error) {
	dir, err := os.MkdirTemp("", "ok-sandbox-*")
	if err != nil {
		return nil, fmt.Errorf("creating sandbox: %w", err)
	}

	if workspaceSource != "" {
		if err := cloneDir(workspaceSource, dir); err != nil {
			os.RemoveAll(dir)
			return nil, fmt.Errorf("cloning workspace: %w", err)
		}
	}

	return &Sandbox{Dir: dir}, nil
}

func (p *DirectorySandbox) Destroy(s *Sandbox) error {
	if s == nil || s.Dir == "" {
		return nil
	}
	return os.RemoveAll(s.Dir)
}

// cloneDir copies workspace source into dst using copy-on-write where available.
// macOS (APFS): cp -Rc (instant clonefile)
// Linux (btrfs/xfs): cp -a --reflink=auto (instant reflink, fallback to copy)
// ponytail: shelling out to cp. Ceiling: non-POSIX systems. Upgrade: syscall-level clonefile.
func cloneDir(src, dst string) error {
	// Remove the temp dir — cp needs to create the target itself
	os.RemoveAll(dst)

	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("cp", "-Rc", src, dst)
	} else {
		cmd = exec.Command("cp", "-a", "--reflink=auto", src, dst)
	}

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cp: %s: %w", string(out), err)
	}
	return nil
}
