-- Oktopus schema (greenfield — drop and recreate during development)

-- Users: identity records. Profiles (local config + tokens) live on disk, not in DB.
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT,
  external_id TEXT,
  status TEXT NOT NULL,
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
  policy_ref TEXT,
  created_at TEXT NOT NULL
);

-- Sandbox instances: user-bound running copies of a sandbox definition.
CREATE TABLE IF NOT EXISTS sandbox_instances (
  id TEXT PRIMARY KEY,
  sandbox_def_id TEXT NOT NULL,
  workspace_id TEXT NOT NULL,
  owner TEXT NOT NULL,
  run_as TEXT,
  provider_sandbox_id TEXT,
  policy_snapshot_json TEXT,
  status TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  last_seen_at TEXT,
  FOREIGN KEY (sandbox_def_id) REFERENCES sandbox_defs(id),
  FOREIGN KEY (workspace_id) REFERENCES workspaces(id)
);

-- Sessions: append-only task capsules. Creation = start.
-- Bare sessions: sandbox_instance_id is optional (empty string for bare sessions).
-- agent + workspace allow sessions without the full sandbox chain.
CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY,
  sandbox_instance_id TEXT NOT NULL DEFAULT '',
  agent TEXT NOT NULL DEFAULT '',
  workspace TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL,
  status TEXT NOT NULL,
  created_by TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

-- Messages: the conversation record within a session.
-- Each message has a role, optional parent (for branching/lineage), and ordered parts.
CREATE TABLE IF NOT EXISTS messages (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL,
  parent_id TEXT,
  role TEXT NOT NULL,
  sequence INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY (session_id) REFERENCES sessions(id),
  FOREIGN KEY (parent_id) REFERENCES messages(id)
);

-- Message parts: ordered content blocks within a message.
-- A single message may contain text + tool_call + reasoning interleaved.
CREATE TABLE IF NOT EXISTS message_parts (
  id TEXT PRIMARY KEY,
  message_id TEXT NOT NULL,
  sequence INTEGER NOT NULL,
  type TEXT NOT NULL,
  content_json TEXT NOT NULL,
  FOREIGN KEY (message_id) REFERENCES messages(id)
);

-- Message attachments: objects associated with a message (inputs or outputs).
CREATE TABLE IF NOT EXISTS message_attachments (
  id TEXT PRIMARY KEY,
  message_id TEXT NOT NULL,
  direction TEXT NOT NULL,
  name TEXT NOT NULL,
  uri TEXT NOT NULL,
  media_type TEXT,
  size_bytes INTEGER,
  sha256 TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY (message_id) REFERENCES messages(id)
);

-- Capabilities: unified registry index. All capability kinds in one table.
-- Skills, tools, workflows, personas, runtimes, verifiers, adapters, policies, environments.
CREATE TABLE IF NOT EXISTS capabilities (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  name TEXT NOT NULL,
  version TEXT NOT NULL,
  description TEXT,
  author TEXT,
  source_type TEXT NOT NULL,
  source_uri TEXT,
  source_ref TEXT,
  trust_level TEXT,
  scope TEXT,
  status TEXT NOT NULL,
  requirements_json TEXT,
  manifest_path TEXT,
  manifest_json TEXT NOT NULL,
  hash TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (kind, name, version, scope)
);

-- Episodes: raw captures from sessions. Harness-neutral.
CREATE TABLE IF NOT EXISTS episodes (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL DEFAULT '',
  workspace_id TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL,
  content TEXT NOT NULL,
  captured_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_episodes_session ON episodes(session_id);
CREATE INDEX IF NOT EXISTS idx_episodes_workspace ON episodes(workspace_id, captured_at);
CREATE INDEX IF NOT EXISTS idx_episodes_source ON episodes(workspace_id, source, captured_at);

-- Profiles: user runtime identity. One active profile per user.
CREATE TABLE IF NOT EXISTS profiles (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL DEFAULT 'local',
  default_model TEXT NOT NULL DEFAULT '',
  default_runtime TEXT NOT NULL DEFAULT '',
  config_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_user ON profiles(user_id);

-- Captures: metadata for IO stream recording sessions.
CREATE TABLE IF NOT EXISTS captures (
  id TEXT PRIMARY KEY,
  session_id TEXT,
  agent TEXT NOT NULL,
  started_at TEXT NOT NULL,
  ended_at TEXT,
  status TEXT NOT NULL DEFAULT 'running',
  log_path TEXT NOT NULL,
  FOREIGN KEY (session_id) REFERENCES sessions(id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_capabilities_kind ON capabilities(kind, status);
CREATE INDEX IF NOT EXISTS idx_capabilities_author ON capabilities(author);
CREATE INDEX IF NOT EXISTS idx_capabilities_scope ON capabilities(scope, kind);
CREATE INDEX IF NOT EXISTS idx_sandbox_instances_workspace ON sandbox_instances(workspace_id);
CREATE INDEX IF NOT EXISTS idx_sandbox_instances_owner ON sandbox_instances(owner, status);
CREATE INDEX IF NOT EXISTS idx_sessions_agent ON sessions(agent, created_at);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id, sequence);
CREATE INDEX IF NOT EXISTS idx_messages_parent ON messages(parent_id);
CREATE INDEX IF NOT EXISTS idx_message_parts_message ON message_parts(message_id, sequence);
CREATE INDEX IF NOT EXISTS idx_message_attachments_message ON message_attachments(message_id);
CREATE INDEX IF NOT EXISTS idx_captures_session ON captures(session_id);
CREATE INDEX IF NOT EXISTS idx_captures_agent ON captures(agent, started_at);
