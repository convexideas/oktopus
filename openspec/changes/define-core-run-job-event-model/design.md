## Design

### Goal

Define the durable execution model for Oktopus.

The model must support:

- Local-first operation.
- Distributed controller/worker operation.
- Workflow DAGs.
- Fan-out/fan-in.
- Worker leases.
- Retries and repair attempts.
- Artifacts and evidence.
- Approval and policy gates.
- Adversarial review gates using isolated reviewer jobs/subagents.
- Event-based debugging and observability.

### Entity Overview

```text
Project
  └── Run
        ├── Job
        │     ├── Attempt
        │     │     ├── Step
        │     │     ├── Event
        │     │     └── Artifact
        │     └── Lease
        ├── Approval
        ├── PolicyDecision
        ├── Artifact
        └── Event

Worker
Environment
CapabilityRef
```

### Project

A project is the boundary for repository, product, tenant, or client context.

Minimum fields:

```text
id
name
slug
root_uri
metadata
created_at
updated_at
```

Examples:

- `convexideas/network-manager`
- `client-x/vlm-quality-harness`
- `internal/aiops-remediation`

### Run

A run is one invocation of a workflow or task.

Minimum fields:

```text
id
project_id
title
intent
workflow_ref
workflow_version
status
trigger_type
trigger_payload
resolved_capabilities
created_by
created_at
started_at
finished_at
parent_run_id
metadata
```

Run states:

```text
created
running
waiting_approval
canceling
canceled
failed
succeeded
archived
```

Run state transitions:

```text
created      → running          (first job queued)        → emit run.running
running      → waiting_approval (approval gate hit)       → emit run.waiting_approval
running      → canceling        (cancellation requested)  → emit run.canceling
running      → failed           (terminal job failure)    → emit run.failed
running      → succeeded        (all jobs succeeded)      → emit run.succeeded
waiting_approval → running      (approval resolved)       → emit run.running
waiting_approval → canceled     (approval rejected/expired) → emit run.canceled
canceling    → canceled         (all jobs wound down)     → emit run.canceled
succeeded    → archived         (manual or TTL)           → emit run.archived
failed       → archived         (manual or TTL)           → emit run.archived
```

`planning` and `queued` are job-level concepts and do not apply to runs. A run is `running` as soon as any job is queued. `blocked` is not a run state — runs are `waiting_approval` when gated; individual jobs may be `blocked` by unmet dependencies or policy denial.

### Job

A job is a schedulable node in the workflow DAG.

Minimum fields:

```text
id
run_id
key
name
kind
runtime_ref
capability_refs
depends_on
status
priority
max_attempts
timeout_seconds
input_artifacts
output_contract
output_kind
policy_ref
environment_ref
created_at
started_at
finished_at
metadata
```

`output_kind` declares what the job produces and is orthogonal to the run mode (one-shot, interactive, long-running, scheduled, event-triggered) defined in `define-enterprise-admin-marketplace-sessions`. Run mode is the engagement lifecycle; `output_kind` is the output modality. They compose freely (e.g. an interactive job may stream, a one-shot job may produce an artifact or an effect).

```text
artifact  -- a durable typed blob stored via the artifact store (default)
stream    -- a conversational/stdout message delivered to an output destination;
             may be logged as a transcript but is not required to be a stored artifact
effect    -- a state change in an external system (codebase mutation, commit, deploy,
             API call); verified by its consequence, not by a stored output blob
```

Acceptance branches on `output_kind` (see the evidence-contract enforcement in `define-artifacts-verifiers-policy`): artifact jobs require their declared artifacts; stream jobs require delivery (and optional transcript); effect jobs require their consequence verifier to pass. `output_kind` defaults to `artifact` when unset, preserving existing behavior.

Job states:

```text
created
waiting_deps
queued
leased
running
waiting_approval
blocked
canceling
canceled
failed
succeeded
skipped
expired
```

`kind` examples:

```text
agent
tool
verifier
approval
synthesis
inference
ci_remediation
telemetry_triage
docs_generation
adversarial_review
```

### Attempt

An attempt records one execution try for a job.

Minimum fields:

```text
id
job_id
attempt_number
worker_id
status
started_at
finished_at
exit_code
error_code
error_message
summary
metadata
```

