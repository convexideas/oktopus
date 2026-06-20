## Design

### Goal

Build the smallest local Oktopus that proves the platform spine.

The MVP must support:

- local capability registry validation
- SQLite state store
- run creation from a workflow manifest
- job expansion from workflow steps
- event emission
- artifact directory creation
- local worker lease protocol
- simple shell/verifier execution
- adversarial review workflow shape, even if reviewer execution is stubbed first

### Repository Layout

```text
oktopus/
  cmd/oktopus/                 # Go CLI entrypoint
  internal/
    cli/                       # command wiring
    db/                        # SQLite connection + migrations
    registry/                  # manifest loading + validation
    runs/                      # run/job state transitions
    scheduler/                 # eligible job queries + leases
    worker/                    # local worker loop
    artifacts/                 # filesystem artifact store
    events/                    # event append helpers
    verifier/                  # built-in verifiers
    policy/                    # MVP policy decisions
  migrations/                  # SQL migrations
  registry/                    # local capability manifests
  runs/                        # local run artifacts/logs/events mirror
  openspec/                    # OpenSpec roadmap
```

### Runtime and Packaging

Use Go for core implementation.

Use Cobra for CLI command structure because Oktopus will have nested command groups and scriptable JSON output.

MVP commands run as one binary:

```text
oktopus
```

Local state defaults:

```text
.oktopus/oktopus.db      # SQLite DB, if running inside project
runs/                    # run artifacts/logs when in repository root
registry/                # local manifests
```

CLI flags:

```text
--db <path>
--registry <path>
--runs-dir <path>
--json
--verbose
```

Default path resolution:

1. explicit flag
2. `OKTOPUS_*` env var
3. current repo `oktopus/` paths
4. user config later

### Initial CLI Commands

```text
oktopus version
oktopus db init
oktopus db migrate
oktopus registry validate
oktopus registry list
oktopus registry show <kind:name[@version]>
oktopus registry propose --from <artifact-path> [--kind <kind>] [--name <name>]
oktopus registry activate <kind:name[@version]> --approval <approval-id>
oktopus runs create <workflow> --title <title> [--project <project>] [--intent <text>]
oktopus runs list
oktopus runs show <run-id>
oktopus jobs list --run <run-id>
oktopus jobs retry <job-id>
oktopus events list --run <run-id>
oktopus events tail --run <run-id>
oktopus artifacts list --run <run-id>
oktopus worker local --kind <kind> [--once]
oktopus verifier run <verifier> --run <run-id> [args]
```

Use plural nouns for run/job/artifact/event commands.

### SQLite Tables

Initial tables:

```text
schema_migrations
projects
project_counters
capabilities
runs
jobs
attempts
leases
workers
steps
artifacts
events
verifier_results
policy_decisions
approvals
```

#### project_counters

```text
project_id TEXT PRIMARY KEY
run_count  INTEGER NOT NULL DEFAULT 0
FOREIGN KEY(project_id) REFERENCES projects(id)
```

Used for sequential run number generation without race conditions. Run creation atomically does:

```sql
UPDATE project_counters SET run_count = run_count + 1 WHERE project_id = ?;
SELECT run_count FROM project_counters WHERE project_id = ?;
```

Both statements execute in the same transaction that creates the run. `SetMaxOpenConns(1)` on the SQLite connection serializes this correctly. In Postgres, replace with a sequence.

#### projects

```text
id TEXT PRIMARY KEY
name TEXT NOT NULL
slug TEXT UNIQUE NOT NULL
root_uri TEXT
metadata_json TEXT
created_at TEXT NOT NULL
updated_at TEXT NOT NULL
```

#### capabilities

```text
id TEXT PRIMARY KEY
kind TEXT NOT NULL
name TEXT NOT NULL
version TEXT NOT NULL
source_json TEXT
status TEXT NOT NULL
manifest_path TEXT
manifest_json TEXT NOT NULL
created_at TEXT NOT NULL
updated_at TEXT NOT NULL
UNIQUE(kind, name, version)
```

#### runs

```text
id TEXT PRIMARY KEY
project_id TEXT
number INTEGER
status TEXT NOT NULL
title TEXT NOT NULL
intent TEXT
workflow_ref TEXT NOT NULL
workflow_version TEXT
trigger_type TEXT
trigger_payload_json TEXT
resolved_capabilities_json TEXT
created_by TEXT
parent_run_id TEXT
created_at TEXT NOT NULL
started_at TEXT
finished_at TEXT
metadata_json TEXT
```

