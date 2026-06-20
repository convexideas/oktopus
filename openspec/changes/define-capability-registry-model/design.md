## Design

### Goal

Define Oktopus capability registry as the source of truth for what the platform can do and under what governance.

The registry should support:

- Local-first manifests.
- Imported upstream packs.
- Organization-level overrides.
- Client-specific solution packs.
- Runtime/tool/persona/workflow composition.
- Capability provenance and version pinning.
- Review gates before activation.
- Distributed registry backends later.

### Capability Kinds

```text
tool          = callable action: CLI/API/MCP/script/container/service
skill         = reusable procedure/playbook
skillpack     = imported collection of skills/personas/workflows/references
persona       = role/instruction bundle with allowed tools and output format
workflow      = DAG/state machine composed of jobs and gates
verifier      = evidence-producing check
environment   = execution boundary: local, worktree, Docker, OpenShell, VM, Kubernetes
policy        = permissions, approvals, secrets, budgets, network/filesystem rules
runtime       = agent/model execution backend: Pi, Claude Code, Codex CLI, Antigravity CLI, OpenCode, Kiro, Aider, Goose, shell, VLM service
adapter       = integration bridge to CI/CD, telemetry, docs, ticketing, git, artifact stores
solution_pack = domain offering composed from capabilities
```

### Manifest Envelope

Capability manifests are YAML-first. TOML may be accepted for compatibility with early local seed manifests, but YAML is the default authoring format because it supports nested workflows, policies, and solution packs cleanly.

Every capability manifest uses a common envelope.

```yaml
kind: tool
name: graphify
version: 0.1.0
description: Build code/content knowledge graph artifacts.
source:
  type: git|npm|local|oci|http|builtin
  uri: https://github.com/example/graphify-pack
  ref: commit-or-version
trust:
  level: untrusted|trusted|org|builtin
  signed: false
  reviewed_by: []
status: draft|validated|active|deprecated|archived
metadata: {}
```

### Tool Manifest

```yaml
kind: tool
name: graphify
version: 0.1.0
entrypoint:
  type: cli|api|mcp|script|container|runtime
  command: graphify
inputs:
  schema: ./schemas/graphify-input.schema.json
outputs:
  artifacts:
    - graphify-out/GRAPH_REPORT.md
    - graphify-out/graph.json
permissions:
  filesystem: read_project
  network: false
  secrets: none
verifiers:
  - artifact-exists:graphify-out/graph.json
```

### Skill and Skillpack Manifest

Agent Skills-compatible directories can be imported without rewriting their contents.

```yaml
kind: skillpack
name: addyosmani-agent-skills
version: 0.1.0
source:
  type: git
  uri: https://github.com/addyosmani/agent-skills
  ref: a5f0b176381e9fea24a61aefc243506686aa2435
imports:
  skills: ["skills/*/SKILL.md"]
  personas: ["agents/*.md"]
  workflows: ["commands/*.toml", ".claude/commands/*.md", ".gemini/commands/*.toml"]
  references: ["references/*.md"]
policy:
  imported_capabilities: read_only
  custom_overrides_allowed: true
```

### Persona Manifest

```yaml
kind: persona
name: security-auditor
version: 0.1.0
role: Security Engineer
instructions_ref: ./security-auditor.md
allowed_tools:
  - git-read
  - dependency-audit
  - artifact-read
skills:
  - security-and-hardening
output_contract:
  type: markdown_report
composition:
  can_spawn_personas: false
  can_mutate_source: false
```

### Workflow Manifest

```yaml
kind: workflow
name: guarded-build
version: 0.1.0
steps:
  - id: build
    kind: agent
    runtime: pi
    uses: skill:incremental-implementation
  - id: correctness-review
    kind: adversarial_review
    needs: [build]
    persona: code-reviewer
  - id: security-review
    kind: adversarial_review
    needs: [build]
    persona: security-auditor
  - id: synthesize
    kind: synthesis
    needs: [correctness-review, security-review]
gates:
  - requires: synthesize.accepted
```

### Solution Pack Manifest

A solution pack is the packaging unit for Convex Ideas offerings.

