## Design

### Goal

Integrate Archon as a governed workflow capability inside Oktopus.

Archon should help with:

- generating workflow proposals
- executing explicit structured AI workflows as bounded jobs
- importing workflow catalogs as inactive proposals
- documenting repeatable processes

Archon should not:

- decide active workflow selection without policy
- write active registry entries directly
- mutate trigger rules silently
- grant itself tools, secrets, or network access
- bypass Oktopus events, artifacts, policies, approvals, or leases

### Capability Shape

Register Archon twice if useful:

```yaml
kind: adapter
name: archon
version: 0.1.0
description: Adapter for invoking Archon structured AI workflows under Oktopus control.
entrypoint:
  type: cli
  command: archon
permissions:
  filesystem: scoped_workspace
  network: policy_controlled
  secrets: explicit_allowlist
```

```yaml
kind: tool
name: archon
version: 0.1.0
description: Run explicit Archon workflow commands and produce Oktopus artifacts.
entrypoint:
  type: adapter
  adapter: archon
outputs:
  artifacts:
    - archon-output.json
    - proposal.yaml
    - report.md
```

### Supported Modes

#### 1. Proposal Generation

Archon can generate Oktopus capability proposals.

Examples:

```text
workflow proposal
persona proposal
skill proposal
verifier proposal
solution pack proposal
```

Output location:

```text
runs/<run-id>/jobs/<job-key>/attempts/<n>/artifacts/proposals/<kind>/<name>.yaml
```

Proposal artifacts are not active capabilities.

#### 2. Explicit Workflow Execution

Oktopus can schedule a job that invokes Archon for a known workflow-like operation.

Example:

```yaml
- id: repo-audit
  kind: tool
  uses: tool:archon
  input:
    mode: execute
    workflow: repo-audit
```

Oktopus still owns:

- job lease
- worker identity
- policy checks
- events
- logs
- artifacts
- verifier gates
- final acceptance

Archon output is captured as artifacts and summaries.

#### 3. Catalog Import

Archon may expose a catalog of workflows. Oktopus can import catalog entries as draft proposals.

```text
archon catalog list
→ Oktopus proposal artifacts
→ schema validation
→ adversarial review
→ approval
→ active registry entries
```

No catalog entry is active by default.

### No-Auto-Install Rule

Auto-install means any generated capability becomes live without explicit activation.

Oktopus SHALL NOT allow Archon or an agent to automatically:

```text
write active registry/workflows/*.yaml
write active registry/tools/*.yaml
write active registry/policies/*.yaml
modify trigger rules
grant permissions/secrets
mark generated capability active
make future runs use generated workflow
```

Allowed by default:

```text
write proposal artifact
validate proposal schema
produce diff/report
request approval
```

Activation flow:

```text
proposal artifact
  → schema validation
  → dependency resolution
  → permission diff
  → sandbox dry-run where applicable
  → adversarial review for risky capabilities
  → human/CI approval
  → registry activation
```

### Policy Checks

Before running Archon, Oktopus checks:

```text
mode: proposal|execute|import
requested filesystem scope
network access
secret access
source mutation
registry write intent
trigger rule mutation
capability activation intent
budget limits
```

Default policy:

```text
proposal generation: allowed with scoped workspace write to artifacts only
explicit execution: allowed only if workflow/job grants it
catalog import: proposal-only
active registry mutation: denied unless activation command and approval gate
```

### Events

Minimum events:

```text
archon.invocation_requested
archon.invocation_started
archon.output_received
archon.proposal_created
archon.catalog_imported
archon.invocation_succeeded
archon.invocation_failed
capability.proposal_created
capability.validation_started
capability.validation_failed
capability.validation_succeeded
capability.activation_requested
capability.activation_approved
capability.activated
capability.activation_rejected
```

### Artifacts

Archon job artifacts:

```text
archon-command.json          # command, args, environment policy, redactions
archon-output.json           # structured output when available
archon-report.md             # human-readable report
proposals/<kind>/<name>.yaml # generated capability proposals
logs/stdout.log
logs/stderr.log
```

Proposal metadata should include:

```text
producer: archon
run_id
job_id
attempt_id
source_prompt_or_task
requested_by
created_at
risk_summary
permission_diff
```

### Review Integration

Archon-generated proposals can require adversarial review.

Review triggers:

- requests secrets
- requests network egress
- can mutate source
- can deploy/remediate production
- creates or modifies workflow triggers
- grants new tools
- uses external APIs
- client data access

Review jobs inspect:

```text
proposal artifact
permission diff
workflow DAG
tool/runtime references
policy changes
expected artifacts/verifiers
```

### Workflow Selection Boundary

Archon may recommend a workflow, but Oktopus selection authority remains policy-controlled.

Allowed:

```text
Archon recommends: use ci-remediation workflow
Oktopus records recommendation artifact
Oktopus policy/user chooses run workflow
```

Not allowed by default:

```text
Archon chooses and starts workflow without explicit Oktopus/user/policy selection
```

### Implementation Notes

MVP can stub Archon adapter as a shell command wrapper that records command and output artifacts.

Later adapter can add:

- structured JSON output parsing
- proposal extraction
- catalog import
- validation pipeline
- activation workflow

### Design Constraints

- Archon is a capability, not the control plane.
- Oktopus remains source of truth for active registry and run state.
- Proposal is safe; activation is governed.
- All Archon output becomes artifacts before it can influence active platform state.
