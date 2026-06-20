## Design

### Goal

Define the enterprise product model for Oktopus.

Oktopus should support:

- administrator-defined standard capabilities
- internal marketplace for personas, workflows, skills, MCPs, tools, data connectors, commands, policies, verifiers, runtimes, models, and solution packs
- user-customizable agents and personal/team configurations
- locked standardized workflows for governed tasks
- one-shot ad hoc tasks
- scheduled jobs
- event-triggered jobs
- long-running resumable conversation threads
- preconfigured interactive VM/workspace environments
- multiple input channels and output destinations
- enterprise governance, audit, and data boundaries

### Enterprise Capability Catalog

Administrators can manage enterprise standard capabilities:

```text
personas
skills
workflows
workflow blueprints
commands
tools
MCP servers
data/knowledge sources
secret references
models/providers
agent runtimes
sandbox/workspace profiles
verifiers/evals
policies
approval chains
schedules/triggers
output destinations/templates
solution packs
```

Each capability has:

```text
kind
name
version
owner
scope
source/provenance
permissions
required secrets
allowed data sources
supported runtimes/models
sandbox requirements
risk level
approval requirements
eval/verifier evidence
security review status
audit history
lifecycle status
```

### Marketplace

The internal marketplace distributes approved capability packages.

Package examples:

```text
repo-governance-pack
ci-remediation-pack
aiops-remediation-pack
documentation-maintainer-pack
vlm-quality-harness-pack
client-support-agent-pack
security-review-pack
```

Package lifecycle:

```text
draft
  → schema_validated
  → sandbox_validated
  → security_reviewed
  → approved
  → marketplace_published
  → installed
  → deprecated
  → archived
```

Marketplace rules:

1. Marketplace packages are versioned.
2. Packages declare permissions and dependencies.
3. Admins can install packages at organization, client, team, or project scope.
4. Users can request package installation where allowed.
5. Agent-generated packages are proposals until approved.
6. Deprecated packages remain available for historical runs but not new default runs unless explicitly allowed.

### Standardized Workflows vs User Customization

Oktopus must support both governed standards and user flexibility.

#### Standardized Workflow Mode

For regulated or repeatable tasks, administrators define locked workflows.

Examples:

```text
production remediation
security review
client data processing
repo standardization
CI failure remediation
VLM deployment gate
```

Locked workflow properties:

```text
fixed workflow DAG
fixed required reviewers/verifiers
fixed policy gates
fixed output destinations
restricted tool/runtime set
restricted data sources
admin-controlled updates
```

Users can provide inputs but cannot silently change execution authority.

#### User-Custom Agent Mode

For exploratory or personal/team work, users can customize agents.

Customizable fields may include:

```text
persona
model/provider
runtime
skills
tools within allowed set
knowledge sources within allowed set
memory scope
output destination
working style
schedules
notification preferences
```

Constraints:

1. User customization cannot exceed policy boundaries.
2. User custom agents cannot access unauthorized secrets or data sources.
3. User custom agents cannot replace required standardized workflows for governed tasks.
4. User custom agents can be promoted to marketplace proposals after review.

### Thread / Session Capsule

A thread is a resumable session container.

Thread record:

```text
id
name
owner
scope
participants
mode
status
agent_config_refs
workflow_refs
runtime_refs
memory_scope
input_channels
output_destinations
workspace_ref
active_run_ids
artifact_refs
context_snapshots
approval_state
created_at
last_active_at
retention_policy
metadata
```

Thread modes:

```text
one_shot
interactive
scheduled
event_triggered
long_running
incident_room
research_loop
client_support
```

Thread states:

```text
created
active
waiting_user
waiting_approval
paused
suspended
completed
archived
expired
```

Thread capsule contents:

```text
conversation history
summaries/context snapshots
memories
run history
artifacts
approvals
policy decisions
tool/event history
input channel bindings
output destination bindings
workspace/VM state reference
```

Resume behavior:

```text
user selects thread
→ Oktopus restores thread metadata and memory pointers
→ loads latest context snapshot and relevant artifacts
→ reconnects allowed input/output channels
→ schedules next job or resumes interaction
```

### Memory Model

Memory is scoped and governable.

Memory scopes:

```text
thread
user
team
project
client
organization
external_knowledge_source
```

Rules:

