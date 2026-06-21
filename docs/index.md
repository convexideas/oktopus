# The Oktopus Manifesto

> The best harness is the one you own.

Oktopus is a vendor-agnostic control plane for agent orchestration and governance. The agent is replaceable; your context, your harness, your proprietary integrations, and your governance are yours.

## The problem

When an organization adopts agents, the work fragments immediately. Everyone reaches for a different tool, and each one is a silo with no shared control:

- **No standardization.** Every person runs agents their own way. There is no shared, repeatable definition of how a task should be done.
- **No shared knowledge.** What one agent produces or learns stays trapped in one person's tool. The organization can't accumulate or reuse it.
- **No enforceable policy.** Permissions, approvals, and guardrails live in prompts — advisory at best — not in anything that can actually stop an unsafe action.
- **No compliance visibility.** There is no record of what ran, on whose data, under which controls. You cannot audit or monitor what you cannot see.

And reaching your internal systems — your data lake, your bespoke tools, your private APIs — means bolting a custom hook onto each vendor's agent: possible, but redone per tool and ungoverned. `The agent` is the interchangeable part. The durable, organization-critical layer — standardized workflows, shared knowledge, enforced policy, and an auditable record — has nowhere to live. That layer should belong to you.

## The approach

Oktopus owns the durable layer and treats agent runtimes as interchangeable workers. One root conviction:

> **The control plane owns capabilities and memory; agents plug in as workers.**

From there, three pillars — compose your work, govern it, account for it:

### Compose

- **Standardization & reuse** — workflows and capabilities are shared registry entries, repeatable across the organization. *Enforced by [define-capability-registry-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-capability-registry-model).*
- **Extensibility** — bring your own skills, tools, commands, secrets, and data sources (internal or third-party) as first-class, governed capabilities — connected once and reused by every agent, not re-wired per tool. *Enforced by [define-client-server-runtime-config-sources](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-client-server-runtime-config-sources), [define-capability-registry-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-capability-registry-model).*
- **Distribution** — share and consume capabilities and workflows across teams, and from a governed marketplace of open and third-party providers. *Enforced by [define-enterprise-admin-marketplace-sessions](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-enterprise-admin-marketplace-sessions), [define-capability-registry-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-capability-registry-model).*

### Govern

- **Policy enforcement** — access and process gates (tool use, secret reads, network/filesystem scope, task ordering, prerequisites, destructive actions, deployment, budgets) are decided in code at execution time, not left to the prompt, and are configurable hard / soft / waivable. Conduct *within* a step isn't code-enforced — it's gated by verifiers on the output, not trusted. *Enforced by [define-artifacts-verifiers-policy](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-artifacts-verifiers-policy), [define-archon-adapter](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-archon-adapter).*
- **Governance** — capabilities are scoped, versioned, and approval-gated; agent-created ones stay proposals until reviewed. You control who can add, override, or run what. *Enforced by [define-capability-registry-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-capability-registry-model), [define-archon-adapter](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-archon-adapter).*
- **Credentials & budgets** — distribute and scope API keys and licenses centrally instead of scattering them across everyone's environment; access is policy-gated, attributable in the event log, and spend is governed against budgets. *Enforced by [define-client-server-runtime-config-sources](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-client-server-runtime-config-sources), [define-artifacts-verifiers-policy](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-artifacts-verifiers-policy).*

### Account

- **Auditability** — every state transition is an append-only event; outputs are captured as evidence. *Enforced by [define-core-run-job-event-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-core-run-job-event-model), [define-artifacts-verifiers-policy](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-artifacts-verifiers-policy).*
- **Portability** — runs, artifacts, and scoped memory (user, team, org) live in the control plane. Context transfers across agents: switch the runtime, keep the context. *Enforced by [define-core-run-job-event-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-core-run-job-event-model), [define-enterprise-admin-marketplace-sessions](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-enterprise-admin-marketplace-sessions).*

In an organization, all of it is scope-governed — per org, client, project, and run.

## Read next

- [Why a control plane](why.md) — the problems, how they're handled today, what changes
- [Concepts](concepts.md) — the model: entities, workers, evidence, primitives
- [Extending the Harness](extending.md) — capabilities, sources, governance
- [Specifications](specifications.md) — normative requirements (the source of truth)
- [Roadmap](roadmap.md) — build phases

Specs are the source of truth for behavior; this book links to them rather than restating them.
