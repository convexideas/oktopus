## ADDED Requirements

### Requirement: Durable Core Execution Model

Oktopus SHALL define durable records for projects, runs, jobs, attempts, steps, leases, workers, artifacts, events, approvals, policy decisions, and capability references.

#### Scenario: Create a run from workflow intent

- **WHEN** a user, API, CI/CD system, telemetry alert, or scheduled trigger creates a run
- **THEN** Oktopus records the project, run intent, trigger payload, workflow reference, resolved capability references, and initial run status
- **AND** this state is stored outside any agent/model context window

#### Scenario: Preserve retry history

- **WHEN** a job fails and is retried
- **THEN** Oktopus creates a new attempt record
- **AND** preserves previous attempt status, errors, logs, artifacts, and events for audit and debugging

### Requirement: Workflow Jobs as Schedulable DAG Nodes

Oktopus SHALL represent workflow DAG nodes as jobs with dependency, runtime, policy, environment, input artifact, output contract, timeout, retry, and priority metadata.

#### Scenario: Queue only dependency-ready jobs

- **WHEN** a workflow run contains jobs with dependencies
- **THEN** Oktopus marks jobs with unsatisfied dependencies as `waiting_deps`
- **AND** only marks a job `queued` after its dependency policy is satisfied

#### Scenario: Synthesize fan-out outputs

- **WHEN** several parallel jobs feed a downstream synthesis job
- **THEN** Oktopus provides the synthesis job with artifact references and status summaries from the completed upstream jobs

### Requirement: Atomic Worker Lease Protocol

Oktopus SHALL assign queued jobs to workers through an atomic lease protocol.

#### Scenario: Prevent duplicate execution

- **WHEN** multiple workers request the same eligible job
- **THEN** only one worker can acquire the active lease
- **AND** Oktopus records the worker identity, lease expiration, heartbeat timestamp, and attempt number

#### Scenario: Recover expired work

- **WHEN** a worker stops heartbeating before completing a leased job
- **THEN** Oktopus expires the lease
- **AND** marks the active attempt expired
- **AND** requeues or fails the job according to retry policy

### Requirement: Append-Only Event Log

Oktopus SHALL emit append-only events for every important state transition and observation.

#### Scenario: Trace a failed run

- **WHEN** an operator inspects a failed run
- **THEN** Oktopus exposes an ordered event timeline that includes run, job, attempt, step, artifact, approval, policy, worker, and review events
- **AND** each event includes timestamp, type, actor, correlation, causation, and payload metadata where applicable

#### Scenario: Query current state efficiently

- **WHEN** a client lists runs or jobs
- **THEN** Oktopus can read current state from normalized state tables
- **AND** the append-only event log remains the audit/debug source of truth

### Requirement: Artifact Evidence Model

Oktopus SHALL store artifact metadata in durable state and artifact bytes outside the relational database.

#### Scenario: Finalize artifact evidence

- **WHEN** a worker writes a report, diff, log, graph, eval result, verifier result, review finding, ADR, incident report, or runbook
- **THEN** Oktopus records artifact type, URI, media type, size, digest, producer, provenance, retention policy, and creation time

### Requirement: Adversarial Review Gate

Oktopus SHALL support adversarial review as a first-class workflow gate using isolated review jobs and optional reviewer subagents/workers.

#### Scenario: Review before accepting high-risk work

- **WHEN** a workflow marks a job or transition as requiring adversarial review
- **THEN** Oktopus schedules one or more `adversarial_review` jobs before downstream acceptance jobs become eligible
- **AND** downstream jobs remain blocked until review policy is satisfied

#### Scenario: Parallel specialist review

- **WHEN** a workflow requests parallel correctness, security, test-gap, or architecture review
- **THEN** Oktopus schedules independent review jobs with isolated context and explicit artifact inputs
- **AND** a synthesis job can merge review reports into an accept, reject, or repair recommendation

#### Scenario: Keep review non-mutating by default

- **WHEN** an adversarial review job runs
- **THEN** it can read assigned artifacts and context
- **AND** it cannot mutate source, execute destructive tools, activate capabilities, or access secrets unless policy explicitly allows it

#### Scenario: Require evidence-backed findings

- **WHEN** a review job records a finding
- **THEN** the finding is captured as an artifact or structured event
- **AND** it cites reviewed artifacts, files, logs, tests, traces, specs, or source evidence

### Requirement: Approval and Policy Decision Records

Oktopus SHALL record human approvals and policy decisions as durable, auditable records.

#### Scenario: Gate risky action

- **WHEN** a job requests a destructive command, production deployment, secret, network route, filesystem path, or capability activation outside automatic policy
- **THEN** Oktopus records a policy decision
- **AND** either denies the action or creates an approval request before execution proceeds

### Requirement: Storage Adapter Boundary

Oktopus SHALL keep core execution logic behind storage adapter or repository boundaries so local and distributed deployments can share the same run/job/event semantics.

#### Scenario: Avoid SQLite coupling in core logic

- **WHEN** Oktopus implements run creation, scheduling, worker leasing, state transitions, artifact metadata, approvals, or policy decisions
- **THEN** that logic uses store/repository interfaces instead of directly constructing SQLite connections
- **AND** SQLite-specific locking, pragmas, and SQL dialect choices remain isolated to the local storage adapter or migration layer

### Requirement: Local-First Persistence with Distributed Upgrade Path

Oktopus SHALL support local-first persistence while preserving a path to distributed storage adapters.

#### Scenario: Run local MVP

- **WHEN** Oktopus runs on a single machine
- **THEN** it can store relational state in SQLite and artifact bytes on the local filesystem

#### Scenario: Upgrade to team deployment

- **WHEN** Oktopus runs for a team or organization
- **THEN** the same execution model can use Postgres for state, object storage for artifacts, and stream/queue backends for events and worker dispatch