1. Thread memory is resumable by default within retention policy.
2. Cross-thread memory requires explicit scope and policy.
3. Client data cannot leak across client boundaries.
4. Sensitive memory requires redaction/access policy.
5. Memory writes are events/artifacts, not hidden model-only state.
6. Users can inspect/export/delete memory where policy permits.

### Input Channels

Oktopus can receive inputs from:

```text
CLI
web UI
REST/gRPC API
Slack/Teams/Discord
email
webhook
CI/CD event
telemetry alert
ticket update
schedule/cron
message queue
document update
human form
```

Input records include:

```text
source
actor
thread_id
payload
attachments
correlation_id
security context
received_at
```

### Output Destinations

Oktopus can send outputs to designated spaces:

```text
stdout/CLI
web UI thread
Slack/Teams/Discord channel
email
GitHub/GitLab PR
issue/ticket comment
Confluence/Notion/docs repo
object store
artifact registry
dashboard
incident report space
ADR repository
runbook repository
vector memory
client portal
```

Output destination record:

```text
id
name
type
scope
connection_ref
policy_ref
template_ref
allowed_artifact_types
metadata
```

Rules:

1. Workflows declare allowed output destinations.
2. Sensitive outputs require redaction and policy checks.
3. Final run receipts are stored even when output is pushed elsewhere.
4. Delivery attempts emit events.

### Preconfigured Interactive Workspaces

Interactive tasks may require a preconfigured VM/workspace.

Workspace profile:

```text
id
name
image_ref
base_tools
agent_runtimes
MCP servers
mounted_data_sources
secret_refs
network_policy
filesystem_policy
resource_limits
persistence_policy
snapshot_policy
idle_timeout
owner_scope
```

Workspace examples:

```text
coding-agent-workspace
security-review-workspace
aiops-remediation-workspace
vlm-eval-workspace
client-support-workspace
```

Interactive flow:

```text
user starts interactive thread
→ Oktopus provisions workspace from profile
→ installs/attaches allowed tools, MCPs, skills, data sources, runtimes
→ binds thread to workspace_ref
→ agents and user can interact through approved channels
→ workspace snapshots memory/artifacts/state
→ session can pause/resume
```

Workspace rules:

1. Workspace profile is a governed capability.
2. Users may customize workspace within policy.
3. Standardized workflows may require fixed workspace profiles.
4. Secrets are injected by policy, not baked into images.
5. Workspace snapshots are artifacts or external state references.
6. Admins can revoke, suspend, or archive workspaces.

### Run Modes

```text
one_shot
  one request → one run → final output

scheduled
  trigger by schedule → run workflow → push output

event_triggered
  trigger by webhook/alert/ticket → run workflow

interactive
  long-running thread with user/agent exchange

background_loop
  autonomous bounded loop with checkpoints and stop rules
```

### Governance

Enterprise controls:

```text
RBAC/ABAC
tenant/client boundaries
secret manager integration
data loss prevention
audit logs
policy-as-code
approval chains
budget/quota controls
model/provider routing
sandbox/workspace profiles
retention/legal hold
PII redaction
eval gates
observability/SLOs
incident controls
capability signing/provenance
marketplace review workflow
```

### Admin UX Direction

Administrators can define task standards:

```yaml
task_standard:
  name: repo-standardization
  workflow: workflow:repo-standardization@3.0.0
  persona: persona:convexideas-platform-engineer@1.2.0
  skills:
    - skill:agent-tooling
    - skill:repo-governance
  tools:
    - tool:graphify
    - tool:openspec
    - tool:git
    - tool:nix
  mcp_servers:
    - mcp:github
    - mcp:filesystem
  data_sources:
    - knowledge:depot-docs
    - knowledge:org-standards-vault
  output_destinations:
    - github_pr
    - slack_platform_channel
  approvals:
    source_mutation: required
```

Users invoke the standard:

```text
/run repo-standardization network-manager
"standardize this repo"
```

Oktopus resolves the standard configuration and creates governed runs.

### Design Constraints

- Standardized workflows are admin-governed.
- User-custom agents are allowed within policy boundaries.
- Long-running threads are resumable containers of memory, state, artifacts, permissions, and I/O bindings.
- Interactive work should run in governed preconfigured workspaces/VMs where setup drift and local trust are concerns.
- Outputs go to configured destinations, not only stdout.
- Inputs come from configured channels, not only stdin.
- All capability installation and sharing goes through marketplace lifecycle.
