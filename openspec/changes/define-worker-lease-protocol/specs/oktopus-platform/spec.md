## ADDED Requirements

### Requirement: Worker Registration and Capability Advertisement

Oktopus SHALL allow workers to register identity, kind, version, labels, platform, capabilities, limits, and status before receiving jobs.

#### Scenario: Register agent worker

- **WHEN** a Pi, Claude Code, Codex CLI, Antigravity CLI, OpenCode, Kiro, Aider, Goose, or other agent worker starts
- **THEN** it registers as an `agent` worker with runtime capabilities, labels, platform, version, and concurrency limits
- **AND** Oktopus records the worker as available for matching jobs

#### Scenario: Register review worker

- **WHEN** an adversarial review worker starts
- **THEN** it registers as a `review` worker
- **AND** its default advertised policy denies source mutation, secret access, destructive tools, and capability activation unless explicitly overridden

### Requirement: Scheduler Matching

Oktopus SHALL match queued jobs to workers using job kind, runtime/tool references, labels, platform, resource limits, capability permissions, environment requirements, and policy constraints.

#### Scenario: Match job to compatible worker

- **WHEN** a worker requests work
- **THEN** Oktopus returns only jobs whose kind, runtime/tool requirements, labels, platform, and policy constraints are compatible with the worker

#### Scenario: Deny incompatible worker

- **WHEN** a worker requests a job requiring GPU, network access, trusted sandbox, secret access, or a specific runtime it does not advertise
- **THEN** Oktopus does not assign that job to the worker

### Requirement: Atomic Job Lease Acquisition

Oktopus SHALL assign queued jobs using an atomic lease acquisition operation.

#### Scenario: One worker wins lease

- **WHEN** multiple workers try to accept the same queued job
- **THEN** exactly one worker obtains the active lease
- **AND** all other workers receive lease denial and must request another job

#### Scenario: Record lease context

- **WHEN** a lease is granted
- **THEN** Oktopus records job ID, attempt ID, worker ID, lease ID, lease expiration, heartbeat timestamp, and attempt number
- **AND** emits `job.leased` and attempt lifecycle events

### Requirement: Lease Heartbeats and Expiration

Oktopus SHALL require active workers to heartbeat leases before expiration and SHALL recover expired leases.

#### Scenario: Extend active lease

- **WHEN** a worker heartbeats an active lease it owns before expiration
- **THEN** Oktopus updates the heartbeat timestamp and extends the lease expiration according to policy

#### Scenario: Expire abandoned job

- **WHEN** a worker fails to heartbeat before the lease expiration and grace period
- **THEN** Oktopus marks the lease expired
- **AND** marks the active attempt expired
- **AND** requeues or fails the job according to retry policy

### Requirement: At-Least-Once Execution Semantics

Oktopus SHALL assume at-least-once job execution in distributed mode and SHALL reject stale completions.

#### Scenario: Reject stale completion

- **WHEN** a worker reports completion for a lease that is expired, canceled, released, or no longer owned by that worker
- **THEN** Oktopus rejects the completion as stale
- **AND** records an event for audit

#### Scenario: Preserve idempotent artifacts

- **WHEN** a retried job writes artifacts
- **THEN** artifacts are namespaced by run, job, and attempt
- **AND** artifact finalization records content digest and producing attempt

### Requirement: Worker Cancellation and Draining

Oktopus SHALL support cooperative cancellation and worker draining.

#### Scenario: Cancel active job

- **WHEN** a run or job is canceled while leased
- **THEN** Oktopus marks the job or run canceling
- **AND** notifies the owning worker through heartbeat response, cancel stream, or polling
- **AND** records partial logs and artifacts before final cancellation when possible

#### Scenario: Drain worker

- **WHEN** a worker enters draining state
- **THEN** Oktopus stops assigning it new jobs
- **AND** allows existing leased jobs to complete, cancel, or expire

### Requirement: Worker Event, Log, and Artifact Reporting

Oktopus SHALL allow workers to report observations while the controller remains authoritative for state transitions.

#### Scenario: Worker emits execution observations

- **WHEN** a worker starts steps, receives runtime output, records review findings, runs verifiers, or creates artifacts
- **THEN** it reports events, log lines, and artifact metadata linked to the active run, job, attempt, and lease

#### Scenario: Controller records authoritative transition

- **WHEN** a worker reports success, failure, cancellation, or expiration-relevant data
- **THEN** Oktopus validates the active lease and records the authoritative state transition event

### Requirement: Local Worker Uses Same Protocol

Oktopus SHALL use the same worker lease protocol for local MVP workers and remote distributed workers.

#### Scenario: Run local shell worker

- **WHEN** Oktopus runs a local shell worker in MVP mode
- **THEN** the worker still registers, requests jobs, accepts leases, heartbeats, reports events, uploads artifacts, and completes jobs through the same state model used by remote workers

### Requirement: Review Workers Use Isolated Context by Default

Oktopus SHALL run adversarial review workers with isolated context and non-mutating defaults unless workflow policy explicitly grants additional permissions.

#### Scenario: Execute adversarial review job

- **WHEN** an adversarial review job is leased by a review worker
- **THEN** the worker receives explicit review scope, artifact references, claims to check, and allowed context
- **AND** it cannot mutate source, activate capabilities, access secrets, or run destructive tools unless policy allows it
