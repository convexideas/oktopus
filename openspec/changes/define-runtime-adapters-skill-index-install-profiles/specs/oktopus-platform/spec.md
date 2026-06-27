## ADDED Requirements

### Requirement: Runtime Adapter Contract

Oktopus SHALL define a runtime adapter contract for translating bounded Oktopus jobs into runtime-native execution mechanisms.

#### Scenario: Register runtime adapter

- **WHEN** a runtime adapter for Pi, Claude Code, Codex, Kiro, OpenCode, Gemini CLI, Hermes, shell, container, or inference service is available
- **THEN** it can report runtime identity, detection status, supported capabilities, config scope, labels, and execution constraints to Oktopus

#### Scenario: Execute job through adapter

- **WHEN** a worker leases a job for a specific runtime
- **THEN** the runtime adapter prepares a scoped invocation from the Oktopus job handoff, executes bounded work, streams observations, collects artifacts, and completes or fails the attempt through Oktopus lease validation

### Requirement: Runtime Capability Matrix

Oktopus SHALL maintain a runtime capability matrix for workflow selection, admin visibility, and policy checks.

#### Scenario: Select runtime by capability

- **WHEN** a workflow requires isolated review context, MCP support, model override, interactive mode, artifact writeback, or policy-aware execution
- **THEN** Oktopus schedules the job only to runtimes and workers advertising compatible capabilities

#### Scenario: Display runtime support

- **WHEN** an administrator inspects configured runtimes
- **THEN** Oktopus exposes which runtimes support skills, MCP, subagents, slash commands, model overrides, streaming, artifacts, workspace config, permission hooks, background execution, and interactive mode

### Requirement: Skill Source Index

Oktopus SHALL index skills from configured skill sources while preserving exact source paths and original skill content as the source of truth.

#### Scenario: Index skills from source

- **WHEN** Oktopus refreshes a skill source
- **THEN** it records each skill's name, full description, source name, source type, source URI, source revision, scope, exact path or URI, format, trust level, version, hash, and metadata

#### Scenario: Handoff selected skills to worker

- **WHEN** an agent or review job requires skills
- **THEN** Oktopus passes exact skill references to the worker rather than generated summaries
- **AND** the worker or adapter loads the original skill content according to runtime capabilities

#### Scenario: Preserve override precedence

- **WHEN** two skill sources define the same skill name
- **THEN** Oktopus applies deterministic scope precedence and records which skill source won for the run

### Requirement: Delegation Policy Rules

Oktopus SHALL support configurable delegation policy rules that turn complexity signals into workflow actions.

#### Scenario: Delegate exploration after broad file reads

- **WHEN** a task requires reading more files than the configured exploration threshold
- **THEN** Oktopus can create or require an exploration job before implementation continues

#### Scenario: Require review for multi-file writes

- **WHEN** a job touches multiple non-trivial files or prepares a pull request after code changes
- **THEN** Oktopus can require a fresh review, verifier, approval, or waiver before completion

#### Scenario: Pause long monolithic session

- **WHEN** a long-running interactive thread crosses configured complexity thresholds such as tool calls, exploratory reads, or non-mechanical edits
- **THEN** Oktopus can pause, re-plan, delegate, or request justification before continuing

### Requirement: Staged Capability Install and Activation Pipeline

Oktopus SHALL install and activate capabilities through a staged, auditable pipeline.

#### Scenario: Activate capability package

- **WHEN** an administrator installs or activates a runtime adapter, skill, persona, workflow, MCP bundle, tool, knowledge connector, workspace profile, solution pack, or marketplace package
- **THEN** Oktopus plans dependencies and permissions, snapshots current state, applies draft records or assets, verifies schema and policy, runs sandbox or smoke checks where applicable, performs review where required, activates the capability, and records audit events

#### Scenario: Roll back failed activation

- **WHEN** apply, verification, review, or activation fails
- **THEN** Oktopus keeps the capability inactive and rolls back or records required manual cleanup according to the capability's rollback support

### Requirement: Package Profiles and Presets

Oktopus SHALL support profiles or presets as install-time bundles for solution packs and marketplace packages.

#### Scenario: Install package profile

- **WHEN** an administrator installs a profile such as `repo-governance/basic`, `repo-governance/strict`, `aiops/read-only`, `aiops/remediation-enabled`, `vlm/eval-only`, or `vlm/deploy-gated`
- **THEN** Oktopus resolves the included workflows, tools, verifiers, policies, output destinations, and runtime requirements through the normal capability activation pipeline

#### Scenario: Record selected profile

- **WHEN** a run uses capabilities installed through a profile
- **THEN** Oktopus records the profile and capability versions in the effective configuration for reproducibility and audit

### Requirement: Configurator Boundary

Oktopus SHALL distinguish runtime/configuration helper behavior from control-plane authority.

#### Scenario: Configure runtime safely

- **WHEN** Oktopus or an adapter writes runtime-specific setup, indexes skills, or installs local worker dependencies
- **THEN** it does so only through explicit capability install or setup workflows with events, policy checks, and rollback where possible

#### Scenario: Preserve control-plane authority

- **WHEN** a runtime-native config, agent local file, or external orchestrator suggests a workflow or capability change
- **THEN** Oktopus treats it as input or proposal until the control plane validates and activates it
