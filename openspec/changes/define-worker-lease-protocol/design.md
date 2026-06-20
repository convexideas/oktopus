## Design

### Goal

Define the worker lease protocol for Oktopus.

The protocol must support:

- Local worker subprocess in MVP.
- Remote workers later.
- Atomic job assignment.
- Crash recovery.
- Cancellation.
- Heartbeats.
- Capability and policy matching.
- Sandboxed execution.
- Event/log/artifact streaming.
- Adversarial review workers/subagents.

### Worker Model

A worker is a process or service that executes one or more job kinds.

Worker kinds:

```text
shell       = run commands/scripts in environment
agent       = invoke agent runtime such as Pi, Claude Code, Codex CLI, Antigravity CLI, OpenCode, Kiro, Aider, Goose
tool        = run a tool adapter such as Graphify, OpenSpec, Beads, Archon
verifier    = run evidence checks
review      = run adversarial review persona/subagent
inference   = run model eval/inference tasks such as VLM/RAG workflows
synthesis   = merge upstream artifacts into decisions/reports
```

Worker record:

```text
id
name
kind
version
status
labels
platform
capabilities
max_concurrency
current_jobs
last_seen_at
started_at
metadata
```

`current_job_id` (singular) is not on the worker record. Active job assignments are represented by the `leases` table (`WHERE worker_id = ? AND status = 'active'`). `current_jobs` is a denormalized count for scheduler eligibility queries only; the `leases` table is authoritative.

Worker statuses:

```text
starting
idle
busy
draining
offline
unhealthy
retired
```

### Capability Advertisement

Workers advertise capabilities and constraints.

```yaml
worker:
  kind: agent
  runtime: pi
  version: 0.1.0
  labels:
    os: darwin
    arch: arm64
    sandbox: local
    network: limited
    gpu: false
  capabilities:
    runtimes:
      - pi
    tools:
      - read
      - bash
      - edit
    skills:
      - graphify
  limits:
    max_concurrency: 1
    max_job_seconds: 3600
```

Scheduler matches jobs by:

```text
kind
runtime_ref
tool_ref
required_labels
forbidden_labels
platform
policy constraints
resource limits
secret/network/filesystem requirements
```

### Protocol Overview

```text
Register/Heartbeat
  worker → controller: identity, version, capabilities, labels, limits

Request
  worker → controller: kind, labels, capacity, accepted policies
  controller → worker: candidate job or no work

Accept
  worker → controller: job_id, worker_id, lease_ttl, idempotency_key
  controller → worker: lease granted or denied

Execute
  worker → controller: events, logs, artifact uploads, heartbeats

Complete/Fail
  worker → controller: final status, exit code, summary, artifact refs

Cancel/Drain
  controller → worker: cancellation or drain signal
```

### Atomic Lease Acquisition

Lease acquisition must be compare-and-swap on job state.

Preconditions:

```text
job.status = queued
no active lease exists for job
job dependencies satisfied
job policy allows worker
job attempt_count < max_attempts
```

On success:

```text
create attempt N
create active lease
set job.status = leased
set worker.current_jobs += job
emit job.leased
emit attempt.started when worker starts execution
```

On conflict:

```text
return lease_denied
worker requests another job
```

### Lease TTL and Heartbeats

Each active lease has `expires_at` and `heartbeat_at`.

Rules:

1. Worker heartbeats before `expires_at`.
2. Controller extends lease only if `worker_id` owns active lease.
3. Missed heartbeat does not instantly kill job; controller marks lease expired after grace period.
4. Expired lease marks attempt `expired` and job `queued` or `failed` based on retry policy.
5. Worker completion after lease expiration is rejected unless controller can safely reconcile it.

Suggested MVP defaults:

```text
lease_ttl_seconds = 60
heartbeat_interval_seconds = 20
lease_grace_seconds = 30
max_attempts = 2
```

### Execution Semantics

Oktopus assumes **at-least-once job execution** in distributed mode. Workers must declare their retry policy so the controller can enforce the correct behavior on lease expiry.

Every job carries a `retry_policy` with three modes:

```text
idempotent   — default; on lease expiry, requeue job and increment attempt number
at-most-once — on lease expiry, mark job failed immediately; do not requeue
manual       — on lease expiry, mark job blocked; require explicit `oktopus jobs retry <id>`
```

Jobs default to `idempotent`. Workflow authors must explicitly declare `at-most-once` or `manual` for jobs that:

- have `permissions.network != false`
- have `source_mutation: true`
- reference a policy that includes `production_deploy` or `destructive_command`

The controller enforces `retry_policy.mode` at the requeue decision point — not at lease grant. `max_attempts` is respected for `idempotent` jobs; `at-most-once` jobs fail on first lease expiry regardless of `max_attempts`.