```yaml
kind: solution_pack
name: vlm-quality-harness
version: 0.1.0
domain: vision-language-models
includes:
  workflows:
    - dataset-intake
    - golden-set-eval
    - human-review
    - deployment-gate
    - drift-monitoring
  tools:
    - image-dataset-loader
    - vlm-evaluator
    - confusion-matrix-reporter
  verifiers:
    - accuracy-threshold
    - latency-budget
    - bias-check
  policies:
    - pii-redaction
    - client-data-boundary
```

### Lifecycle

```text
draft
  → schema_validated
  → sandbox_validated
  → reviewed
  → active
  → deprecated
  → archived
```

Activation gates:

1. Schema validation.
2. Dependency resolution.
3. Permission diff review.
4. Sandbox dry-run when executable.
5. Human or CI approval for org/client scope.
6. Signature or pinned source ref for shared registry.

### Override and Precedence Rules

Registry scopes (canonical precedence ladder, shared with configuration resolution in `define-client-server-runtime-config-sources`):

```text
builtin < upstream_pack < org < client < project < workflow < run_override
```

This is the single authoritative precedence vocabulary for both capability resolution and configuration resolution. Not every tier is present in every deployment — local MVP typically uses only `builtin`, `upstream_pack`, `project`, and `run_override` — but the order is fixed and tiers are never reordered.

Rules:

1. Higher scope may override lower scope by name and kind.
2. Overrides must record parent capability and reason.
3. Workflows must resolve exact versions at run creation. `resolved_capabilities_json` on the Run record is the authoritative snapshot. Workers use pinned refs from the lease grant, not the live registry.
4. Run records keep resolved capability refs forever for audit.
5. Deprecated capabilities remain resolvable for old runs but unavailable for new runs unless explicitly allowed.

`run_override` scope: a run-level override is represented as an entry in `resolved_capabilities_json` with `scope: run_override` and must include `parent_capability_ref` and `reason`. Run overrides are set at run creation only; they cannot be modified after the run starts.

### Agent-Created Capability Proposals

Agents may generate capability proposals. Proposals are artifacts only — they are never active capabilities.

**Enforcement boundary:** Workers only get write access to `runs/<run-id>/jobs/<job-key>/attempts/<n>/`. The `registry/` directory is outside all worker filesystem scopes. Workers cannot write to `registry/` directly. The controller is the only process that writes active registry entries.

Proposal artifact locations (worker-writable):

```text
runs/<run-id>/jobs/<job-key>/attempts/<n>/artifacts/proposals/<kind>/<name>.yaml
```

Intake command (human or CI invokes; controller mediates):

```bash
oktopus registry propose --from <artifact-path> [--kind <kind>] [--name <name>]
```

This command:
1. Copies the artifact to `registry/proposals/<kind>/<name>.yaml`
2. Runs schema validation
3. Emits `capability.proposal_created`
4. Creates a `proposal_validation` run using a built-in review workflow

Activation command (requires prior approval):

```bash
oktopus registry activate <kind>:<name>[@version] --approval <approval-id>
```

No proposal becomes active without explicit execution of this command. No automation path bypasses it.

Proposal flow:

```text
agent writes proposal artifact (in runs/ dir only)
→ human runs: oktopus registry propose --from <artifact-path>
→ registry lints schema
→ sandbox evaluates behavior
→ adversarial review checks safety/usefulness
→ human/CI approves (approval-id issued)
→ human runs: oktopus registry activate <ref> --approval <approval-id>
→ capability becomes active in registry/
```

No agent-created capability is live by default.

### Registry Storage

Local MVP:

```text
registry/**/*.yaml|toml
SQLite index for resolved capabilities
run record stores resolved capability refs
```

Distributed future:

```text
signed git registry or OCI registry
Postgres capability index
object store for capability assets
policy engine for activation gates
```

### Initial Capability Packs

Seed registry should include:

- `addyosmani-agent-skills` skillpack.
- `graphify` tool.
- `openspec` tool.
- `pi` runtime adapter.
- `shell` runtime adapter.
- `artifact-exists` verifier.
- `sdlc-default` workflow.
- `guarded-build` workflow with adversarial review.

### Design Constraints

- Registry is declarative first; executable behavior lives behind adapters.
- Capabilities are versioned and immutable once active.
- Policy is enforced by controller/workers, not by prompt instruction alone.
- Local developer ergonomics must remain simple: file manifests and CLI validation.
- Distribution must allow client-specific packs without forking core platform.
