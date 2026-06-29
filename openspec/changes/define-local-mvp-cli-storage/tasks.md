## Tasks

### Specification

- [x] Define local MVP repository layout.
- [x] Define Go binary and CLI command surface.
- [x] Define installer PATH behavior for the CLI.
- [x] Define storage boundary and migration approach.
- [x] Define registry loading and validation behavior.
- [x] Reduce MVP scope to coding sessions only.
- [x] Define Profile → Workspace → Sandbox → Session model.
- [x] Simplify MVP data model to profiles, reusable workspace templates, sandbox instances, sessions, events, and artifacts.
- [x] Define workspace materialization for sandbox startup.
- [x] Define OpenShell as the first local sandbox provider.
- [x] Define that secrets stay out of materialized workspace files/staging.
- [x] Defer workflows/runs/jobs/workers/verifiers until session MVP works.
- [x] Decide TOML-first vs YAML-first manifest parser for MVP: YAML-first.
- [x] Decide exact Go CLI library or stdlib-only CLI: Cobra.
- [x] Validate OpenSpec with `openspec validate --all`.

### Implementation Follow-up

- [x] Replace/remove early Python prototype or quarantine it as historical spike.
- [x] Create Go module and `cmd/oktopus` entrypoint.
- [x] Add Cobra CLI with `version`, `db init`, `db migrate`, `registry validate`, `registry list`, and `registry show`.
- [x] Add migration runner and initial SQLite schema.
- [x] Add store boundary over SQLite local adapter.
- [x] Add registry loader for seed manifests.
- [x] Quarantine non-MVP registry artifacts.
- [x] Add `profiles`, `workspaces`, `sessions`, `session_events`, and `session_artifacts` migration.
- [x] Add `profile init local` command.
- [x] Add immutable workspace fields and `sandbox_instances` migration.
- [x] Simplify profile/workspace/session/sandbox tables for MVP-only fields.
- [ ] Add `workspace create/list/show/fork` commands.
- [ ] Add `session start/list/show/log` commands.
- [ ] Add workspace materialization for OpenShell startup.
- [ ] Add generated session `AGENTS.md`.
- [ ] Add local session artifact metadata with sha256 finalization.
- [ ] Add OpenShell provider wrapper for sandbox create/connect/exec/logs.
- [ ] Add `session attach` through OpenShell connect/TTY.
- [ ] Add installer/package-manager setup so `oktopus` is available on `PATH`.
- [ ] Add smoke test for profile → workspace → session → sandbox preparation.

### Deferred

- [ ] Run creation and workflow expansion.
- [ ] Transactional job state transition helpers.
- [ ] Worker registration/request/lease/heartbeat/complete flow.
- [ ] Verifier execution.
- [ ] `registry propose` and `registry activate` commands.
- [ ] `jobs retry` command.
