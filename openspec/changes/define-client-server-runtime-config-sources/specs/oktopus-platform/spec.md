## ADDED Requirements

### Requirement: Client/Server Control Plane

Oktopus SHALL operate as a client/server control plane where the server owns configuration, registry, routing, scheduling, policy, state, events, artifacts, and worker coordination.

#### Scenario: CLI uses server-compatible model

- **WHEN** Oktopus runs in local CLI mode
- **THEN** it uses the same run, job, worker, registry, policy, artifact, and event concepts intended for server mode

#### Scenario: Background agents execute bounded jobs

- **WHEN** Pi, Claude Code, Codex CLI, Antigravity CLI, OpenCode, Kiro, Aider, Goose, shell, tool, verifier, review, inference, or knowledge workers execute work
- **THEN** they register as workers and lease bounded jobs from Oktopus rather than acting as autonomous control planes

### Requirement: Message and Command Input Surfaces

Oktopus SHALL support both natural-language message inputs and imperative command inputs.

#### Scenario: Imperative command creates run

- **WHEN** a user or API submits an explicit command naming a workflow and inputs
- **THEN** Oktopus validates the workflow, inputs, policy, and permissions before creating a run

#### Scenario: Message input routes through policy

- **WHEN** a user sends a natural-language message requesting work
- **THEN** Oktopus may classify intent and recommend a workflow
- **AND** it asks for clarification or approval when confidence is low or requested action is risky
- **AND** it creates a run only when user selection, trigger rules, or policy permit it

### Requirement: Configurable Capability Sources

Oktopus SHALL support configurable sources for skills, tools, personas, workflows, runtimes, adapters, verifiers, policies, and solution packs.

#### Scenario: Load skill source

- **WHEN** configuration declares a skill source such as filesystem, git, npm, or built-in package
- **THEN** Oktopus discovers declared skills, personas, command recipes, references, and provenance according to that source adapter

#### Scenario: Load tool source

- **WHEN** configuration declares a tool source such as local manifests, MCP servers, CLI allowlists, APIs, or built-ins
- **THEN** Oktopus makes those tools available as governed capabilities subject to policy and workflow scope

#### Scenario: Preserve source provenance

- **WHEN** Oktopus loads any capability from a configured source
- **THEN** it records source type, URI/path, version/ref, trust level, and manifest path or digest where available

### Requirement: Configurable Knowledge Sources

Oktopus SHALL support configurable knowledge sources such as vaults, documentation repositories, OpenSpec specs, service catalogs, vector databases, runbooks, tickets, incidents, telemetry stores, and prior run artifacts.

#### Scenario: Query vault for context

- **WHEN** a job is allowed to use a configured vault or knowledge base
- **THEN** Oktopus queries or retrieves relevant context through the source adapter
- **AND** includes only policy-allowed excerpts, summaries, or artifact references in the worker context bundle

#### Scenario: Protect sensitive knowledge

- **WHEN** a knowledge source contains secrets, PII, client data, incident-sensitive data, or restricted organizational knowledge
- **THEN** Oktopus applies access policy, redaction, and audit events before exposing content to a worker

### Requirement: Configuration Hierarchy

Oktopus SHALL resolve configuration using the canonical precedence ladder `builtin < upstream_pack < org < client < project < workflow < run_override`, the same ladder used for capability resolution.

#### Scenario: Resolve scoped configuration

- **WHEN** Oktopus prepares a run or job
- **THEN** it resolves configuration in deterministic order using the canonical ladder from `builtin` to `run_override`
- **AND** records the effective configuration references that influenced workflow selection and worker handoff

#### Scenario: Gate permission broadening

- **WHEN** a higher-priority configuration broadens permissions for tools, secrets, network, filesystem, data access, workflow activation, or deployment
- **THEN** Oktopus requires policy evaluation and approval where configured

### Requirement: Workflow Selection Authority

Oktopus SHALL keep workflow selection under server, user, trigger, and policy authority rather than worker authority.

#### Scenario: Select workflow from command

- **WHEN** a user explicitly requests an active workflow
- **THEN** Oktopus may create a run after validation and policy checks

#### Scenario: Record worker recommendation

- **WHEN** an agent, Archon, or background worker recommends a workflow
- **THEN** Oktopus records the recommendation as an artifact or event
- **AND** does not start the recommended workflow unless selected by user, configured trigger, or policy-controlled router

### Requirement: Source/Blueprint/Proposal/Workflow Taxonomy

Oktopus SHALL distinguish sources, blueprints, proposals, active workflows, runs, jobs, and steps.

#### Scenario: Import blueprint

- **WHEN** Oktopus imports a reusable recipe from addyosmani/agent-skills commands, Archon catalog, built-in library, or org library
- **THEN** it represents the item as a workflow blueprint or equivalent inactive catalog item until instantiated or approved as an active workflow

#### Scenario: Store proposal

- **WHEN** Archon, an agent, or a human draft produces a new workflow candidate
- **THEN** Oktopus stores it as a proposal artifact or draft capability
- **AND** it does not expand into jobs until activated as an active workflow

#### Scenario: Expand active workflow only

- **WHEN** Oktopus creates a run
- **THEN** only an approved active workflow capability expands into jobs

### Requirement: Scoped Worker Handoff Context

Oktopus SHALL prepare a scoped context bundle before handing a job to a worker.

#### Scenario: Handoff agent job

- **WHEN** an agent worker leases a job
- **THEN** Oktopus provides job spec, resolved capability refs, allowed tools, skill/persona refs, scoped context bundle, input artifact refs, policy constraints, output contract, and evidence contract

#### Scenario: Limit context by policy and budget

- **WHEN** the context bundle is built
- **THEN** Oktopus limits content by configured source permissions, sensitivity, context budget, and relevance
- **AND** large knowledge source results are summarized or referenced as artifacts unless full content is explicitly allowed

### Requirement: Configurable Agent Runtime Workers

Oktopus SHALL support configurable background agent runtimes as worker capabilities.

#### Scenario: Register coding agent worker

- **WHEN** a Pi, Claude Code, Codex CLI, Antigravity CLI, OpenCode, Kiro, Aider, Goose, or similar coding agent process is available
- **THEN** it can register with Oktopus as an agent worker with runtime name, version, labels, supported skill formats, supported tools, and policy constraints

#### Scenario: Route job to selected runtime

- **WHEN** a workflow or policy selects a particular runtime for a job
- **THEN** Oktopus schedules the job only to compatible workers advertising that runtime and satisfying policy constraints
