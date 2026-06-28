package skillindex

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Source struct {
	Name       string         `json:"name" yaml:"name"`
	Type       string         `json:"type" yaml:"type"`
	URI        string         `json:"uri,omitempty" yaml:"uri"`
	Ref        string         `json:"ref,omitempty" yaml:"ref"`
	Scope      string         `json:"scope,omitempty" yaml:"scope"`
	RootPath   string         `json:"root_path,omitempty" yaml:"root_path"`
	TrustLevel string         `json:"trust_level,omitempty" yaml:"trust_level"`
	Version    string         `json:"version,omitempty" yaml:"version"`
	Metadata   map[string]any `json:"metadata,omitempty" yaml:"metadata"`
}

type Skill struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	SourceName  string         `json:"source_name"`
	SourceType  string         `json:"source_type"`
	SourceURI   string         `json:"source_uri,omitempty"`
	SourceRef   string         `json:"source_ref,omitempty"`
	Scope       string         `json:"scope,omitempty"`
	Path        string         `json:"path"`
	Format      string         `json:"format"`
	TrustLevel  string         `json:"trust_level,omitempty"`
	Version     string         `json:"version"`
	Hash        string         `json:"hash"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type frontmatter struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Version     string         `yaml:"version"`
	Metadata    map[string]any `yaml:"metadata"`
}

func IndexFilesystem(src Source) ([]Skill, error) {
	if src.Name == "" {
		return nil, fmt.Errorf("skill source missing name")
	}
	if src.Type == "" {
		src.Type = "filesystem"
	}
	if src.RootPath == "" {
		return nil, fmt.Errorf("skill source %q missing root_path", src.Name)
	}
	if src.Version == "" {
		src.Version = "0.1.0"
	}

	var skills []Skill
	err := filepath.WalkDir(src.RootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}
		skill, err := readSkill(src, path)
		if err != nil {
			return err
		}
		skills = append(skills, skill)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return skills, nil
}

func readSkill(src Source, path string) (Skill, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return Skill{}, fmt.Errorf("read skill %s: %w", path, err)
	}
	fm, err := parseFrontmatter(body)
	if err != nil {
		return Skill{}, fmt.Errorf("parse skill %s: %w", path, err)
	}
	if fm.Name == "" {
		return Skill{}, fmt.Errorf("skill %s missing name", path)
	}
	if fm.Description == "" {
		return Skill{}, fmt.Errorf("skill %s missing description", path)
	}
	version := fm.Version
	if version == "" {
		version = src.Version
	}
	hashBytes := sha256.Sum256(body)
	hash := "sha256:" + hex.EncodeToString(hashBytes[:])
	id := fmt.Sprintf("skill:%s@%s#%s", fm.Name, version, hash[7:19])
	return Skill{
		ID:          id,
		Name:        fm.Name,
		Description: fm.Description,
		SourceName:  src.Name,
		SourceType:  src.Type,
		SourceURI:   src.URI,
		SourceRef:   src.Ref,
		Scope:       src.Scope,
		Path:        path,
		Format:      "agent-skills",
		TrustLevel:  src.TrustLevel,
		Version:     version,
		Hash:        hash,
		Metadata:    fm.Metadata,
	}, nil
}

func parseFrontmatter(body []byte) (frontmatter, error) {
	text := string(body)
	if !strings.HasPrefix(text, "---\n") && !strings.HasPrefix(text, "---\r\n") {
		return frontmatter{}, fmt.Errorf("missing YAML frontmatter")
	}
	text = strings.TrimPrefix(text, "---\r\n")
	text = strings.TrimPrefix(text, "---\n")
	idx := strings.Index(text, "\n---")
	if idx < 0 {
		return frontmatter{}, fmt.Errorf("unterminated YAML frontmatter")
	}
	var fm frontmatter
	if err := yaml.Unmarshal([]byte(text[:idx]), &fm); err != nil {
		return frontmatter{}, err
	}
	return fm, nil
}

func SyncSkills(ctx context.Context, db *sql.DB, src Source, skills []Skill) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin skill sync: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)
	metadataJSON := "{}"
	if len(src.Metadata) > 0 {
		b, err := json.Marshal(src.Metadata)
		if err != nil {
			return 0, fmt.Errorf("marshal source metadata: %w", err)
		}
		metadataJSON = string(b)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO skill_sources
  (id, name, type, uri, ref, scope, trust_level, version, root_path, metadata_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(name, scope) DO UPDATE SET
  type=excluded.type, uri=excluded.uri, ref=excluded.ref, trust_level=excluded.trust_level,
  version=excluded.version, root_path=excluded.root_path, metadata_json=excluded.metadata_json,
  updated_at=excluded.updated_at;`,
		sourceID(src), src.Name, src.Type, src.URI, src.Ref, src.Scope, src.TrustLevel, src.Version, src.RootPath, metadataJSON, now, now,
	); err != nil {
		return 0, fmt.Errorf("upsert skill source: %w", err)
	}

	const upsert = `
INSERT INTO skill_index
  (id, name, description, source_name, source_type, source_uri, source_ref, scope, path, format, trust_level, version, hash, metadata_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(source_name, scope, name, version, path) DO UPDATE SET
  description=excluded.description, source_type=excluded.source_type, source_uri=excluded.source_uri,
  source_ref=excluded.source_ref, format=excluded.format, trust_level=excluded.trust_level,
  hash=excluded.hash, metadata_json=excluded.metadata_json, updated_at=excluded.updated_at;`

	count := 0
	for _, skill := range skills {
		meta := "{}"
		if len(skill.Metadata) > 0 {
			b, err := json.Marshal(skill.Metadata)
			if err != nil {
				return count, fmt.Errorf("%s: marshal metadata: %w", skill.Name, err)
			}
			meta = string(b)
		}
		if _, err := tx.ExecContext(ctx, upsert,
			skill.ID, skill.Name, skill.Description, skill.SourceName, skill.SourceType,
			skill.SourceURI, skill.SourceRef, skill.Scope, skill.Path, skill.Format,
			skill.TrustLevel, skill.Version, skill.Hash, meta, now, now,
		); err != nil {
			return count, fmt.Errorf("%s: upsert skill: %w", skill.Name, err)
		}
		count++
	}
	if err := tx.Commit(); err != nil {
		return count, fmt.Errorf("commit skill sync: %w", err)
	}
	return count, nil
}

func sourceID(src Source) string {
	base := strings.Join([]string{src.Scope, src.Name}, ":")
	h := sha256.Sum256([]byte(base))
	return "skill_source:" + hex.EncodeToString(h[:8])
}