Artifact writes use content digests and attempt IDs. Completion calls include lease ID and attempt ID. Controller rejects completion from stale leases.

Oktopus does not promise global exactly-once execution. It promises at-most-one active lease at a time.

### Cancellation

Cancellation can target run, job, attempt, or worker.

Flow:

```text
controller marks run/job canceling
controller emits cancellation event
worker observes cancellation through heartbeat response or cancel stream
worker sends best-effort termination to runtime/sandbox
worker uploads partial logs/artifacts
worker completes attempt as canceled
controller marks job/run canceled or failed according to policy
```

Cancellation is cooperative first. Sandboxes may enforce hard kill after timeout.

### Draining

Workers can enter `draining` state.

Rules:

1. Draining worker receives no new jobs.
2. Existing jobs continue until completion, cancellation, or lease expiration.
3. Once no active jobs remain, worker becomes `retired` or `offline`.

### Event Responsibilities

Controller emits authoritative state transition events.

Workers emit observation events.

Worker-emitted event examples:

```text
worker.heartbeat
step.started
step.succeeded
step.failed
artifact.created
artifact.finalized
log.line
runtime.output
review.finding_recorded
verifier.result
```

Controller validates and records them with run/job/attempt/lease context.

### Logs and Artifacts

Workers stream logs and upload artifacts under attempt-scoped paths.

Local MVP layout:

```text
runs/<run-id>/jobs/<job-key>/attempts/<n>/logs/stdout.log
runs/<run-id>/jobs/<job-key>/attempts/<n>/artifacts/<artifact-name>
```

Distributed future:

```text
object://artifacts/<run-id>/<job-id>/<attempt-id>/...
```

Artifact metadata is recorded in DB and linked to producing step/attempt/job/run.

### Policy Integration

Before lease grant, controller checks:

```text
worker trust level
worker labels/platform
job policy
capability permissions
required secrets
network policy
filesystem scope
sandbox requirement
approval state
budget limits
```

During execution, worker must request privileged operations through controller or policy-aware adapter when possible.

### Capability Pinning at Execution Time

Workers MUST use the capability refs from `run.resolved_capabilities_json`, not the live registry.

The controller includes the resolved capability manifest in the lease grant response. Workers are forbidden from re-resolving capability names at execution time.

At run creation, `resolved_capabilities_json` is written as a snapshot of all capabilities the workflow references, including their full manifests — not just refs. The controller validates this snapshot is still resolvable at lease grant (capability not deleted or quarantined). If a capability was deprecated between run creation and lease grant, the controller allows it (the run was already resolved) but emits a `capability.deprecated_in_use` warn-severity event.

### Adversarial Review Workers

Review workers are normal workers with stricter defaults.

Defaults:

```text
can_mutate_source: false
can_activate_capabilities: false
can_access_secrets: false
can_run_destructive_tools: false
context_mode: isolated
```

Review context isolation is enforced by the controller at lease grant. The controller resolves the artifact refs listed in the workflow step's `context` block, creates a clean working directory containing only those files, and sets that as the worker's filesystem root. The worker's `scoped_workspace` is the clean directory plus its own output path. It cannot traverse to the broader run directory unless `read_run_artifacts: all` is explicitly set in the workflow step.

See the `context` block specification in the run/job model design.

### Local-First Implementation

MVP can implement protocol in-process:

```text
oktopus worker local --kind shell
oktopus worker local --kind verifier
```

The same state transitions and lease tables are used even when worker is local. This prevents redesign when moving to remote workers.

### Future API Shape

Possible API endpoints:

```text
POST /workers/register
POST /workers/{id}/heartbeat
POST /jobs/request
POST /jobs/{id}/accept
POST /leases/{id}/heartbeat
POST /leases/{id}/complete
POST /leases/{id}/fail
POST /leases/{id}/cancelled
POST /artifacts
POST /events
```

Wire protocol can be HTTP first, gRPC later if needed.

### Failure Modes

| Failure | Required behavior |
|---|---|
| worker crashes | lease expires, attempt expired, job retried or failed per retry_policy.mode |
| controller restarts | leases recovered from DB, expired leases reconciled |
| duplicate accept | one accept wins; others denied |
| stale completion | rejected if lease/attempt no longer active |
| artifact upload partial | artifact not finalized; event records failure |
| cancellation ignored | hard kill via sandbox after timeout where supported |
| policy changes mid-job | controller can cancel, block next privileged operation, or let current lease finish based on policy |

### Design Constraints

- Controller remains source of truth.
- Workers are disposable.
- Leases are bounded.
- State transitions are explicit and evented.
- Execution is at-least-once, not globally exactly-once.
- Sandboxing is an environment capability, not an afterthought.
