-- 0002_policy_retries.sql
-- Adds spine fields from the adversarial review resolutions:
--   * project_counters: race-free sequential run numbering per project
--   * jobs.retry_policy_json: {"mode": "idempotent|at-most-once|manual"}
--   * approvals.attempt_id: approvals are valid for the attempt that requested them
-- (policy_decisions.attempt_id already exists in 0001_init.sql)

CREATE TABLE IF NOT EXISTS project_counters (
  project_id TEXT PRIMARY KEY,
  run_count  INTEGER NOT NULL DEFAULT 0,
  FOREIGN KEY(project_id) REFERENCES projects(id)
);

ALTER TABLE jobs ADD COLUMN retry_policy_json TEXT;

ALTER TABLE approvals ADD COLUMN attempt_id TEXT;