#### jobs

```text
id TEXT PRIMARY KEY
run_id TEXT NOT NULL
key TEXT NOT NULL
name TEXT NOT NULL
kind TEXT NOT NULL
status TEXT NOT NULL
runtime_ref TEXT
capability_refs_json TEXT
depends_on_json TEXT
priority INTEGER DEFAULT 0
max_attempts INTEGER DEFAULT 1
retry_policy_json TEXT
timeout_seconds INTEGER
input_artifacts_json TEXT
output_contract_json TEXT
output_kind TEXT
policy_ref TEXT
environment_ref TEXT
created_at TEXT NOT NULL
started_at TEXT
finished_at TEXT
metadata_json TEXT
UNIQUE(run_id, key)
```

`retry_policy_json` stores `{"mode": "idempotent|at-most-once|manual"}`. Defaults to `idempotent` when null.

`output_kind` stores `artifact | stream | effect`; defaults to `artifact` when null. It selects how the controller evaluates the job's evidence contract at completion (artifact existence, stream delivery, or effect consequence verifier).

#### attempts

```text
id TEXT PRIMARY KEY
job_id TEXT NOT NULL
attempt_number INTEGER NOT NULL
worker_id TEXT
status TEXT NOT NULL
started_at TEXT
finished_at TEXT
exit_code INTEGER
error_code TEXT
error_message TEXT
summary TEXT
metadata_json TEXT
UNIQUE(job_id, attempt_number)
```

#### leases

```text
id TEXT PRIMARY KEY
job_id TEXT NOT NULL
attempt_id TEXT NOT NULL
worker_id TEXT NOT NULL
status TEXT NOT NULL
leased_at TEXT NOT NULL
expires_at TEXT NOT NULL
heartbeat_at TEXT
released_at TEXT
release_reason TEXT
```

MVP enforces one active lease per job in application code and transaction. Later Postgres can use partial unique index.

#### events

```text
id TEXT PRIMARY KEY
run_id TEXT
job_id TEXT
attempt_id TEXT
step_id TEXT
type TEXT NOT NULL
severity TEXT NOT NULL
message TEXT
payload_json TEXT
actor_type TEXT
actor_id TEXT
timestamp TEXT NOT NULL
correlation_id TEXT
causation_id TEXT
```

#### artifacts

```text
id TEXT PRIMARY KEY
run_id TEXT NOT NULL
job_id TEXT
attempt_id TEXT
step_id TEXT
type TEXT NOT NULL
name TEXT NOT NULL
uri TEXT NOT NULL
media_type TEXT
size_bytes INTEGER
sha256 TEXT
status TEXT NOT NULL
producer_ref TEXT
provenance_json TEXT
retention_policy TEXT
redaction_level TEXT
created_at TEXT NOT NULL
finalized_at TEXT
metadata_json TEXT
```

#### approvals

```text
id TEXT PRIMARY KEY
run_id TEXT
job_id TEXT
attempt_id TEXT
step_id TEXT
type TEXT NOT NULL
status TEXT NOT NULL
requested_by TEXT
requested_at TEXT NOT NULL
resolved_by TEXT
resolved_at TEXT
expires_at TEXT
reason TEXT
payload_json TEXT
```

`attempt_id` is required for approvals requested during job execution. On retry, the controller checks for a reusable approval on the same `(job_id, action, resource)` tuple within the run.

#### policy_decisions

```text
id TEXT PRIMARY KEY
run_id TEXT
job_id TEXT
attempt_id TEXT
step_id TEXT
policy_ref TEXT
subject_type TEXT
subject_id TEXT
action TEXT NOT NULL
resource TEXT
outcome TEXT NOT NULL
reason TEXT
requires_approval INTEGER NOT NULL DEFAULT 0
approval_id TEXT
created_at TEXT NOT NULL
metadata_json TEXT
```

`attempt_id` is required for decisions made during job execution.

### Migration Approach

- SQL migrations live under `migrations/`.
- `oktopus db init` creates DB and applies all migrations.
- `schema_migrations` records applied versions.
- Migrations are append-only.
- Tests run migrations against temporary SQLite DB.

### Registry Loading

