-- Oktopus MVP schema (flat, no incremental migrations during development)

-- Profiles: named local/user identity for defaults and preferences.
CREATE TABLE IF NOT EXISTS profiles (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

-- Workspaces: immutable shared definitions of what a coding environment looks like.
-- Change by forking, not mutating.
CREATE TABLE IF NOT EXISTS workspaces (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL
);

-- Workspace refs: normalized references that compose a workspace definition.
-- kind: source, tool, knowledge, environment, resource, io_connector, etc.
-- slot: disambiguation when multiple refs of same kind exist.
CREATE TABLE IF NOT EXISTS workspace_refs (
  workspace_id TEXT NOT NULL,
  kind TEXT NOT NULL,
  slot TEXT NOT NULL DEFAULT '',
  ref TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (workspace_id, kind, slot, ref),
  FOREIGN KEY (workspace_id) REFERENCES workspaces(id)
);

-- Sandbox definitions: immutable templates for what kind of sandbox to create.
CREATE TABLE IF NOT EXISTS sandbox_defs (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  provider TEXT NOT NULL,
  base_image_ref TEXT,
  created_at TEXT NOT NULL
);

-- Sandbox instances: user-bound running copies of a sandbox definition.
-- Records what was actually provisioned after policy/resource resolution.
CREATE TABLE IF NOT EXISTS sandbox_instances (
  id TEXT PRIMARY KEY,
  sandbox_def_id TEXT NOT NULL,
  workspace_id TEXT NOT NULL,
  owner_subject TEXT NOT NULL,
  execution_principal TEXT,
  resource_limits_json TEXT,
  grants_json TEXT,
  environment_json TEXT,
  provider_sandbox_id TEXT,
  status TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  last_seen_at TEXT,
  FOREIGN KEY (sandbox_def_id) REFERENCES sandbox_defs(id),
  FOREIGN KEY (workspace_id) REFERENCES workspaces(id)
);

-- Sessions: append-only task capsules. Creation = start.
CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY,
  sandbox_instance_id TEXT NOT NULL,
  title TEXT NOT NULL,
  status TEXT NOT NULL,
  created_by TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (sandbox_instance_id) REFERENCES sandbox_instances(id)
);

-- Session events: append-only timeline of what happened in a session.
CREATE TABLE IF NOT EXISTS session_events (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL,
  type TEXT NOT NULL,
  actor_type TEXT,
  actor_id TEXT,
  message TEXT,
  payload_json TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY (session_id) REFERENCES sessions(id)
);

-- Session artifacts: outputs produced during a session.
CREATE TABLE IF NOT EXISTS session_artifacts (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL,
  type TEXT NOT NULL,
  name TEXT NOT NULL,
  uri TEXT NOT NULL,
  media_type TEXT,
  size_bytes INTEGER,
  sha256 TEXT,
  producer_ref TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY (session_id) REFERENCES sessions(id)
);

-- Capabilities: registry index (loaded from YAML manifests).
CREATE TABLE IF NOT EXISTS capabilities (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  name TEXT NOT NULL,
  version TEXT NOT NULL,
  source_json TEXT,
  status TEXT NOT NULL,
  manifest_path TEXT,
  manifest_json TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (kind, name, version)
);

-- Skill sources: where skills are imported from.
CREATE TABLE IF NOT EXISTS skill_sources (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  uri TEXT,
  ref TEXT,
  scope TEXT,
  trust_level TEXT,
  version TEXT,
  root_path TEXT,
  metadata_json TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (name, scope)
);

-- Skill index: exact-path skill records.
CREATE TABLE IF NOT EXISTS skill_index (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  source_name TEXT NOT NULL,
  source_type TEXT NOT NULL,
  source_uri TEXT,
  source_ref TEXT,
  scope TEXT,
  path TEXT NOT NULL,
  format TEXT NOT NULL,
  trust_level TEXT,
  version TEXT NOT NULL,
  hash TEXT NOT NULL,
  metadata_json TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (source_name, scope, name, version, path)
);

-- Runtime capabilities: what each runtime supports.
CREATE TABLE IF NOT EXISTS runtime_capabilities (
  id TEXT PRIMARY KEY,
  runtime TEXT NOT NULL,
  version TEXT,
  skills INTEGER NOT NULL DEFAULT 0,
  mcp INTEGER NOT NULL DEFAULT 0,
  subagents INTEGER NOT NULL DEFAULT 0,
  slash_commands INTEGER NOT NULL DEFAULT 0,
  model_override INTEGER NOT NULL DEFAULT 0,
  streaming_output INTEGER NOT NULL DEFAULT 0,
  artifact_writeback TEXT,
  workspace_config INTEGER NOT NULL DEFAULT 0,
  global_config INTEGER NOT NULL DEFAULT 0,
  permission_hooks TEXT,
  native_approvals INTEGER NOT NULL DEFAULT 0,
  background_execution INTEGER NOT NULL DEFAULT 0,
  interactive_mode INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (runtime, version)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sandbox_instances_workspace ON sandbox_instances(workspace_id);
CREATE INDEX IF NOT EXISTS idx_sandbox_instances_owner ON sandbox_instances(owner_subject, status);
CREATE INDEX IF NOT EXISTS idx_sessions_sandbox ON sessions(sandbox_instance_id);
CREATE INDEX IF NOT EXISTS idx_session_events_session ON session_events(session_id, created_at);
CREATE INDEX IF NOT EXISTS idx_session_artifacts_session ON session_artifacts(session_id);
CREATE INDEX IF NOT EXISTS idx_skill_index_name ON skill_index(name);
CREATE INDEX IF NOT EXISTS idx_skill_index_scope ON skill_index(scope);
