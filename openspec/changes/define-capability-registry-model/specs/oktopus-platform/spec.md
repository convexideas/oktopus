## ADDED Requirements

### Requirement: Versioned Capability Registry

Oktopus SHALL maintain a versioned registry for tools, skills, skillpacks, personas, workflows, verifiers, environments, policies, runtimes, adapters, and solution packs.

#### Scenario: Resolve capability at run creation

- **WHEN** a run is created from a workflow
- **THEN** Oktopus resolves every referenced capability by kind, name, version, source, and revision
- **AND** stores the resolved capability references in the run record for audit and reproducibility

#### Scenario: Validate capability before activation

- **WHEN** a capability manifest is added or updated
- **THEN** Oktopus validates the manifest schema, dependencies, source provenance, and requested permissions before the capability can become active

### Requirement: Declarative Capability Manifests

Oktopus SHALL represent capabilities using declarative manifests with common envelope fields and kind-specific configuration.

#### Scenario: Parse common manifest envelope

- **WHEN** Oktopus loads a capability manifest
- **THEN** it reads common fields including kind, name, version, description, source, trust, status, and metadata
- **AND** it validates kind-specific fields according to the capability kind

#### Scenario: Reject malformed manifest

- **WHEN** a manifest is missing required fields or contains invalid kind-specific configuration
- **THEN** Oktopus rejects the manifest
- **AND** reports validation errors with file path and field context

### Requirement: Skillpack Import

Oktopus SHALL import Agent Skills-compatible skillpacks as read-only capability packs without rewriting upstream skill contents.

#### Scenario: Import upstream Agent Skills pack

- **WHEN** Oktopus imports a skillpack such as `addyosmani-agent-skills`
- **THEN** it records the upstream source, pinned revision, imported skill paths, persona paths, workflow/command paths, and reference paths
- **AND** imported capabilities are read-only unless explicitly overridden by higher-scope local manifests

### Requirement: Capability Scopes and Overrides

Oktopus SHALL resolve capabilities across scopes using deterministic precedence and audit every override.

#### Scenario: Apply project override

- **WHEN** a project-level capability overrides an upstream capability with the same kind and name
- **THEN** Oktopus selects the project-level capability for new runs in that project
- **AND** records the overridden parent capability and override reason

#### Scenario: Resolve using the canonical precedence ladder

- **WHEN** the same capability kind and name exists at multiple scopes
- **THEN** Oktopus resolves it using the fixed order `builtin < upstream_pack < org < client < project < workflow < run_override`
- **AND** this is the same precedence ladder used for configuration resolution

#### Scenario: Preserve old run reproducibility

- **WHEN** a capability is overridden, deprecated, or archived after a run was created
- **THEN** historical runs continue to reference the exact capability versions resolved at creation time

### Requirement: Capability Lifecycle Gates

Oktopus SHALL manage capability lifecycle states from draft to archived with validation and review gates before activation.

#### Scenario: Activate executable capability

- **WHEN** a tool, runtime, adapter, verifier, workflow, or solution pack requests activation
- **THEN** Oktopus requires schema validation, dependency resolution, permission review, sandbox dry-run where applicable, and human or CI approval for org/client scope

#### Scenario: Prevent unreviewed execution

- **WHEN** a capability is in draft, schema-validating, sandbox-validating, or review state
- **THEN** Oktopus does not allow production workflows to execute it unless an explicit run override permits experimental use

### Requirement: Agent-Created Capability Proposals

Oktopus SHALL allow agents to propose capabilities but SHALL NOT activate them automatically.

#### Scenario: Store proposal artifact

- **WHEN** an agent creates a new tool, skill, persona, workflow, verifier, environment, policy, runtime, adapter, or solution pack
- **THEN** Oktopus stores it as a proposal artifact linked to the run, job, and attempt that produced it
- **AND** the proposal remains inactive until validated, reviewed, and approved

#### Scenario: Review proposed capability adversarially

- **WHEN** a proposed capability requests permissions, secrets, network access, source mutation, deployment access, or activation into org/client scope
- **THEN** Oktopus can require adversarial review jobs before approval

### Requirement: Solution Packs

Oktopus SHALL support solution packs that compose workflows, tools, skills, personas, verifiers, environments, policies, runtimes, and adapters for a specific domain or client offering.

#### Scenario: Install VLM solution pack

- **WHEN** a client project installs a VLM quality harness solution pack
- **THEN** Oktopus makes its workflows, tools, verifiers, policies, and dashboards discoverable for that project
- **AND** the solution pack uses the same run, job, artifact, verifier, policy, and event primitives as other workflows

#### Scenario: Install engineering ops solution pack

- **WHEN** a Convex Ideas engineering project installs the engineering ops pack
- **THEN** Oktopus exposes spec, plan, build, test, review, ship, CI remediation, documentation, and adversarial review workflows as versioned capabilities

### Requirement: Policy-Aware Capability Use

Oktopus SHALL enforce declared capability permissions at runtime.

#### Scenario: Deny unauthorized tool capability

- **WHEN** a job attempts to use a tool or runtime capability outside its declared permissions or workflow policy
- **THEN** Oktopus denies the action or creates an approval request
- **AND** records a policy decision event
