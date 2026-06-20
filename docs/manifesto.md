# The Oktopus Manifesto

> The best harness is the one you own.

## The problem

When an organization adopts agents, the work fragments immediately. Everyone reaches for a different tool, and each one is a silo with no shared control:

- **No standardization.** Every person runs agents their own way. There is no shared, repeatable definition of how a task should be done.
- **No shared knowledge.** What one agent produces or learns stays trapped in one person's tool. The organization can't accumulate or reuse it.
- **No enforceable policy.** Permissions, approvals, and guardrails live in prompts — advisory at best — not in anything that can actually stop an unsafe action.
- **No compliance visibility.** There is no record of what ran, on whose data, under which controls. You cannot audit or monitor what you cannot see.

And reaching your internal systems — your data lake, your bespoke tools, your private APIs — means bolting a custom hook onto each vendor's agent: possible, but redone per tool and ungoverned.

The agent is the interchangeable part. The durable, organization-critical layer — standardized workflows, shared knowledge, enforced policy, and an auditable record — has nowhere to live. That layer should belong to you. ([Why a control plane](why.md) breaks down each problem: how it's handled today, and what changes.)

## The inversion

Oktopus is a control plane that owns the durable layer and treats agent runtimes as interchangeable workers. One root conviction drives everything:

> **The control plane owns capabilities and memory; agents plug in as workers.**

## What it provides

Concrete capabilities — first to compose work, then to govern it, then to account for it:

- **Standardization & reuse** — workflows and capabilities are shared registry entries, repeatable across the organization.
- **Extensibility** — bring your own skills, tools, commands, secrets, and data sources (internal or third-party) as first-class, governed capabilities — connected once and reused by every agent, not re-wired per tool.
- **Distribution** — share and consume capabilities and workflows across teams, and from a governed marketplace of open and third-party providers.
- **Policy enforcement** — access and process gates (tool use, secret reads, network/filesystem scope, task ordering, prerequisites, destructive actions, deployment, budgets) are decided in code at execution time, not left to the prompt, and are configurable hard / soft / waivable. Conduct *within* a step isn't code-enforced — it's gated by verifiers on the output, not trusted.
- **Governance** — capabilities are scoped, versioned, and approval-gated; agent-created ones stay proposals until reviewed. You control who can add, override, or run what.
- **Credentials & budgets** — distribute and scope API keys and licenses centrally instead of scattering them across everyone's environment; access is policy-gated, attributable in the event log, and spend is governed against budgets.
- **Auditability** — every state transition is an append-only event; outputs are captured as evidence.
- **Portability** — runs, artifacts, and scoped memory (user, team, org) live in the control plane. Context transfers across agents: switch the runtime, keep the context.

In an organization, all of it is scope-governed — per org, client, project, and run.

## What makes this real

A manifesto that isn't enforced is decoration. Each principle links to the spec that enforces it — the normative, testable contract. Principles change rarely; specs evolve under them.

### 1. The control plane is the source of truth; workers are disposable

The controller holds authoritative state; workers lease bounded work and can crash or be replaced without losing or duplicating it. That is what makes runtimes interchangeable.
*Enforced by:* [define-worker-lease-protocol](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-worker-lease-protocol), [define-core-run-job-event-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-core-run-job-event-model)

### 2. Everything is evented and auditable

Every state transition emits an append-only event. State tables say what is true now; the event log is the audit truth — you can reconstruct what happened, who did it, and which versions were involved.
*Enforced by:* [define-core-run-job-event-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-core-run-job-event-model)

### 3. Evidence over confidence — and output has modalities

A job is done when its work *verifies* in its own modality — artifact stored, message delivered, or effect confirmed — never on an agent's say-so. Not all output is a file.
*Enforced by:* [define-artifacts-verifiers-policy](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-artifacts-verifiers-policy) (output kinds: artifact / stream / effect)

### 4. Adversarial review is first-class, with real isolation

Review is a workflow gate, not a prompt: reviewer jobs see only the artifacts handed to them, and their findings are structured evidence that can block downstream work.
*Enforced by:* [define-core-run-job-event-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-core-run-job-event-model)

### 5. Policy is code-enforced, not prompt-enforced

Permissions, secrets, network, filesystem, deployment, and capability activation are gated by the controller, not by a prompt a model may ignore. Prompts guide; policy decides.
*Enforced by:* [define-artifacts-verifiers-policy](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-artifacts-verifiers-policy), [define-archon-adapter](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-archon-adapter)

### 6. Agent-created capabilities are proposals, not live assets

An agent-produced tool, workflow, or persona is a proposal until schema validation, sandbox dry-run, review, and approval pass. No agent installs itself into production.
*Enforced by:* [define-capability-registry-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-capability-registry-model), [define-archon-adapter](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-archon-adapter)

### 7. Capabilities are versioned, pinned, and provenance-tracked

Runs pin exact capability versions and reproduce against them; images and snapshots are content-addressed, and nothing backs a workspace without provenance.
*Enforced by:* [define-capability-registry-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-capability-registry-model), [define-image-and-snapshot-store](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-image-and-snapshot-store)

### 8. The harness is user-owned, extensible, and scope-governed

Skills, tools, data sources, commands, and secrets are first-class capabilities you add and compose — connected once, governed by scope and policy — from a supply chain of builtins, open packs, and third-party providers.
*Enforced by:* [define-client-server-runtime-config-sources](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-client-server-runtime-config-sources), [define-enterprise-admin-marketplace-sessions](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-enterprise-admin-marketplace-sessions), [define-capability-registry-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-capability-registry-model)

### 9. Local-first, distributed-ready — adapters change, concepts don't

It runs on one machine (SQLite + filesystem) and scales to Postgres, object stores, queues, and worker pools by swapping adapters; the run/job/capability concepts are identical in both.
*Enforced by:* [define-oktopus-platform-roadmap](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-oktopus-platform-roadmap), and the upgrade-path section of every change

## Two axes of work

Work spans two independent axes, not a fixed product category:

- **Engagement lifecycle** — one-shot, interactive, long-running (start/pause/resume), scheduled, event-triggered.
- **Output modality** — artifact, stream, or effect.

A batch report is *one-shot × artifact*. But the axes are independent: an interactive session can stream a conversation *or* edit a codebase (*interactive × effect*), and an effect can be one-shot ("apply this fix") or long-running. Any lifecycle pairs with any modality, and the same spine serves them all.

## The stance

Own your context. Own and extend your harness. Plug in your proprietary world. Consume a governed marketplace. The agent is replaceable; everything that matters is yours.
