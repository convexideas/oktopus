## ADDED Requirements

### Requirement: Durable Artifact Evidence

Oktopus SHALL store typed artifact metadata durably and artifact bytes outside the relational database.

#### Scenario: Finalize artifact

- **WHEN** a worker, controller, human, or external system produces a report, diff, log, graph, screenshot, verifier result, review finding, incident report, runbook, ADR, eval result, or dataset manifest
- **THEN** Oktopus records artifact type, name, URI, media type, size, digest, producer, provenance, retention policy, redaction level, and timestamps
- **AND** stores artifact bytes in the configured artifact backend

#### Scenario: Preserve immutable evidence

- **WHEN** an artifact is finalized
- **THEN** its bytes are immutable
- **AND** any replacement creates a new artifact record with its own digest and provenance

### Requirement: Artifact Provenance and Redaction

Oktopus SHALL record provenance and sensitivity metadata for artifacts.

#### Scenario: Track artifact origin

- **WHEN** an artifact is finalized
- **THEN** Oktopus records the producing worker, capability reference, command/tool/runtime reference, input artifact references, source URI or revision when available, and redaction policy

#### Scenario: Restrict sensitive artifact

- **WHEN** an artifact contains secrets, PII, client data, proprietary data, or incident-sensitive information
- **THEN** Oktopus marks it with restricted or redacted access metadata
- **AND** prevents unauthorized access or external sharing according to policy

### Requirement: Verifier Capabilities and Results

Oktopus SHALL model verifiers as capabilities that produce durable verifier result records and artifacts.

#### Scenario: Run verifier

- **WHEN** a workflow requires a verifier such as command exit, artifact existence, OpenSpec validation, tests, lint, security scan, review finding check, accuracy threshold, latency budget, or telemetry recovery check
- **THEN** Oktopus schedules or executes the verifier
- **AND** records verifier status, summary, inputs, output artifact, exit code, errors, and timestamps

#### Scenario: Block on failed verifier

- **WHEN** a required verifier fails or errors
- **THEN** Oktopus blocks dependent jobs unless workflow policy explicitly allows failure or an approved waiver exists

### Requirement: Evidence Contracts

Oktopus SHALL support evidence contracts that define required artifacts, verifier results, approvals, and review outcomes before a job or run is accepted. Contracts SHALL be interpreted according to the job's output kind (`artifact`, `stream`, or `effect`), because not every job produces a stored artifact.

#### Scenario: Accept artifact-producing job

- **WHEN** a job with output kind `artifact` declares required output artifacts and verifiers
- **THEN** Oktopus marks the job accepted only after required artifacts are finalized and required verifiers pass or are waived

#### Scenario: Accept conversational (stream) job without an artifact

- **WHEN** a job with output kind `stream` delivers its message to an approved output destination
- **THEN** Oktopus accepts the job on successful delivery
- **AND** does not require a stored artifact, though a transcript may be captured as optional evidence

#### Scenario: Accept codebase (effect) job by consequence

- **WHEN** a job with output kind `effect` acts on an external system such as a codebase or deployment
- **THEN** Oktopus accepts the job only when its consequence verifiers pass for the attempt
- **AND** the acceptance criterion is the effect's verified consequence, not a stored output blob

#### Scenario: Produce final receipts

- **WHEN** a run completes
- **THEN** Oktopus records a final evidence summary that lists required artifacts, verifier results, delivered stream destinations, effect consequence checks, approvals, review findings, and waivers

### Requirement: Waiver Records

Oktopus SHALL treat waivers as explicit auditable exceptions.

#### Scenario: Waive failed verifier

- **WHEN** a human or policy grants a waiver for a failed verifier, missing artifact, unresolved review finding, or policy exception
- **THEN** Oktopus records actor, reason, scope, expiration, linked evidence, and approval reference
- **AND** emits a waiver event

#### Scenario: Prevent silent bypass

