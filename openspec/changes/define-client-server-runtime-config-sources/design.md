## Design

### Goal

Define Oktopus as a configurable client/server agentic operations platform.

The server owns:

- configuration resolution
- capability registry
- skill/tool/persona/workflow source loading
- knowledge source indexing/query hooks
- routing from messages/commands/triggers to workflows
- run/job state
- scheduler and leases
- policy, approvals, and secrets
- artifacts, events, and observability

Background agents and tools act as workers:

- Pi worker
- Claude Code worker
- Codex CLI worker
- Antigravity CLI worker
- OpenCode worker
- Kiro worker
- Aider worker
- Goose worker
- shell worker
- tool worker
- verifier worker
- review worker
- inference worker
- knowledge worker

Workers execute bounded jobs under server configuration and policy.

### System Shape

```text
Clients / Inputs
  CLI | Web UI | API | Chat/message | CI/CD webhook | telemetry alert | schedule
        │
        ▼
Oktopus Server
  ├── Router: intent/command/trigger → workflow selection
  ├── Config Resolver: builtin → org → client → project → run
  ├── Registry: skills, tools, personas, workflows, runtimes, adapters, verifiers, policies
  ├── Knowledge Sources: vaults, docs, repos, vector DBs, APIs
  ├── Policy: approvals, secrets, network, filesystem, budgets, data boundaries
  ├── Scheduler: DAG jobs, leases, retries, cancellation
  ├── State: runs, jobs, attempts, workers, approvals, policy decisions
  └── Evidence: events, logs, artifacts, verifier results
        │
        ▼
Background Workers
  Pi | Claude | Codex | Kiro | shell | tools | verifiers | review | inference | knowledge
```

### Input Surfaces

#### Imperative Commands

Commands are explicit and bypass most intent ambiguity.

Examples:

```text
oktopus runs create repo-standardization --project network-manager --apply=false
/run repo-standardization --repo network-manager
POST /runs { workflow: "repo-standardization", inputs: {...} }
```

The router validates the requested workflow, inputs, policy, and permissions, then creates a run.

#### Messages

Messages are natural language requests from chat, web UI, CLI, or API.

Examples:

```text
"Standardize this repo and open a PR"
"CI is failing on network-manager; diagnose and propose fix"
"Investigate this alert and draft remediation"
```

The router may:

1. classify intent
2. retrieve relevant routing config
3. suggest workflow
4. ask clarification if confidence is low or action is risky
5. create run only after policy/user selection allows it

#### Triggers

Triggers are configured event sources.

Examples:

```text
github.pull_request
ci.failed
prometheus.alert
grafana.alert
schedule.daily
docs.changed
```

Triggers map to workflows through policy-controlled routing rules.

### Configurable Sources

Source definitions are declarative and scoped.

```yaml
skill_sources:
  - name: addy-agent-skills
    type: git
    url: https://github.com/addyosmani/agent-skills
    ref: a5f0b176381e9fea24a61aefc243506686aa2435
    paths:
      skills: skills/*/SKILL.md
      personas: agents/*.md
      commands: commands/*.toml

  - name: convexideas-skills
    type: filesystem
    path: ./registry/skills
```

```yaml
tool_sources:
  - name: local-tools
    type: filesystem
    path: ./registry/tools

  - name: mcp-tools
    type: mcp
    servers:
      - github
      - filesystem
      - browser

  - name: agent-cli-tools
    type: path
    allowlist:
      - graphify
      - openspec
      - archon
      - beads
```

```yaml
knowledge_sources:
  - name: obsidian-vault
    type: filesystem
    path: ~/Vaults/ConvexIdeas
    index: markdown

  - name: docs-repo
    type: git
    url: git@github.com:convexideas/docs.git
    ref: main

  - name: engineering-vector-db
    type: qdrant
    collection: engineering

  - name: confluence
    type: api
    connector: confluence
```

```yaml
runtime_sources:
  - name: local-agent-runtimes
    type: builtin
    runtimes:
      - pi
      - claude-code
      - codex
      - kiro
      - shell
```

### Source Kinds

Oktopus should support these configurable source kinds over time:

```text
filesystem
git
npm
oci
http
mcp
api
builtin
vector_db
secret_manager
message_bus
```

