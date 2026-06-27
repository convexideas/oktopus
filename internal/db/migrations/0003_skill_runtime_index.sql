-- 0003_skill_runtime_index.sql
-- Adds exact-path skill indexing and runtime capability matrix records.

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
  UNIQUE(name, scope)
);

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
  UNIQUE(source_name, scope, name, version, path)
);

CREATE INDEX IF NOT EXISTS idx_skill_index_name ON skill_index(name);
CREATE INDEX IF NOT EXISTS idx_skill_index_scope ON skill_index(scope);

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
  metadata_json TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(runtime, version)
);
