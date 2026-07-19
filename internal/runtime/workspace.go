package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/convexideas/oktopus/internal/config"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// WorkspaceYAMLName is the filename for workspace config in managed workspaces.
const WorkspaceYAMLName = "workspace.yaml"

// ProjectYAMLName is the filename for workspace config embedded in a project root.
const ProjectYAMLName = ".ok.yaml"

// Home returns the base directory for all ok state. Delegates to config.Home().
func Home() string {
	return config.Home()
}

// ResolveWorkspaceDir returns the managed workspace directory path.
func ResolveWorkspaceDir(name string) string {
	return filepath.Join(Home(), name)
}

// WorkspaceFile is the user-facing YAML representation of a workspace.
type WorkspaceFile struct {
	Name            string            `koanf:"name"`
	Source          string            `koanf:"source"`
	SandboxDefaults SandboxConfig     `koanf:"sandbox_defaults"`
	Gateway         GatewayConfig     `koanf:"gateway"`
	Defaults        WorkspaceDefaults `koanf:"defaults"`
}

// GatewayConfig holds gateway provider routing config.
type GatewayConfig struct {
	Provider string `koanf:"provider" json:"provider,omitempty"`
}

// WorkspaceDefaults holds user-facing defaults for ok run.
type WorkspaceDefaults struct {
	Harness string `koanf:"harness" json:"harness,omitempty"`
	Model   string `koanf:"model" json:"model,omitempty"`
}

// LoadWorkspaceFile loads workspace config from a directory.
// Tries: dir/workspace.yaml, then dir/.ok.yaml.
func LoadWorkspaceFile(dir string) (*WorkspaceFile, error) {
	k := koanf.New(".")

	path := filepath.Join(dir, WorkspaceYAMLName)
	if _, err := os.Stat(path); err != nil {
		path = filepath.Join(dir, ProjectYAMLName)
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("no workspace config in %s", dir)
		}
	}

	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	var wf WorkspaceFile
	if err := k.Unmarshal("", &wf); err != nil {
		return nil, fmt.Errorf("unmarshalling workspace config: %w", err)
	}
	return &wf, nil
}

// ToWorkspace converts to a Workspace entity for DB storage.
func (wf *WorkspaceFile) ToWorkspace() *Workspace {
	cfg := WorkspaceConfig{Sandbox: wf.SandboxDefaults}
	cfgJSON, _ := json.Marshal(cfg)
	return &Workspace{
		Name:       wf.Name,
		ConfigJSON: string(cfgJSON),
	}
}
