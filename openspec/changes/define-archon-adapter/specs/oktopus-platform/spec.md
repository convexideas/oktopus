## ADDED Requirements

### Requirement: Archon Adapter Capability

Oktopus SHALL represent Archon as a governed adapter and tool capability rather than as the platform control plane.

#### Scenario: Register Archon capability

- **WHEN** the Oktopus registry includes Archon
- **THEN** Archon is represented as an adapter or tool capability with declared entrypoint, permissions, outputs, and policy constraints

#### Scenario: Run Archon under Oktopus control

- **WHEN** a job invokes Archon
- **THEN** Oktopus still owns the job lease, worker identity, events, logs, artifacts, verifier gates, policy decisions, approvals, and final acceptance

### Requirement: Archon Proposal Generation

Oktopus SHALL allow Archon to generate capability proposals as artifacts.

#### Scenario: Generate workflow proposal

- **WHEN** Archon creates a proposed workflow, skill, persona, verifier, policy, adapter, runtime, tool, or solution pack
- **THEN** Oktopus stores the output as a proposal artifact linked to the producing run, job, attempt, and worker
- **AND** the proposed capability remains inactive

#### Scenario: Capture proposal metadata

- **WHEN** a proposal artifact is created by Archon
- **THEN** Oktopus records producer, source task, requested actor, creation time, risk summary, permission diff where available, and provenance metadata

### Requirement: No Auto-Install for Archon Workflows

Oktopus SHALL NOT allow Archon or agent workers to automatically install, activate, or route future runs through generated workflows or capabilities.

#### Scenario: Block silent workflow activation

- **WHEN** Archon generates or imports a workflow
- **THEN** Oktopus does not write it to the active registry or select it for future triggers by default
- **AND** it remains a proposal until validation, review, and approval complete

#### Scenario: Block silent trigger mutation

- **WHEN** Archon proposes a workflow with trigger rules
- **THEN** Oktopus does not activate or modify trigger rules without explicit approval

### Requirement: Archon Capability Activation Gates

Oktopus SHALL require validation and approval before Archon-generated proposals become active capabilities.

#### Scenario: Activate Archon-generated capability

- **WHEN** a user requests activation of an Archon-generated capability proposal
- **THEN** Oktopus validates schema, resolves dependencies, computes permission diff, runs sandbox dry-run where applicable, performs adversarial review when risk rules require it, and records human or CI approval before activation

#### Scenario: Reject risky proposal

- **WHEN** validation, policy, adversarial review, sandbox dry-run, or approval rejects an Archon-generated proposal
- **THEN** Oktopus keeps the proposal artifact inactive
- **AND** records rejection reason and evidence

### Requirement: Archon Explicit Workflow Execution

Oktopus SHALL allow explicit Archon workflow execution as a bounded job when selected by user, workflow, or policy.

#### Scenario: Execute known Archon workflow

- **WHEN** a workflow job explicitly invokes Archon in execution mode
- **THEN** Oktopus captures Archon command, inputs, stdout, stderr, structured output, reports, and generated artifacts
- **AND** applies normal job timeout, lease, policy, verifier, and artifact rules

#### Scenario: Prevent autonomous workflow start

- **WHEN** Archon recommends another workflow
- **THEN** Oktopus records the recommendation as an artifact or event
- **AND** does not start the recommended workflow unless selected by user, configured trigger, or policy-controlled router

### Requirement: Archon Catalog Import as Draft

Oktopus SHALL import Archon workflow catalogs as draft proposal artifacts unless explicit activation is requested.

#### Scenario: Import catalog entry

- **WHEN** Oktopus imports Archon catalog workflows
- **THEN** each imported entry becomes a draft proposal with source and provenance metadata
- **AND** no entry becomes active until the capability activation gates pass

### Requirement: Policy Checks for Archon Invocation

Oktopus SHALL evaluate policy before invoking Archon.

#### Scenario: Check Archon permissions

- **WHEN** an Archon job requests filesystem writes, network access, secrets, source mutation, registry writes, trigger mutation, capability activation, or budget beyond default limits
- **THEN** Oktopus allows, denies, or requires approval according to policy
- **AND** records a policy decision event

### Requirement: Adversarial Review for Risky Archon Proposals

Oktopus SHALL support adversarial review of Archon-generated proposals that request elevated authority.

#### Scenario: Review risky workflow proposal

- **WHEN** an Archon-generated proposal requests secrets, network egress, source mutation, production remediation, workflow triggers, new tool grants, external APIs, or client data access
- **THEN** Oktopus can require adversarial review jobs before activation approval
- **AND** review findings become artifacts or structured events linked to the proposal
