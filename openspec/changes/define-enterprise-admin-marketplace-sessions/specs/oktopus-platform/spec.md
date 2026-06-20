## ADDED Requirements

### Requirement: Enterprise Capability Catalog

Oktopus SHALL support an administrator-managed enterprise capability catalog containing personas, skills, workflows, workflow blueprints, commands, tools, MCP servers, data sources, secret references, models, runtimes, sandbox/workspace profiles, verifiers, policies, approval chains, schedules, triggers, output destinations, and solution packs.

#### Scenario: Admin defines task standard

- **WHEN** an administrator defines a standard for a task such as repo standardization, CI remediation, security review, AIOps remediation, documentation maintenance, or VLM deployment
- **THEN** Oktopus records the required workflow, personas, skills, tools, MCP servers, data sources, output destinations, verifiers, policies, approvals, and workspace profile as versioned capabilities

#### Scenario: Capability declares governance metadata

- **WHEN** a capability is published to the enterprise catalog
- **THEN** it declares owner, scope, version, source/provenance, permissions, required secrets, allowed data sources, supported runtimes/models, sandbox requirements, risk level, approvals, eval evidence, security review status, and lifecycle status

### Requirement: Internal Marketplace

Oktopus SHALL provide an internal marketplace for distributing approved capability packages.

#### Scenario: Publish marketplace package

- **WHEN** a capability package passes schema validation, sandbox validation, security review, and approval
- **THEN** Oktopus can publish it to the internal marketplace for installation at organization, client, team, or project scope

#### Scenario: Prevent unapproved marketplace sharing

- **WHEN** a user or agent creates a new persona, workflow, skill, tool, MCP bundle, data connector, policy, or solution pack
- **THEN** Oktopus treats it as a proposal until marketplace review and approval complete

### Requirement: Standardized Workflow Mode

Oktopus SHALL support locked standardized workflows for governed repeatable tasks.

#### Scenario: Run standardized workflow

- **WHEN** a user invokes a governed task standard
- **THEN** Oktopus uses the administrator-approved workflow, required reviewers, verifiers, tools, policies, output destinations, and workspace profile
- **AND** users cannot silently remove required gates or expand authority beyond policy

#### Scenario: Protect regulated task

- **WHEN** a task involves production remediation, security review, client data processing, source mutation, capability activation, or deployment
- **THEN** Oktopus can require standardized workflow mode and deny user-custom workflow substitution

### Requirement: User-Custom Agent Mode

Oktopus SHALL allow users to customize their own agents for non-standardized or exploratory work within policy boundaries.

#### Scenario: User customizes agent

- **WHEN** a user creates a personal or team agent profile
- **THEN** Oktopus allows customization of persona, model/provider, runtime, skills, allowed tools, knowledge sources, memory scope, output destinations, working style, schedules, and notifications within policy limits

#### Scenario: Prevent policy bypass by customization

- **WHEN** a user-custom agent requests unauthorized secrets, tools, data sources, network access, filesystem mutation, production access, or client data
- **THEN** Oktopus denies the request or creates an approval gate according to policy

#### Scenario: Promote custom agent to marketplace

- **WHEN** a user wants to share a custom agent profile or workflow broadly
- **THEN** Oktopus converts it into a marketplace proposal requiring validation, review, and approval before publication

### Requirement: Resumable Thread Session Capsules

Oktopus SHALL model long-running conversation threads as resumable session capsules containing memory, state, artifacts, permissions, input bindings, output bindings, and workspace references.

#### Scenario: Resume thread

- **WHEN** a user resumes a prior thread
- **THEN** Oktopus restores thread metadata, memory pointers, context snapshots, artifact references, approvals, policy decisions, input/output channel bindings, active run references, and workspace reference where applicable

#### Scenario: Archive thread

- **WHEN** a thread reaches retention limit or is manually archived
- **THEN** Oktopus preserves required audit artifacts and events according to retention policy
- **AND** releases or snapshots associated workspace resources according to workspace policy

### Requirement: Scoped Memory Model

Oktopus SHALL support scoped and governed memory for threads, users, teams, projects, clients, organizations, and external knowledge sources.

#### Scenario: Write thread memory

- **WHEN** an agent or user records memory in a thread
- **THEN** Oktopus stores the memory with scope, provenance, sensitivity, retention, and access policy metadata

#### Scenario: Prevent cross-client memory leak

- **WHEN** a worker builds context for a client-scoped job
- **THEN** Oktopus excludes memories and knowledge sources from other client scopes unless policy explicitly permits access

### Requirement: Input Channels Beyond Stdin

Oktopus SHALL support input channels beyond CLI/stdin.

#### Scenario: Receive channel input

- **WHEN** Oktopus receives input from CLI, web UI, API, chat, email, webhook, CI/CD event, telemetry alert, ticket update, schedule, message queue, document update, or human form
- **THEN** it records source, actor, thread ID where applicable, payload, attachments, correlation ID, security context, and received timestamp

### Requirement: Output Destinations Beyond Stdout

Oktopus SHALL support configured output destinations beyond stdout.

#### Scenario: Push result to designated space

- **WHEN** a workflow declares output destinations such as chat channel, GitHub/GitLab PR, issue comment, docs repo, object store, dashboard, incident space, ADR repository, runbook repository, vector memory, email, or client portal
- **THEN** Oktopus delivers allowed outputs to those destinations and records delivery events

#### Scenario: Redact sensitive output

- **WHEN** output contains sensitive data
- **THEN** Oktopus applies redaction and access policy before delivery outside the run artifact store

### Requirement: Preconfigured Interactive Workspaces

Oktopus SHALL support preconfigured VM/workspace profiles for interactive sessions.

#### Scenario: Start interactive workspace

- **WHEN** a user starts an interactive thread requiring a governed environment
- **THEN** Oktopus provisions or attaches a workspace from an approved profile with required tools, agent runtimes, MCP servers, data sources, secret references, network policy, filesystem policy, resource limits, and snapshot policy

#### Scenario: Resume interactive workspace

- **WHEN** a user resumes an interactive thread with a workspace reference
- **THEN** Oktopus reconnects to the workspace or restores it from snapshot according to workspace persistence policy

#### Scenario: Standard workflow requires fixed workspace

- **WHEN** an administrator standard requires a specific workspace profile
- **THEN** Oktopus uses that profile and prevents users from substituting an unapproved environment

### Requirement: Multiple Run Modes

Oktopus SHALL support one-shot, scheduled, event-triggered, interactive, and background-loop run modes.

#### Scenario: Run scheduled task

- **WHEN** a schedule fires
- **THEN** Oktopus creates a run using the configured workflow, inputs, policies, and output destinations

#### Scenario: Run long interactive thread

- **WHEN** a user engages in a long-running interactive session
- **THEN** Oktopus preserves thread state, memory, artifacts, approvals, and workspace references across turns and pauses

### Requirement: Enterprise Governance Controls

Oktopus SHALL support enterprise governance controls for RBAC/ABAC, tenant/client boundaries, secret manager integration, data loss prevention, audit logs, policy-as-code, approval chains, budgets, model/provider routing, sandbox profiles, retention/legal hold, PII redaction, eval gates, observability/SLOs, incident controls, capability signing, provenance, and marketplace review.

#### Scenario: Enforce enterprise boundary

- **WHEN** a run, worker, tool, knowledge source, output destination, or workspace requests access across tenant, client, project, or user boundaries
- **THEN** Oktopus evaluates policy before allowing access and records the decision
