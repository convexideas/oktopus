## Design

### Goal

Add concrete extension mechanics inspired by multi-agent configurator systems, adapted for Oktopus as an enterprise control plane.

The design covers:

- runtime adapter contract
- runtime capability matrix
- skill source index
- delegation policy rules
- staged capability install/activation pipeline
- package presets/profiles

### Runtime Adapter Contract

A runtime adapter translates an Oktopus job into a native execution mechanism for a specific agent/runtime.

Examples:

```text
Pi
Claude Code
Codex
Kiro
OpenCode
Gemini CLI
Hermes
shell
container
inference service
```

Adapter responsibilities:

```text
Detect/Register
  discover runtime availability, version, config scope, supported features

Advertise Capabilities
  skills, MCP support, subagents, slash commands, streaming, artifacts, model overrides, policy hooks

Prepare Handoff
  convert Oktopus job spec into runtime-native prompt/config/context

Execute Job
  run bounded work under lease and policy

Stream Observations
  logs, events, model/tool output summaries, progress updates

Collect Evidence
  artifacts, verifier results, review findings, output destination records

Finalize
  complete/fail/cancel attempt with lease validation
```

Adapter interface shape:

```text
RuntimeID() string
Detect(ctx) RuntimeDetection
Capabilities(ctx) RuntimeCapabilities
Prepare(ctx, JobHandoff) RuntimeInvocation
Execute(ctx, Lease, RuntimeInvocation) RuntimeResult
Cancel(ctx, Lease) error
```

Adapters do not own global workflow choice, durable memory, active capability registry, or approval policy.

### Runtime Capability Matrix

Administrators need to know what each runtime supports.

Capability dimensions:

```text
skills
MCP
subagents/delegation
slash commands
model override
streaming output
artifact writeback
workspace/project-local config
global config
permission hooks
native approvals
background execution
interactive mode
```

Example matrix row:

```yaml
runtime: pi
skills: true
mcp: true
subagents: true
slash_commands: true
model_override: true
artifact_writeback: adapter
permission_hooks: adapter
interactive_mode: true
```

Workflow selection may use this matrix:

```text
high-risk review → runtime must support isolated review context
interactive thread → runtime must support interactive mode or workspace binding
strict policy job → runtime must support policy-aware tools or run inside sandbox
```

### Skill Source Index

Oktopus should index skills without rewriting them.

Index fields:

```text
id
name
description
source_name
source_type
source_uri
source_ref
scope
path
format
trust_level
version
hash
metadata
```

Rules:

1. The index stores full descriptions and exact `SKILL.md` paths or source URIs.
2. Project/client/org scope precedence is deterministic.
3. Matching uses descriptions and metadata, but execution loads the exact original skill content.
4. Imported skills remain read-only unless overridden by higher-scope manifests.
5. Summaries may be cached, but original skill content remains source of truth.

Runtime handoff includes selected skill references:

```text
skills_to_load:
  - source: addyosmani-agent-skills
    name: security-and-hardening
    path: /.../skills/security-and-hardening/SKILL.md
    hash: sha256:...
```

### Delegation Policy Rules

Delegation should be policy-driven, not left entirely to agent mood.

Rule examples:

```yaml
delegation_rules:
  exploration:
    when_file_reads_gte: 4
    action: create_exploration_job

  multi_file_write:
    when_nontrivial_files_touched_gte: 2
    action: require_single_writer_and_fresh_review

  pull_request:
    before_pr_after_code_change: true
    action: require_fresh_review

  incident:
    when_wrong_cwd_or_git_recovery_or_env_confusion: true
    action: require_fresh_audit

  long_session:
    when_tool_calls_gte: 20
    action: pause_replan_or_delegate
```

Oktopus maps delegation rules to workflow behavior:

```text
create exploration job
create adversarial review job
block until verifier passes
request human approval
record waiver if bypassed
```

### Staged Capability Install/Activation Pipeline

Installing or activating a capability is operational work.

Pipeline:

```text
plan
  resolve package, dependencies, permissions, target scopes

snapshot
  capture current active capability/config state

apply
  write draft capability records or install package assets

verify
  schema, dependency, policy, sandbox, smoke, verifier checks

review
  adversarial/security review if risk rules require it

activate
  mark capability active at requested scope

rollback
  restore snapshot if apply/verify/activation fails

audit
  record events, artifacts, policy decisions, approval records
```

The pipeline applies to:

```text
runtime adapters
skills
personas
workflows
MCP bundles
tools
knowledge connectors
workspace profiles
solution packs
marketplace packages
```

### Presets and Package Profiles

Packages can expose install profiles.

Examples:

```text
repo-governance/basic
repo-governance/strict
aiops/read-only
aiops/remediation-enabled
vlm/eval-only
vlm/deploy-gated
```

Profile manifest:

```yaml
kind: profile
name: repo-governance/strict
package: solution_pack:repo-governance
includes:
  workflows:
    - repo-standardization
    - openspec-review
  tools:
    - graphify
    - openspec
  verifiers:
    - command-exit-zero
    - artifact-exists
policies:
  source_mutation: approval_required
  fresh_review_before_pr: true
```

Rules:

1. Profiles are convenience bundles, not separate authority boundaries.
2. Profile activation still expands into normal capability activation pipeline.
3. Admins can install profiles at org/client/team/project scope.
4. Users can request profile installation where policy allows.
5. Profile choices are recorded in effective configuration for runs.

### Configurator vs Control Plane Boundary

Oktopus may configure workers or runtimes, but should avoid becoming only a local dotfile/config manager.

Boundary:

```text
Configurator pattern useful for:
  detecting runtimes
  writing adapter-specific setup when explicitly approved
  installing local worker dependencies
  building capability indexes

Control-plane responsibility remains:
  active registry
  workflow selection
  run/job state
  policies
  approvals
  artifacts
  memory/context
  audit
```

### Design Constraints

- Runtime native features are used, not flattened away.
- Original skills remain source of truth.
- Delegation rules are configurable and waivable with audit.
- Capability activation is staged and reversible where possible.
- Profiles/presets simplify installation but do not bypass policy.