`oktopus registry validate` loads manifests from `registry/**/*.yaml`, `registry/**/*.yml`, and optionally `registry/**/*.toml` for compatibility with early seed manifests.

YAML is the default manifest format for MVP because capability packs, workflows, policies, and solution packs need nested structures and will later map cleanly to JSON Schema/OpenAPI-style validation.

Validation:

1. required envelope fields
2. known kind
3. unique `kind/name/version`
4. workflow references resolve where possible
5. policy fields are structurally valid

### Run Creation

`oktopus runs create <workflow>` does:

```text
load registry
resolve workflow capability
resolve project (see project auto-creation rules below)
snapshot resolved capabilities into resolved_capabilities_json
atomically increment project run counter and assign run.number
create run with status=created, emit run.created
expand workflow steps into jobs (with retry_policy from step or workflow defaults)
set jobs with dependencies to waiting_deps, emit job.waiting_deps
set root jobs to queued, emit job.queued
emit run.running
create run artifact directories
```

**Project auto-creation rules:**

1. If `--project <slug>` given: resolve by slug. Error if not found — do not auto-create.
2. If no `--project`: derive slug from `git remote get-url origin` normalized to `owner/repo`. If not a git repo, use current directory name.
3. If derived slug does not exist: create project with `auto_created: true` in metadata and insert a `project_counters` row. Emit `project.created`.
4. If derived slug exists but `root_uri` differs from current directory: error with "project slug conflict — use --project to specify".

### State Transitions

State changes must go through helper functions:

```text
CreateRun
QueueJob
LeaseJob
StartAttempt
CompleteAttempt      — includes evidence contract evaluation
FailAttempt
ExpireAttempt        — checks retry_policy.mode before requeueing
CompleteJob
FailJob
BlockJob             — for manual retry_policy on lease expiry
CancelJob
FinalizeArtifact
RequestApproval
ResolveApproval
CarryForwardApproval — reuse approval across attempts
```

Each helper updates current state and emits event in same DB transaction where possible.

### Local Worker MVP

`oktopus worker local --kind shell --once` flow:

```text
register worker
request eligible job by kind
accept lease transactionally
create attempt
heartbeat during execution
run command or stub action
write logs/artifacts
complete or fail attempt
release lease
advance dependent jobs
```

For non-shell job kinds, MVP can support stub execution:

```text
agent/review/synthesis → write placeholder report artifact + succeed when --stub allowed
```

This lets workflow DAG, leases, artifacts, events, and review gates be tested before real adapters.

### Initial Seed Capabilities

Keep registry seed small:

```text
skillpack:addyosmani-agent-skills
runtime:shell
runtime:pi
工具/tool:graphify
tool:openspec
verifier:artifact-exists
verifier:command-exit-zero
workflow:sdlc-default
workflow:guarded-build
```

### Initial Workflows

#### hello-local

Single shell job that writes a log and artifact.

#### guarded-build

```text
build-stub
  ├── correctness-review-stub
  ├── security-review-stub
  └── test-gap-review-stub
      ↓
review-synthesis-stub
      ↓
artifact-exists verifier
```

This proves adversarial review shape.

### Test Strategy

Unit tests:

- migration applies cleanly
- registry validates seed manifests
- run creation expands jobs
- dependency-ready query works
- lease conflict grants one worker only
- stale completion rejected
- artifact finalization computes digest
- state transitions emit events

Integration tests:

```text
oktopus db init
oktopus registry validate
oktopus runs create hello-local --title smoke
oktopus worker local --kind shell --once
oktopus runs show <id>
oktopus events list --run <id>
oktopus artifacts list --run <id>
```

### Acceptance Criteria

Local MVP is accepted when:

1. Fresh checkout can initialize DB.
2. Seed registry validates.
3. A workflow run creates jobs.
4. A local worker leases and completes a job.
5. Events are emitted for run/job/attempt/lease/artifact transitions.
6. At least one artifact is finalized with digest.
7. A verifier can pass/fail based on artifact existence or command exit.
8. Guarded workflow can model review jobs and synthesis job.

### Design Constraints

- Do not implement distributed service before local state model works.
- Do not special-case Pi in the core model.
- Do not bypass event emission for direct DB updates.
- Do not store artifact bytes in SQLite.
- Do not require external services for MVP.
- Keep CLI boring and scriptable.
