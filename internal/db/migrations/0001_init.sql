CREATE TABLE IF NOT EXISTS projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  root_uri TEXT,
  metadata_json TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

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
  UNIQUE(kind, name, version)
);

CREATE TABLE IF NOT EXISTS runs (
  id TEXT PRIMARY KEY,
  project_id TEXT,
  number INTEGER,
  status TEXT NOT NULL,
  title TEXT NOT NULL,
  intent TEXT,
  workflow_ref TEXT NOT NULL,
  workflow_version TEXT,
  trigger_type TEXT,
  trigger_payload_json TEXT,
  resolved_capabilities_json TEXT,
  created_by TEXT,
  parent_run_id TEXT,
  created_at TEXT NOT NULL,
  started_at TEXT,
  finished_at TEXT,
  metadata_json TEXT,
  FOREIGN KEY(project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS jobs (
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL,
  key TEXT NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL,
  status TEXT NOT NULL,
  runtime_ref TEXT,
  capability_refs_json TEXT,
  depends_on_json TEXT,
  priority INTEGER DEFAULT 0,
  max_attempts INTEGER DEFAULT 1,
  timeout_seconds INTEGER,
  input_artifacts_json TEXT,
  output_contract_json TEXT,
  policy_ref TEXT,
  environment_ref TEXT,
  created_at TEXT NOT NULL,
  started_at TEXT,
  finished_at TEXT,
  metadata_json TEXT,
  UNIQUE(run_id, key),
  FOREIGN KEY(run_id) REFERENCES runs(id)
);

CREATE TABLE IF NOT EXISTS attempts (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL,
  attempt_number INTEGER NOT NULL,
  worker_id TEXT,
  status TEXT NOT NULL,
  started_at TEXT,
  finished_at TEXT,
  exit_code INTEGER,
  error_code TEXT,
  error_message TEXT,
  summary TEXT,
  metadata_json TEXT,
  UNIQUE(job_id, attempt_number),
  FOREIGN KEY(job_id) REFERENCES jobs(id)
);

CREATE TABLE IF NOT EXISTS workers (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  kind TEXT NOT NULL,
  version TEXT,
  status TEXT NOT NULL,
  labels_json TEXT,
  platform_json TEXT,
  capabilities_json TEXT,
  max_concurrency INTEGER DEFAULT 1,
  current_jobs INTEGER DEFAULT 0,
  last_seen_at TEXT,
  started_at TEXT,
  metadata_json TEXT
);

CREATE TABLE IF NOT EXISTS leases (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL,
  attempt_id TEXT NOT NULL,
  worker_id TEXT NOT NULL,
  status TEXT NOT NULL,
  leased_at TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  heartbeat_at TEXT,
  released_at TEXT,
  release_reason TEXT,
  FOREIGN KEY(job_id) REFERENCES jobs(id),
  FOREIGN KEY(attempt_id) REFERENCES attempts(id),
  FOREIGN KEY(worker_id) REFERENCES workers(id)
);

CREATE INDEX IF NOT EXISTS idx_leases_active_job ON leases(job_id, status);

CREATE TABLE IF NOT EXISTS steps (
  id TEXT PRIMARY KEY,
  attempt_id TEXT NOT NULL,
  sequence INTEGER NOT NULL,
  kind TEXT NOT NULL,
  name TEXT NOT NULL,
  status TEXT NOT NULL,
  started_at TEXT,
  finished_at TEXT,
  input_ref TEXT,
  output_ref TEXT,
  error_message TEXT,
  metadata_json TEXT,
  FOREIGN KEY(attempt_id) REFERENCES attempts(id)
);

CREATE TABLE IF NOT EXISTS artifacts (
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL,
  job_id TEXT,
  attempt_id TEXT,
  step_id TEXT,
  type TEXT NOT NULL,
  name TEXT NOT NULL,
  uri TEXT NOT NULL,
  media_type TEXT,
  size_bytes INTEGER,
  sha256 TEXT,
  status TEXT NOT NULL,
  producer_ref TEXT,
  provenance_json TEXT,
  retention_policy TEXT,
  redaction_level TEXT,
  created_at TEXT NOT NULL,
  finalized_at TEXT,
  metadata_json TEXT,
  FOREIGN KEY(run_id) REFERENCES runs(id),
  FOREIGN KEY(job_id) REFERENCES jobs(id),
  FOREIGN KEY(attempt_id) REFERENCES attempts(id),
  FOREIGN KEY(step_id) REFERENCES steps(id)
);

CREATE TABLE IF NOT EXISTS events (
  id TEXT PRIMARY KEY,
  run_id TEXT,
  job_id TEXT,
  attempt_id TEXT,
  step_id TEXT,
  type TEXT NOT NULL,
  severity TEXT NOT NULL,
  message TEXT,
  payload_json TEXT,
  actor_type TEXT,
  actor_id TEXT,
  timestamp TEXT NOT NULL,
  correlation_id TEXT,
  causation_id TEXT,
  FOREIGN KEY(run_id) REFERENCES runs(id),
  FOREIGN KEY(job_id) REFERENCES jobs(id),
  FOREIGN KEY(attempt_id) REFERENCES attempts(id),
  FOREIGN KEY(step_id) REFERENCES steps(id)
);

CREATE TABLE IF NOT EXISTS verifier_results (
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL,
  job_id TEXT,
  attempt_id TEXT,
  verifier_ref TEXT NOT NULL,
  status TEXT NOT NULL,
  summary TEXT,
  started_at TEXT,
  finished_at TEXT,
  input_artifact_ids_json TEXT,
  output_artifact_id TEXT,
  exit_code INTEGER,
  error_message TEXT,
  metadata_json TEXT,
  FOREIGN KEY(run_id) REFERENCES runs(id),
  FOREIGN KEY(job_id) REFERENCES jobs(id),
  FOREIGN KEY(attempt_id) REFERENCES attempts(id),
  FOREIGN KEY(output_artifact_id) REFERENCES artifacts(id)
);

CREATE TABLE IF NOT EXISTS policy_decisions (
  id TEXT PRIMARY KEY,
  run_id TEXT,
  job_id TEXT,
  attempt_id TEXT,
  step_id TEXT,
  policy_ref TEXT,
  subject_type TEXT,
  subject_id TEXT,
  action TEXT NOT NULL,
  resource TEXT,
  outcome TEXT NOT NULL,
  reason TEXT,
  requires_approval INTEGER NOT NULL DEFAULT 0,
  approval_id TEXT,
  created_at TEXT NOT NULL,
  metadata_json TEXT
);

CREATE TABLE IF NOT EXISTS approvals (
  id TEXT PRIMARY KEY,
  run_id TEXT,
  job_id TEXT,
  step_id TEXT,
  type TEXT NOT NULL,
  status TEXT NOT NULL,
  requested_by TEXT,
  requested_at TEXT NOT NULL,
  resolved_by TEXT,
  resolved_at TEXT,
  expires_at TEXT,
  reason TEXT,
  payload_json TEXT
);