Not every source kind is required in MVP. Source contracts should be stable enough to add adapters later.

### Config Hierarchy

Configuration resolves in this order (the canonical precedence ladder, shared with capability resolution in `define-capability-registry-model`):

```text
builtin
  → upstream_pack
  → organization
  → client
  → project
  → workflow
  → run override
```

This is the single authoritative precedence vocabulary for both configuration resolution and capability resolution. `upstream_pack` applies to imported capability packs; it has no effect on non-capability configuration categories but is retained in the ladder so the order is identical everywhere.

Higher scope can narrow permissions. Broadening permissions requires policy and approval when moving into org/client/project scope.

Configuration categories:

```text
sources
routing
agents/runtimes
policies
secrets
knowledge
budgets
artifact retention
approval rules
```

Example:

```yaml
agents:
  default_coding: pi
  high_risk_review: claude-code
  fast_review: codex

routing:
  intents:
    repo_standardization:
      workflow: repo-standardization
    failed_ci:
      workflow: ci-remediation

policies:
  unknown_workflow: ask
  source_mutation: approval_required
  secrets: deny_by_default
```

### Routing and Workflow Selection

Active workflow selection can come from:

1. explicit user command
2. configured trigger rule
3. policy-controlled router recommendation
4. approved run override

Workers do not select global workflows. Agents may recommend workflows, but recommendations are artifacts/events until selected by server policy or user command.

### Source, Blueprint, Proposal, Workflow Taxonomy

Use this taxonomy:

```text
Source      = where capabilities come from: git, filesystem, MCP, vault, API, etc.
Blueprint   = reusable imported recipe/catalog item, e.g. addyosmani /ship or Archon catalog entry
Proposal    = inactive generated candidate, e.g. Archon/agent draft workflow
Workflow    = approved active executable Oktopus DAG
Run         = one workflow execution
Job         = worker-leased DAG node
Step        = observable action inside a job
```

Rules:

1. Only active `workflow` capabilities expand into jobs.
2. Blueprints can be instantiated into workflow proposals or active workflows after approval.
3. Proposals remain inactive until validation, review, and activation.
4. Archon, agents, and imported packs may produce blueprints/proposals but do not auto-activate workflows.

### Knowledge Sources and Context Scoping

Knowledge sources provide context, not authority by default.

Knowledge source examples:

- Obsidian vaults
- markdown docs
- OpenSpec specs
- repo READMEs
- service catalogs
- vector DBs
- runbooks
- telemetry stores
- tickets/incidents
- prior run artifacts

Context flow:

```text
job created
  → context policy determines allowed sources
  → retrieval/query selects relevant chunks/artifacts
  → context packer summarizes or references large material
  → worker receives scoped context bundle
```

Rules:

1. Workers receive only allowed context for the job.
2. Knowledge source results are captured as artifacts or event references when used materially.
3. Sensitive knowledge sources require policy checks and redaction.
4. Source code, explicit user instructions, and approved specs outrank stale knowledge base content.

### Background Agent Workers

Agent workers register runtime capabilities.

Examples:

```yaml
worker:
  kind: agent
  runtime: pi
  labels:
    os: darwin
    sandbox: local
  capabilities:
    tools: [read, bash, edit]
    skill_formats: [agent-skills]
```

Job handoff to worker includes:

```text
job spec
resolved capability refs
allowed tools
skill/persona refs
scoped context bundle
input artifact refs
policy constraints
output contract
evidence contract
```

Worker returns:

```text
events
logs
artifacts
verifier results
status
summary
```

### Server API Direction

Local CLI remains first client. Server API should later expose:

```text
POST /runs
GET /runs/{id}
GET /runs/{id}/events
GET /runs/{id}/artifacts
POST /messages
POST /commands
POST /workers/register
POST /jobs/request
POST /jobs/{id}/accept
POST /knowledge/query
```

### Design Constraints

- Oktopus server is system of record.
- Background agents are workers, not autonomous controllers.
- Skill/tool/knowledge sources are configurable and policy-scoped.
- Messages can request work; commands specify work.
- Active workflows are approved capabilities.
- Blueprints and proposals are not automatically executable.
- Context is scoped by policy before worker handoff.