Attempt states:

```text
created
running
succeeded
failed
expired
canceled
```

Attempts make retries inspectable instead of overwriting history.

### Step

A step is an observable unit inside an attempt.

Examples:

- model request
- tool call
- shell command
- verifier command
- approval request
- artifact write
- policy decision

Minimum fields:

```text
id
attempt_id
sequence
kind
name
status
started_at
finished_at
input_ref
output_ref
error_message
metadata
```

Step states:

```text
started
succeeded
failed
blocked
canceled
```

### Lease

A lease assigns a queued job to one worker for a bounded time.

Minimum fields:

```text
id
job_id
attempt_id
worker_id
status
leased_at
expires_at
heartbeat_at
released_at
release_reason
```

Lease states:

```text
active
released
expired
canceled
```

Lease rules:

1. Only one active lease may exist for a job.
2. Lease acquisition must be atomic.
3. Heartbeat extends lease only if current worker owns active lease.
4. Expired lease creates an expired attempt unless job policy says otherwise.
5. Requeued jobs increment attempt number.

### Worker

A worker is an executable service/process that polls for jobs.

Minimum fields:

```text
id
name
kind
version
status
labels
platform
last_seen_at
metadata
```

`current_job_id` is not stored on the worker record. Active job assignments are fully represented by the `leases` table (`WHERE worker_id = ? AND status = 'active'`). The `current_jobs` field in the DB schema is a denormalized count for scheduler queries only; the `leases` table is authoritative.

Worker examples:

```text
pi-worker
shell-worker
graphify-worker
openspec-worker
vlm-eval-worker
```

### Artifact

An artifact is durable evidence or output.

Minimum fields:

```text
id
run_id
job_id
attempt_id
step_id
type
name
uri
media_type
size_bytes
sha256
producer_ref
provenance
retention_policy
created_at
metadata
```

Artifact types:

```text
spec
plan
patch
diff
log
trace
graph
report
screenshot
eval_result
model_output
verifier_result
approval_record
review_report
review_finding
incident_report
runbook
adr
```

Artifact states are represented by events plus metadata; the record is immutable after digest finalization except retention/legal metadata.

### Event

Events are append-only records of state transitions and observations.

Minimum fields:

```text
id
run_id
job_id
attempt_id
step_id
type
severity
message
payload
actor_type
actor_id
timestamp
correlation_id
causation_id
```

Event rules:

1. Events are append-only.
2. State tables store current state for query speed.
3. Event log stores audit/debug truth.
4. Every state transition must emit an event.
5. Events must reference capability versions and worker identity when applicable.

Event severity values (required, enforced at ingestion):

```text
debug   — internal trace, not shown by default
info    — normal state transition
warn    — unexpected but recoverable (lease near expiry, retry triggered, deprecated capability in use)
error   — action required or job failed
```

Any worker-submitted event with an unknown severity is rejected at ingestion with a controller-emitted `event.rejected` event.

Minimum event types:

```text
run.created
run.running
run.waiting_approval
run.canceling
run.succeeded
run.failed
run.canceled
run.archived

job.created
job.queued
job.leased
job.started
job.waiting_deps
job.waiting_approval
job.succeeded
job.failed
job.expired
job.canceled
job.skipped

attempt.started
attempt.succeeded
attempt.failed
attempt.expired

step.started
step.succeeded
step.failed

artifact.created
artifact.finalized
artifact.retained
artifact.deleted

approval.requested
approval.approved
approval.rejected
approval.expired
approval.carried_forward

policy.evaluated
policy.denied
policy.allowed

worker.registered
worker.heartbeat
worker.offline

review.requested
review.started
review.finding_recorded
review.completed
review.accepted
review.rejected

event.rejected

capability.deprecated_in_use
```

### Adversarial Review

Adversarial review is a first-class workflow gate, not a casual prompt.

It is modeled as one or more `adversarial_review` jobs with isolated contexts and explicit review scopes. Review jobs may invoke agent runtimes as subagents/workers, but the workflow/controller owns fan-out and fan-in.

Typical review patterns:

```text
single-reviewer
  plan/build job → adversarial_review job → verifier/approval gate

parallel-review
  plan/build job
    ├── correctness-review
    ├── security-review
    ├── test-gap-review
    └── architecture-review
       ↓
    synthesis job → accept/reject gate

red-team-review
  high-risk job → adversarial_review job with hostile assumptions → repair job → re-review
```

Review job inputs:

```text
scope
claims_to_check
artifacts_to_review
allowed_context
review_persona_ref
review_skill_refs
risk_level
stop_rules
```

Review job outputs:

```text
review_report artifact
review_finding artifacts (must include severity field in provenance metadata)
accept/reject recommendation
required_repair_items
residual_risks
```

Review isolation is enforced at the controller level via a `context` block on the workflow step. The controller resolves the listed artifact refs from upstream jobs and creates a clean working directory containing only those files. The reviewer's filesystem scope is limited to that clean directory plus its own output directory.

Example workflow step with context block:

```yaml
- id: security-review
  kind: adversarial_review
  needs: [build]
  persona: persona:security-auditor
  context:
    artifacts:
      - from: build
        names: [patch.diff, build-report.md]
    read_registry: false
    read_run_artifacts: none   # none | explicit | all
```

`read_run_artifacts: all` must be explicitly set and triggers a `warn`-severity `policy.evaluated` event. Default is `none`.

Adversarial review rules:

1. Reviewer jobs run in isolated context. They receive only artifacts explicitly listed in the workflow `context` block.
2. Reviewer jobs cannot mutate source or activate capabilities unless granted by policy.
3. Review findings must cite artifacts, files, logs, tests, or source evidence. Each `review_finding` artifact must include `severity` in its provenance metadata (`debug|info|warn|error`).
4. High-risk workflows can require review acceptance before downstream jobs become eligible.
5. Controller performs fan-out/fan-in; personas do not recursively spawn other personas by default.

### Approval

An approval is a human or policy gate.

Minimum fields:

```text
id
run_id
job_id
attempt_id
step_id
type
status
reason
requested_by
requested_at
resolved_by
resolved_at
expires_at
payload
```

`attempt_id` is required. An approval is valid only for the attempt during which it was requested. On retry (new attempt), the controller checks if a non-expired, non-rejected approval exists for the same `(job_id, action, resource)` tuple within the current run. If found, it carries the approval forward without re-requesting and emits `approval.carried_forward`. Otherwise it requests a new approval.

Approval states:

```text
requested
approved
rejected
expired
canceled
```

### PolicyDecision

A policy decision records allow/deny/escalate outcomes for tools, secrets, network, filesystem, deployment, and capability activation.

Minimum fields:

```text
id
run_id
job_id
attempt_id
step_id
policy_ref
subject_type
subject_id
action
resource
outcome
reason
requires_approval
approval_id
created_at
payload
```

`attempt_id` is required for decisions made during job execution.

Outcomes:

```text
allowed
denied
requires_approval
```

### CapabilityRef

All runtime decisions must preserve capability provenance.

Minimum fields:

```text
kind
name
version
source
sha256_or_revision
```

Examples:

```text
tool:graphify@0.1.0
skillpack:addyosmani-agent-skills@0.1.0
workflow:sdlc-default@0.1.0
runtime:pi-worker@0.1.0
```

### Persistence

Local MVP:

```text
SQLite      → projects, runs, jobs, attempts, leases, workers, approvals, policy decisions, artifact metadata
filesystem  → artifact bytes, event jsonl mirror, logs
```

Distributed future:

```text
Postgres       → relational state
object store   → artifact bytes
queue/streams  → events and worker dispatch
OpenTelemetry  → traces/metrics/log export
```

### CLI Scope for First Implementation

Initial CLI should support:

```text
oktopus db init
oktopus registry validate
oktopus runs create <workflow> --title <title> --project <project>
oktopus runs list
oktopus runs show <run-id>
oktopus jobs list --run <run-id>
oktopus events tail --run <run-id>
```

Worker CLI comes in next change.

### Design Notes

- Use Go for implementation.
- Use UUID/ULID-style IDs for distributed-friendly identity.
- Store timestamps in UTC.
- Store current state in normalized tables; emit event for every transition.
- Keep artifact bytes out of DB.
- Keep model/tool payloads redaction-aware from the start.
- Prefer explicit state transitions over implicit status mutation.
