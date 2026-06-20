## ADDED Requirements

### Requirement: Local-First MVP CLI

Oktopus SHALL provide a local-first CLI that initializes state, validates registry manifests, creates workflow runs, inspects runs/jobs/events/artifacts, and runs local workers.

#### Scenario: Initialize local database

- **WHEN** a developer runs `oktopus db init`
- **THEN** Oktopus creates the configured SQLite database if missing
- **AND** applies all pending migrations exactly once

#### Scenario: Validate local registry

- **WHEN** a developer runs `oktopus registry validate`
- **THEN** Oktopus loads local capability manifests
- **AND** reports schema, uniqueness, and reference validation errors with file and field context

#### Scenario: Create local workflow run

- **WHEN** a developer runs `oktopus runs create <workflow> --title <title>`
- **THEN** Oktopus resolves the workflow capability
- **AND** creates a durable run, expands workflow steps into jobs, emits events, and creates run artifact directories

### Requirement: Go Core Implementation

Oktopus SHALL implement the local MVP core in Go.

#### Scenario: Build single local binary

- **WHEN** Oktopus is built for local MVP
- **THEN** it produces a single `oktopus` CLI binary containing registry, DB, run, scheduler, worker, artifact, event, verifier, and policy MVP logic

### Requirement: SQLite Local State Store

Oktopus SHALL use SQLite as the local MVP relational state store.

#### Scenario: Store core execution state

- **WHEN** Oktopus creates or updates projects, capabilities, runs, jobs, attempts, leases, workers, steps, artifacts, events, verifier results, policy decisions, or approvals
- **THEN** it persists the current state in SQLite

#### Scenario: Apply migrations safely

- **WHEN** migrations are run multiple times
- **THEN** Oktopus applies each migration once
- **AND** records applied migration versions in `schema_migrations`

### Requirement: Filesystem Artifact Store

Oktopus SHALL store local MVP artifact bytes on the filesystem and artifact metadata in SQLite.

#### Scenario: Finalize local artifact

- **WHEN** a worker or verifier writes an artifact
- **THEN** Oktopus stores the bytes under a run/job/attempt-scoped path
- **AND** records URI, size, digest, type, producer, provenance, and finalization timestamp in SQLite

### Requirement: Workflow Expansion into Jobs

Oktopus SHALL expand workflow manifests into durable job records during run creation.

#### Scenario: Expand workflow steps

- **WHEN** a workflow contains ordered or dependency-linked steps
- **THEN** Oktopus creates one job per executable step
- **AND** records job key, name, kind, runtime or capability references, dependency list, output contract, policy, environment, and initial status

#### Scenario: Queue dependency-free jobs

- **WHEN** jobs have no unsatisfied dependencies
- **THEN** Oktopus marks them `queued`
- **AND** leaves dependent jobs in `waiting_deps` until upstream dependency policy is satisfied

### Requirement: Transactional State Transitions with Events

Oktopus SHALL update current state and emit corresponding events through explicit transition helpers.

#### Scenario: Transition job state

- **WHEN** Oktopus queues, leases, starts, completes, fails, expires, cancels, or skips a job
- **THEN** the state update and event emission happen together in the same logical operation

#### Scenario: Prevent silent mutation

- **WHEN** core state changes outside a transition helper
- **THEN** tests or validation should detect missing event emission for the changed state path

### Requirement: Local Worker Uses Lease Protocol

Oktopus SHALL provide a local worker command that uses the same registration, request, accept, lease, heartbeat, artifact, event, and completion model intended for distributed workers.

#### Scenario: Run local worker once

- **WHEN** a developer runs `oktopus worker local --kind shell --once`
- **THEN** the worker registers, requests an eligible job, accepts a lease, creates an attempt, heartbeats during execution, writes logs/artifacts, completes or fails the attempt, releases the lease, and advances downstream jobs where applicable

#### Scenario: Stub unsupported local job kind

- **WHEN** the local MVP encounters an agent, review, or synthesis job without a real adapter and stub mode is enabled
- **THEN** it writes a placeholder report artifact and completes the job using normal lease, artifact, and event flow

### Requirement: Seed Capabilities and Starter Workflows

Oktopus SHALL ship a minimal local registry sufficient to validate the execution spine.

#### Scenario: Validate seed capabilities

- **WHEN** a developer validates the initial registry
- **THEN** Oktopus recognizes seed capabilities for shell runtime, Pi runtime stub, Graphify tool stub, OpenSpec tool stub, artifact-exists verifier, command-exit-zero verifier, and starter workflows

#### Scenario: Run hello workflow

- **WHEN** a developer creates and executes the `hello-local` workflow
- **THEN** Oktopus completes at least one local shell job, emits events, writes logs, and finalizes at least one artifact with digest

#### Scenario: Model guarded workflow

- **WHEN** a developer creates the `guarded-build` workflow
- **THEN** Oktopus creates build, adversarial review, synthesis, and verifier jobs with correct dependency shape

### Requirement: Local MVP Acceptance Tests

Oktopus SHALL include smoke tests that prove the local execution spine works without external services.

#### Scenario: Run local smoke test

- **WHEN** the smoke test runs from a clean temporary workspace
- **THEN** it initializes SQLite, validates registry, creates a workflow run, leases and completes a local worker job, emits events, finalizes an artifact, and lists run/job/event/artifact state successfully
