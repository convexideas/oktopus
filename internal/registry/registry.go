package registry

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var validKinds = map[string]bool{
	"adapter":            true,
	"agent_profile":      true,
	"command":            true,
	"data_source":        true,
	"environment":        true,
	"input_channel":      true,
	"knowledge_source":   true,
	"mcp_server":         true,
	"model_provider":     true,
	"output_destination": true,
	"persona":            true,
	"policy":             true,
	"profile":            true,
	"runtime":            true,
	"skill":              true,
	"skillpack":          true,
	"solution_pack":      true,
	"task_standard":      true,
	"tool":               true,
	"verifier":           true,
	"workflow":           true,
	"workflow_blueprint": true,
}

type Capability struct {
	Kind        string         `json:"kind" yaml:"kind"`
	Name        string         `json:"name" yaml:"name"`
	Version     string         `json:"version" yaml:"version"`
	Description string         `json:"description,omitempty" yaml:"description"`
	Status      string         `json:"status,omitempty" yaml:"status"`
	Path        string         `json:"path" yaml:"-"`
	Raw         map[string]any `json:"raw" yaml:"-"`
}

func (c Capability) ID() string {
	return fmt.Sprintf("%s:%s@%s", c.Kind, c.Name, c.Version)
}

type Registry struct {
	Root         string
	Capabilities []Capability
}

func Load(root string) (*Registry, error) {
	var caps []Capability
	var errs []string

	if _, err := os.Stat(root); err != nil {
		return nil, fmt.Errorf("registry path %q: %w", root, err)
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			errs = append(errs, err.Error())
			return nil
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		cap, err := readCapability(path)
		if err != nil {
			errs = append(errs, err.Error())
			return nil
		}
		caps = append(caps, cap)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(caps, func(i, j int) bool { return caps[i].ID() < caps[j].ID() })
	seen := map[string]string{}
	for _, cap := range caps {
		id := cap.ID()
		if prev, ok := seen[id]; ok {
			errs = append(errs, fmt.Sprintf("duplicate capability %s in %s and %s", id, prev, cap.Path))
			continue
		}
		seen[id] = cap.Path
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("registry validation failed:\n%s", strings.Join(errs, "\n"))
	}
	return &Registry{Root: root, Capabilities: caps}, nil
}

func readCapability(path string) (Capability, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return Capability{}, fmt.Errorf("%s: read: %w", path, err)
	}
	var raw map[string]any
	if err := yaml.Unmarshal(body, &raw); err != nil {
		return Capability{}, fmt.Errorf("%s: yaml: %w", path, err)
	}
	cap := Capability{Path: path, Raw: raw}
	if v, ok := raw["kind"].(string); ok {
		cap.Kind = v
	}
	if v, ok := raw["name"].(string); ok {
		cap.Name = v
	}
	if v, ok := raw["version"].(string); ok {
		cap.Version = v
	}
	if v, ok := raw["description"].(string); ok {
		cap.Description = v
	}
	if v, ok := raw["status"].(string); ok {
		cap.Status = v
	}
	if cap.Version == "" {
		cap.Version = "0.1.0"
	}
	if cap.Status == "" {
		cap.Status = "active"
	}
	if cap.Kind == "" {
		return Capability{}, fmt.Errorf("%s: missing kind", path)
	}
	if !validKinds[cap.Kind] {
		return Capability{}, fmt.Errorf("%s: invalid kind %q", path, cap.Kind)
	}
	if cap.Name == "" {
		return Capability{}, fmt.Errorf("%s: missing name", path)
	}
	if cap.Version == "" {
		return Capability{}, fmt.Errorf("%s: missing version", path)
	}
	return cap, nil
}

func (r *Registry) Find(ref string) (Capability, bool) {
	kind, name, version := parseRef(ref)
	for _, cap := range r.Capabilities {
		if kind != "" && cap.Kind != kind {
			continue
		}
		if cap.Name != name && cap.ID() != ref {
			continue
		}
		if version != "" && cap.Version != version {
			continue
		}
		return cap, true
	}
	return Capability{}, false
}

func parseRef(ref string) (kind, name, version string) {
	left := ref
	if parts := strings.SplitN(ref, "@", 2); len(parts) == 2 {
		left, version = parts[0], parts[1]
	}
	if parts := strings.SplitN(left, ":", 2); len(parts) == 2 {
		kind, name = parts[0], parts[1]
	} else {
		name = left
	}
	return kind, name, version
}

func (c Capability) JSON() ([]byte, error) {
	return json.MarshalIndent(c.Raw, "", "  ")
}

// SyncCapabilities upserts every loaded capability into the relational
// `capabilities` index table. Manifests remain the source of truth on disk, and
// this table is a queryable mirror keyed by (kind, name, version, scope). Returns the
// number of rows written.
func SyncCapabilities(ctx context.Context, db *sql.DB, reg *Registry) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin capability sync: %w", err)
	}
	defer tx.Rollback()

	const upsert = `
INSERT INTO capabilities
  (id, kind, name, version, description, author, source_type, source_uri, source_ref,
   trust_level, scope, status, requirements_json, manifest_path, manifest_json, hash, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  description       = excluded.description,
  author            = excluded.author,
  source_type       = excluded.source_type,
  source_uri        = excluded.source_uri,
  source_ref        = excluded.source_ref,
  trust_level       = excluded.trust_level,
  status            = excluded.status,
  requirements_json = excluded.requirements_json,
  manifest_path     = excluded.manifest_path,
  manifest_json     = excluded.manifest_json,
  hash              = excluded.hash,
  updated_at        = excluded.updated_at;`

	now := time.Now().UTC().Format(time.RFC3339)
	count := 0
	for _, cap := range reg.Capabilities {
		manifestJSON, err := json.Marshal(cap.Raw)
		if err != nil {
			return count, fmt.Errorf("%s: marshal manifest: %w", cap.Path, err)
		}

		// Extract optional fields from manifest.
		var author, sourceURI, sourceRef, trustLevel, scope, requirementsJSON *string
		if v, ok := cap.Raw["author"].(string); ok {
			author = &v
		}
		if src, ok := cap.Raw["source"].(map[string]any); ok {
			if v, ok := src["uri"].(string); ok {
				sourceURI = &v
			}
			if v, ok := src["ref"].(string); ok {
				sourceRef = &v
			}
		}
		if v, ok := cap.Raw["trust_level"].(string); ok {
			trustLevel = &v
		}
		if v, ok := cap.Raw["scope"].(string); ok {
			scope = &v
		}
		if req, ok := cap.Raw["requirements"]; ok {
			b, _ := json.Marshal(req)
			s := string(b)
			requirementsJSON = &s
		}

		if _, err := tx.ExecContext(ctx, upsert,
			cap.ID(), cap.Kind, cap.Name, cap.Version, cap.Description, author,
			"local", sourceURI, sourceRef, trustLevel, scope, cap.Status,
			requirementsJSON, cap.Path, string(manifestJSON), nil, now, now,
		); err != nil {
			return count, fmt.Errorf("%s: upsert capability: %w", cap.ID(), err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return count, fmt.Errorf("commit capability sync: %w", err)
	}
	return count, nil
}
