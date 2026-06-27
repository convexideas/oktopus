package runtimeadapter

import (
	"context"
	"time"
)

// Adapter translates a bounded Oktopus job into a runtime-native invocation.
// The controller remains source of truth for registry, leases, artifacts,
// approvals, memory, and audit; adapters only execute leased work.
type Adapter interface {
	RuntimeID() string
	Detect(context.Context) (Detection, error)
	Capabilities(context.Context) (Capabilities, error)
	Prepare(context.Context, JobHandoff) (Invocation, error)
	Execute(context.Context, Lease, Invocation, Observer) (Result, error)
	Cancel(context.Context, Lease) error
}

type Detection struct {
	Available   bool              `json:"available"`
	Runtime     string            `json:"runtime"`
	Version     string            `json:"version,omitempty"`
	ConfigScope string            `json:"config_scope,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Reason      string            `json:"reason,omitempty"`
}

type Capabilities struct {
	Runtime             string         `json:"runtime" yaml:"runtime"`
	Version             string         `json:"version,omitempty" yaml:"version"`
	Skills              bool           `json:"skills" yaml:"skills"`
	MCP                 bool           `json:"mcp" yaml:"mcp"`
	Subagents           bool           `json:"subagents" yaml:"subagents"`
	SlashCommands       bool           `json:"slash_commands" yaml:"slash_commands"`
	ModelOverride       bool           `json:"model_override" yaml:"model_override"`
	StreamingOutput     bool           `json:"streaming_output" yaml:"streaming_output"`
	ArtifactWriteback   string         `json:"artifact_writeback,omitempty" yaml:"artifact_writeback"`
	WorkspaceConfig     bool           `json:"workspace_config" yaml:"workspace_config"`
	GlobalConfig        bool           `json:"global_config" yaml:"global_config"`
	PermissionHooks     string         `json:"permission_hooks,omitempty" yaml:"permission_hooks"`
	NativeApprovals     bool           `json:"native_approvals" yaml:"native_approvals"`
	BackgroundExecution bool           `json:"background_execution" yaml:"background_execution"`
	InteractiveMode     bool           `json:"interactive_mode" yaml:"interactive_mode"`
	Metadata            map[string]any `json:"metadata,omitempty" yaml:"metadata"`
}

type JobHandoff struct {
	RunID             string            `json:"run_id"`
	JobID             string            `json:"job_id"`
	AttemptID         string            `json:"attempt_id,omitempty"`
	Kind              string            `json:"kind"`
	RuntimeRef        string            `json:"runtime_ref,omitempty"`
	PersonaRef        string            `json:"persona_ref,omitempty"`
	SkillRefs         []SkillRef        `json:"skill_refs,omitempty"`
	AllowedTools      []CapabilityRef   `json:"allowed_tools,omitempty"`
	InputArtifactRefs []ArtifactRef     `json:"input_artifact_refs,omitempty"`
	ContextBundle     ContextBundle     `json:"context_bundle,omitempty"`
	PolicyConstraints map[string]any    `json:"policy_constraints,omitempty"`
	OutputContract    map[string]any    `json:"output_contract,omitempty"`
	EvidenceContract  map[string]any    `json:"evidence_contract,omitempty"`
	Environment       map[string]string `json:"environment,omitempty"`
	Metadata          map[string]any    `json:"metadata,omitempty"`
}

type SkillRef struct {
	Name     string            `json:"name"`
	Source   string            `json:"source"`
	Path     string            `json:"path"`
	Hash     string            `json:"hash"`
	Format   string            `json:"format"`
	Version  string            `json:"version"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type CapabilityRef struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Source   string `json:"source,omitempty"`
	Revision string `json:"revision,omitempty"`
}

type ArtifactRef struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	URI    string `json:"uri"`
	SHA256 string `json:"sha256,omitempty"`
}

type ContextBundle struct {
	Summary   string        `json:"summary,omitempty"`
	Snippets  []Snippet     `json:"snippets,omitempty"`
	Artifacts []ArtifactRef `json:"artifacts,omitempty"`
	Budget    ContextBudget `json:"budget,omitempty"`
}

type Snippet struct {
	Source string `json:"source"`
	Text   string `json:"text"`
}

type ContextBudget struct {
	MaxChars        int `json:"max_chars,omitempty"`
	MaxSkills       int `json:"max_skills,omitempty"`
	MaxTools        int `json:"max_tools,omitempty"`
	MaxArtifactRefs int `json:"max_artifact_refs,omitempty"`
}

type Invocation struct {
	Runtime string            `json:"runtime"`
	Command []string          `json:"command,omitempty"`
	Prompt  string            `json:"prompt,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	WorkDir string            `json:"work_dir,omitempty"`
	Files   map[string]string `json:"files,omitempty"`
	Meta    map[string]any    `json:"meta,omitempty"`
}

type Lease struct {
	ID        string    `json:"id"`
	RunID     string    `json:"run_id"`
	JobID     string    `json:"job_id"`
	AttemptID string    `json:"attempt_id"`
	WorkerID  string    `json:"worker_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Result struct {
	Status       string         `json:"status"`
	Summary      string         `json:"summary,omitempty"`
	ExitCode     int            `json:"exit_code,omitempty"`
	Artifacts    []ArtifactRef  `json:"artifacts,omitempty"`
	VerifierRefs []ArtifactRef  `json:"verifier_refs,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

type Observer interface {
	Event(context.Context, Event) error
	Log(context.Context, string) error
	Artifact(context.Context, ArtifactRef) error
}

type Event struct {
	Type     string         `json:"type"`
	Severity string         `json:"severity,omitempty"`
	Message  string         `json:"message,omitempty"`
	Payload  map[string]any `json:"payload,omitempty"`
}
