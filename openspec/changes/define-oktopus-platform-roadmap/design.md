## Design

### Product Positioning

Oktopus is a governed agentic operations platform.

```text
Oktopus = controller + registry + scheduler + policy + events + artifacts + workers
```

It coordinates agent and tool work across engineering, operations, documentation, and inference-product workflows.

Oktopus owns:

- Capability registry and versioning.
- Workflow DAGs and lifecycle.
- Run, job, step, attempt, artifact, event, approval, and policy state.
- Worker scheduling, leases, cancellation, and retries.
- Verification gates and evidence capture.
- Observability, audit, and remediation records.

Oktopus does not own:

- Chat/channel gateway UX as core product.
- A single built-in agent brain.
- A single model provider strategy.
- Client-specific inference logic in the core platform.

### System Shape

```text
Inputs
  Git/PR | CI/CD | alerts | docs | tickets | chat | API
    │
    ▼
Oktopus Controller API
    │
    ├── Registry: tools, skills, personas, workflows, verifiers, environments, policies, solution packs
    ├── Run DB: runs, jobs, steps, attempts, approvals, leases
    ├── Event Stream: typed append-only events and traces
    ├── Artifact Store: logs, reports, diffs, specs, graphs, screenshots, eval outputs
    └── Policy Engine: permissions, secrets, approvals, budgets, blast-radius rules
    │
    ▼
Scheduler
    │
    ▼
Workers poll and lease jobs by kind/type/labels
    │
    ├── agent worker: Pi, Claude Code, Codex CLI, Antigravity CLI, OpenCode, Kiro, Aider, Goose
    ├── tool worker: Graphify, OpenSpec, Archon, Beads, shell, MCP tools
    ├── sandbox worker: local shell, Docker, OpenShell, microVM, Kubernetes
    ├── inference worker: VLM/RAG/model eval services
    └── review worker: independent persona/verifier passes
```

### Core Entity Model

```text
Project       = tenant/repo/product boundary
Capability    = versioned tool, skill, persona, workflow, verifier, environment, policy, solution pack
Run           = user intent or triggered workflow execution
Job           = schedulable unit in a workflow DAG
Step          = model call, tool call, verifier, approval, or artifact operation inside a job
Attempt       = retry/repair execution of a job or step
Artifact      = durable output with type, path, digest, provenance, and retention
Event         = append-only state transition or observation
Approval      = human or policy decision gate
Worker        = process/service that leases and executes jobs
Environment   = execution boundary and resource policy
```

### Local-First to Distributed Path

Phase 0 uses local files and SQLite. Later phases replace adapters, not concepts.

| Local MVP | Distributed Future |
|---|---|
| SQLite | Postgres |
| local `runs/` | object store + event DB |
| local worker subprocess | worker pool |
| direct command execution | Docker/OpenShell/Kubernetes workers |
| local manifests | signed org registry |
| CLI logs | web UI + OpenTelemetry |

### Worker Lease Pattern

Workers request work by capabilities and constraints.

```text
worker → Request(kind, type, labels, platform)
controller → candidate job
worker → Accept(job_id, worker_id, lease_ttl)
controller → compare-and-swap lease
worker → stream events/logs/artifacts
worker → Complete/Fail/Cancel
```

This avoids duplicate execution when multiple workers see the same queued job.

### Capability Lifecycle

Capabilities are manifest-defined and versioned.

```text
draft → schema validate → sandbox dry-run → review → activate → deprecate/archive
```

Agent-created capabilities are proposals only until approved.

### Solution Packs

Solution packs compose platform primitives into domain offerings.

Examples:

- `engineering-ops-pack`: spec, plan, build, test, review, ship.
- `ci-remediation-pack`: diagnose failed CI, propose patch, verify, open PR.
- `aiops-pack`: alert triage, RCA, runbook selection, safe remediation, postmortem.
- `docs-pack`: architecture graph, ADRs, release notes, runbooks, doc freshness.
- `vlm-harness-pack`: dataset intake, golden set, evals, human review, drift monitoring, deployment gates.

### Roadmap Phases

1. **Spec foundation**: OpenSpec roadmap, requirement deltas, architecture decisions.
2. **Core spine**: registry, SQLite run DB, run/job/event model, CLI.
3. **Worker protocol**: polling, leases, heartbeats, cancellation, local worker.
4. **Artifacts and verifiers**: evidence store, command/artifact checks, final receipts.
5. **Agent/tool adapters**: Pi, shell, Graphify, OpenSpec, Beads, Archon.
6. **Policy gates**: permissions, secrets, approvals, budgets.
7. **Distributed mode**: Postgres, queue/event backend, worker pools, object store.
8. **CI/CD integration**: GitHub Actions, Harness/Drone/GitLab/Jenkins/Tekton adapters.
9. **Telemetry/AIOps**: OpenTelemetry, Prometheus, Loki/ELK, Grafana, PagerDuty/Opsgenie, Sentry.
10. **Documentation workflows**: ADRs, architecture maps, runbooks, postmortems.
11. **Client inference offerings**: VLM/RAG/eval-specific solution packs and dashboards.

### Technology Direction

- Core/controller/scheduler/worker/CLI: Go.
- Web UI: TypeScript/React.
- Inference/eval adapters: Python where useful.
- Manifests: YAML/TOML plus JSON Schema.
- Events/tracing: append-only event model, OpenTelemetry later.

### Design Constraints

- Chat is an input channel, not the primary system shape.
- Run page and artifacts are the primary UX.
- Policy must be code-enforced, not prompt-only.
- Verification requires external evidence, not agent confidence.
- Distributed architecture must preserve local-first ergonomics.
