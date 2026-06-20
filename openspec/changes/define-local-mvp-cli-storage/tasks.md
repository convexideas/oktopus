## Tasks

### Specification

- [x] Define local MVP repository layout.
- [x] Define Go binary and CLI command surface.
- [x] Define SQLite tables for core state.
- [x] Define migration approach.
- [x] Define registry loading and validation behavior.
- [x] Define run creation and workflow expansion behavior.
- [x] Define state transition helper requirement.
- [x] Define local worker MVP behavior.
- [x] Define seed capabilities and starter workflows.
- [x] Define test strategy and acceptance criteria.
- [x] Decide TOML-first vs YAML-first manifest parser for MVP: YAML-first, with TOML compatibility for early seed manifests if needed.
- [x] Decide exact Go CLI library or stdlib-only CLI: Cobra.
- [x] Add `project_counters` table for race-free sequential run number generation.
- [x] Define project auto-creation rules for `oktopus runs create`.
- [x] Add `retry_policy_json` column to `jobs` table.
- [x] Add `attempt_id` to `approvals` and `policy_decisions` tables.
- [x] Add `oktopus registry propose` and `oktopus registry activate` to CLI command surface.
- [x] Add `oktopus jobs retry` to CLI command surface.
- [x] Add `ExpireAttempt` and `BlockJob` and `CarryForwardApproval` to state transition helpers.
- [x] Validate OpenSpec with `openspec validate --all`.

### Implementation Follow-up

- [x] Replace/remove early Python prototype or quarantine it as historical spike.
- [x] Create Go module and `cmd/oktopus` entrypoint.
- [x] Add Cobra CLI with `version`, `db init`, `db migrate`, `registry validate`, `registry list`, and `registry show`.
- [x] Add migration runner and initial SQLite schema.
- [x] Add registry loader for seed manifests.
- [x] Add `project_counters` table to migration (0002_policy_retries.sql).
- [x] Add `retry_policy_json` column to `jobs` table in migration (0002_policy_retries.sql).
- [x] Add `attempt_id` column to `approvals` table in migration (0002); `policy_decisions.attempt_id` already in 0001.
- [ ] Add `output_kind` column to `jobs` table in migration.
- [x] Differentiate `db init` (create + apply all) from `db migrate` (apply pending, error if uninitialized).
- [x] Write capability SQLite index from `registry validate` (with `--no-index` for file-only validation).
- [ ] Add run creation that expands workflow steps into jobs with project auto-creation.
- [ ] Add transactional state transition helpers with event emission.
- [ ] Add local artifact store with sha256 finalization.
- [ ] Add local worker registration/request/lease/heartbeat/complete flow.
- [ ] Add `artifact-exists` and `command-exit-zero` verifiers.
- [ ] Add `registry propose` and `registry activate` commands.
- [ ] Add `jobs retry` command.
- [x] Add `hello-local` workflow.
- [x] Add `guarded-build` stub workflow with adversarial review jobs.
- [ ] Add smoke integration test.
