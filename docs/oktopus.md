# Oktopus

Oktopus is a vendor-agnostic control plane for agent orchestration, governance, and memory — how you [own the harness](index.md) in practice.

The *model* is replaceable; the harness is yours — standardized where the organization needs repeatability, personalized where people need room to explore.

## The problem

When an organization adopts agents, the work fragments immediately. Everyone reaches for a different tool, and each one is a silo with no shared control:

- **No standardization.** Every person runs agents their own way. There is no shared, repeatable definition of how a task should be done.
- **No shared knowledge.** What one agent produces or learns stays trapped in one person's tool. The organization can't accumulate or reuse it.
- **No enforceable policy.** Permissions, approvals, and guardrails live in prompts — advisory at best — not in anything that can actually stop an unsafe action.
- **No process discipline.** The hardest rules are behavioral — every task tracing to a detailed user story, steps running in the right order — and today they rely on human discipline that slips under deadline.
- **No compliance visibility.** There is no record of what ran, on whose data, under which controls. You cannot audit or monitor what you cannot see.

And reaching your internal systems — your data lake, your bespoke tools, your private APIs, your docs, your incident history — means bolting a custom hook onto each vendor's agent: possible, but redone per tool and ungoverned. The model is the interchangeable part. The durable, organization-critical layer — standardized workflows, shared knowledge, scoped memory, enforced policy, and an auditable record — has nowhere to live. That layer should belong to you.

## The approach

Oktopus owns the durable layer and treats agent runtimes as interchangeable workers. One root conviction:

> **The control plane owns capabilities and memory; agents plug in as workers.**

From there, five pillars — compose the work, connect your systems, govern the execution, monitor what happens, and own what persists:

### Compose

- **Standardization & reuse** — workflows and capabilities are shared registry entries, repeatable across the organization. *Enforced by [define-capability-registry-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-capability-registry-model).*
- **Extensibility** — bring your own skills, tools, commands, MCP servers, secrets, data sources, and output destinations (internal or third-party) as first-class, governed capabilities — connected once and reused by every agent, not re-wired per tool. *Enforced by [define-client-server-runtime-config-sources](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-client-server-runtime-config-sources), [define-capability-registry-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-capability-registry-model).*
- **Distribution** — share and consume capabilities and workflows across teams, and eventually from a governed marketplace of internal, open, and third-party providers. *Enforced by [define-enterprise-admin-marketplace-sessions](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-enterprise-admin-marketplace-sessions), [define-capability-registry-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-capability-registry-model).*

### Connect

- **Inputs** — work arrives from chat, CLI, CI, alerts, tickets, schedules, or APIs. *Enforced by [define-client-server-runtime-config-sources](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-client-server-runtime-config-sources).*
- **Outputs** — results land in pull requests, docs, incident rooms, dashboards, object stores, client portals, or wherever the workflow designates. *Enforced by [define-client-server-runtime-config-sources](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-client-server-runtime-config-sources), [define-enterprise-admin-marketplace-sessions](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-enterprise-admin-marketplace-sessions).*

### Govern

- **Policy enforcement** — access and process gates (tool use, secret reads, network/filesystem scope, task ordering, prerequisites, destructive actions, deployment, budgets) are decided in code at execution time, not left to the prompt, and are configurable as hard, soft, or waivable. The harness may not control every token inside a model call, but it controls what the model can access, what actions execute, and what evidence is required before work is accepted. *Enforced by [define-artifacts-verifiers-policy](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-artifacts-verifiers-policy), [define-archon-adapter](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-archon-adapter).*
- **Governance** — capabilities are scoped, versioned, and approval-gated; agent-created ones stay proposals until reviewed. You control who can add, override, or run what — scoped per org, client, project, thread, and run. Standard workflows stay governed; exploratory agents stay customizable within policy. *Enforced by [define-capability-registry-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-capability-registry-model), [define-archon-adapter](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-archon-adapter).*
- **Credentials & budgets** — distribute and scope API keys, model access, and licenses centrally instead of scattering them across everyone's environment; access is policy-gated, attributable in the event log, and spend is governed against budgets. *Enforced by [define-client-server-runtime-config-sources](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-client-server-runtime-config-sources), [define-artifacts-verifiers-policy](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-artifacts-verifiers-policy).*

### Monitor

- **Auditability** — every state transition is an append-only event and outputs are captured as evidence, so usage, cost, and compliance stay visible on the record. *Enforced by [define-core-run-job-event-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-core-run-job-event-model), [define-artifacts-verifiers-policy](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-artifacts-verifiers-policy).*

### Own

- **Portability** — runs, artifacts, and scoped memory (thread, user, team, org) live in the control plane. Context transfers across models: switch the runtime, keep the context. *Enforced by [define-core-run-job-event-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-core-run-job-event-model), [define-enterprise-admin-marketplace-sessions](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-enterprise-admin-marketplace-sessions).*
- **Resumability** — work does not disappear when a chat ends. Threads preserve memory, artifacts, approvals, policies, workspace references, and output bindings so a task can pause, resume, or move to another runtime. *Enforced by [define-enterprise-admin-marketplace-sessions](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-enterprise-admin-marketplace-sessions).*

## Read next

- [Concepts](concepts.md) — the model: entities, workers, evidence, primitives
- [Extending the Harness](extending.md) — capabilities, sources, governance
- [Specifications](specifications.md) — normative requirements (the source of truth)
- [Roadmap](roadmap.md) — build phases

Specs are the source of truth for behavior; this book links to them rather than restating them.