- **WHEN** evidence is missing or failed
- **THEN** Oktopus SHALL NOT silently mark the job or run accepted without an approved waiver

### Requirement: Policy Decision Enforcement

Oktopus SHALL record and enforce policy decisions for tools, secrets, network, filesystem, source mutation, destructive commands, capability activation, production deployment, external API calls, budgets, and human/client data access.

#### Scenario: Evaluate privileged action

- **WHEN** a job requests a privileged action
- **THEN** Oktopus evaluates applicable policy
- **AND** records whether the action is allowed, denied, or requires approval

#### Scenario: Deny action outside policy

- **WHEN** a requested action violates policy and no approval path exists
- **THEN** Oktopus denies the action
- **AND** records a policy denial event linked to the run, job, attempt, and step

### Requirement: Approval Gates

Oktopus SHALL support scoped approval gates for actions that require human or higher-trust authorization.

#### Scenario: Request approval

- **WHEN** policy requires approval for destructive command, secret access, network access, source mutation, production deploy, capability activation, waiver, client data access, or remediation execution
- **THEN** Oktopus creates an approval record with action, resource, risk, proposed command/tool details, requester, expiration, and scope

#### Scenario: Scope approval grant

- **WHEN** an approval is granted
- **THEN** the grant applies only to the recorded run, job, action, resource, and expiration unless explicitly modeled as a policy or capability update

### Requirement: Adversarial Review Evidence

Oktopus SHALL capture adversarial review findings as durable evidence that can gate downstream workflow execution.

#### Scenario: Record review finding

- **WHEN** a review worker finds a correctness, security, test-gap, architecture, performance, policy, or documentation issue
- **THEN** Oktopus records the finding as a structured event or artifact
- **AND** requires citations to reviewed artifacts, files, logs, traces, tests, specs, or source evidence

#### Scenario: Block on critical finding

- **WHEN** a review gate produces critical unresolved findings
- **THEN** Oktopus blocks downstream acceptance unless a repair job resolves the finding or an approved waiver is recorded

### Requirement: CI/CD Evidence Capture

Oktopus SHALL represent CI/CD inputs and outcomes as artifacts and verifier results.

#### Scenario: Ingest failed CI

- **WHEN** a CI/CD integration starts a remediation run
- **THEN** Oktopus stores build logs, test reports, changed files, pipeline metadata, and failure summaries as artifacts
- **AND** uses verifier results to prove whether remediation succeeded

### Requirement: Telemetry and AIOps Evidence Capture

Oktopus SHALL represent telemetry inputs, incident analysis, remediation actions, and recovery checks as artifacts, policy decisions, approvals, and verifier results.

#### Scenario: Verify incident recovery

- **WHEN** an AIOps remediation workflow executes or recommends a remediation
- **THEN** Oktopus records RCA artifacts, runbook artifacts, policy decisions, approval records where required, and recovery verifier results based on telemetry signals

### Requirement: Client Inference Evidence Capture

Oktopus SHALL represent inference product evaluation and deployment gates as artifacts and verifiers.

#### Scenario: Gate VLM deployment

- **WHEN** a VLM solution pack evaluates a model or prompt for deployment
- **THEN** Oktopus records dataset manifests, golden sets, eval results, confusion matrices, human review reports, latency reports, and policy decisions as artifacts
- **AND** blocks deployment until required quality, latency, safety, and data-boundary verifiers pass or are waived

### Requirement: Local-First Artifact Storage with Distributed Upgrade Path

Oktopus SHALL support filesystem-backed artifact storage in local mode and object-store-backed artifact storage in distributed mode.

#### Scenario: Store local artifact

- **WHEN** Oktopus runs on a single machine
- **THEN** it stores artifact bytes under run/job/attempt-scoped filesystem paths and records metadata in SQLite

#### Scenario: Store distributed artifact

- **WHEN** Oktopus runs for a team or organization
- **THEN** it can store artifact bytes in object storage while preserving the same artifact metadata, digest, provenance, and event model
