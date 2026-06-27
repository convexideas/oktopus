-- 0005_workspace_sandbox_model.sql
-- Refines MVP semantics:
--   * profile fields are authoring defaults, not Oktopus storage settings
--   * workspaces are immutable shared definitions
--   * sandboxes are user-bound private runtime instances

ALTER TABLE profiles ADD COLUMN scope TEXT NOT NULL DEFAULT 'user';
ALTER TABLE profiles ADD COLUMN policy_ref TEXT;
ALTER TABLE profiles ADD COLUMN default_sandbox_provider TEXT;
ALTER TABLE profiles ADD COLUMN default_sandbox_profile TEXT;
ALTER TABLE profiles ADD COLUMN default_agent_ref TEXT;
ALTER TABLE profiles ADD COLUMN default_model_ref TEXT;
ALTER TABLE profiles ADD COLUMN allowed_provider_refs_json TEXT;
ALTER TABLE profiles ADD COLUMN budget_ref TEXT;

ALTER TABLE workspaces ADD COLUMN scope TEXT NOT NULL DEFAULT 'org';
ALTER TABLE workspaces ADD COLUMN version TEXT NOT NULL DEFAULT '0.1.0';
ALTER TABLE workspaces ADD COLUMN parent_workspace_id TEXT;
ALTER TABLE workspaces ADD COLUMN definition_hash TEXT;
ALTER TABLE workspaces ADD COLUMN source_type TEXT;
ALTER TABLE workspaces ADD COLUMN source_uri TEXT;
ALTER TABLE workspaces ADD COLUMN source_ref TEXT;
ALTER TABLE workspaces ADD COLUMN base_image_ref TEXT;
ALTER TABLE workspaces ADD COLUMN environment_ref TEXT;
ALTER TABLE workspaces ADD COLUMN policy_ref TEXT;
ALTER TABLE workspaces ADD COLUMN resource_refs_json TEXT;
ALTER TABLE workspaces ADD COLUMN tool_refs_json TEXT;
ALTER TABLE workspaces ADD COLUMN knowledge_refs_json TEXT;
ALTER TABLE workspaces ADD COLUMN io_connector_refs_json TEXT;
ALTER TABLE workspaces ADD COLUMN created_by TEXT;

CREATE TABLE IF NOT EXISTS sandbox_instances (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL,
  profile_id TEXT,
  session_id TEXT,
  owner_subject TEXT NOT NULL,
  provider TEXT NOT NULL,
  provider_sandbox_id TEXT,
  status TEXT NOT NULL,
  credential_scope TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  last_seen_at TEXT,
  metadata_json TEXT,
  FOREIGN KEY(workspace_id) REFERENCES workspaces(id),
  FOREIGN KEY(profile_id) REFERENCES profiles(id),
  FOREIGN KEY(session_id) REFERENCES sessions(id)
);

CREATE INDEX IF NOT EXISTS idx_sandbox_instances_workspace ON sandbox_instances(workspace_id);
CREATE INDEX IF NOT EXISTS idx_sandbox_instances_owner ON sandbox_instances(owner_subject, status);
CREATE INDEX IF NOT EXISTS idx_sandbox_instances_session ON sandbox_instances(session_id);
