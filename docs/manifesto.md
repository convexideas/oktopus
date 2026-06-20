# The Oktopus Manifesto

> The best harness is the one you own.

## The problem

When an organization adopts agents, the work fragments immediately. Everyone reaches for a different tool, and each one is a silo with no shared control:

- **No standardization.** Every person runs agents their own way. There is no shared, repeatable definition of how a task should be done.
- **No shared knowledge.** What one agent produces or learns stays trapped in one person's tool. The organization can't accumulate or reuse it.
- **No enforceable policy.** Permissions, approvals, and guardrails live in prompts — advisory at best — not in anything that can actually stop an unsafe action.
- **No compliance visibility.** There is no record of what ran, on whose data, under which controls. You cannot audit or monitor what you cannot see.

And no vendor agent will reach your internal systems: your data lake, your bespoke tools, and your private APIs are off-limits to tools built to integrate with someone else's world.

The agent is the interchangeable part. The durable, organization-critical layer — standardized workflows, shared knowledge, enforced policy, and an auditable record — has nowhere to live. That layer should belong to you. ([Why a control plane](why.md) breaks down each problem: how it's handled today, and what changes.)

## The inversion

Oktopus is a control plane that owns the durable layer and treats agent runtimes as interchangeable workers. One root conviction drives everything:

> **The control plane owns capabilities and memory; agents plug in as workers.**

The agent is interchangeable; everything that matters stays with you.

## What it provides

Concrete capabilities, each answering a problem above:

- **Policy enforcement** — guardrails on tools, secrets, actions, and process steps, evaluated in code rather than prompts, and configurable hard / soft / waivable.
- **Governance** — capabilities are scoped, versioned, and approval-gated; agent-created ones stay proposals until reviewed. You control who can add, override, or run what.
- **Standardization & reuse** — workflows and capabilities are shared registry entries, repeatable across the organization.
- **Auditability** — every state transition is an append-only event; outputs are captured as evidence.
- **Extensibility** — bring your own skills, tools, commands, secrets, and data sources (internal or third-party) as first-class capabilities, including systems no vendor agent reaches.
- **Distribution** — share and consume capabilities and workflows across teams, and from a governed marketplace of open and third-party providers.
- **Portability** — runs, memory, and artifacts live in the control plane, vendor-agnostic.

In an organization, all of it is scope-governed — per org, client, project, and run.

## What makes this real

A manifesto that isn't enforced is decoration. Each principle below names the specification that makes it true. The principles change rarely; the specs they point to are the normative, testable contracts.

### 1. The control plane is the source of truth; workers are disposable

The controller holds authoritative state. Workers register, lease bounded work, and can crash or be replaced without losing or duplicating work. This is what makes runtimes interchangeable and the system recoverable.
*Enforced by:* `define-worker-lease-protocol`, `define-core-run-job-event-model`

### 2. Evidence over confidence — and output has modalities

A job is done when its work *verifies* in its own modality — an artifact stored, a message delivered, or an effect confirmed on a real system — never on an agent's say-so. Not all useful output is a stored file; a chat turn and a codebase change are complete without one.
*Enforced by:* `define-artifacts-verifiers-policy` (output kinds: artifact / stream / effect)

### 3. Policy is code-enforced, not prompt-enforced

Permissions, secrets, network, filesystem, deployment, and capability activation are gated by the controller and workers — not by instructions in a prompt that a model may ignore. Prompts guide; policy decides.
*Enforced by:* `define-artifacts-verifiers-policy`, `define-archon-adapter`

### 4. Agent-created capabilities are proposals, not live assets

When an agent produces a new tool, workflow, or persona, it becomes a proposal artifact — never a live capability — until schema validation, sandbox dry-run, review, and approval pass. No agent installs itself into production.
*Enforced by:* `define-capability-registry-model`, `define-archon-adapter`

### 5. Capabilities are versioned, pinned, and provenance-tracked

Runs resolve exact capability versions and reproduce against what they resolved. Images and snapshots are content-addressed; nothing backs a workspace without recorded provenance. Audit is built in, not bolted on.
*Enforced by:* `define-capability-registry-model`, `define-image-and-snapshot-store`

### 6. The harness is user-owned, extensible, and scope-governed

Skills, tools, proprietary and internal data sources, commands, and secrets are first-class capabilities you add and compose — including systems no vendor agent supports — drawn from a governed supply chain of builtins, open packs, and third-party marketplace providers. Every addition and override is controlled by scope and policy.
*Enforced by:* `define-client-server-runtime-config-sources`, `define-enterprise-admin-marketplace-sessions`, `define-capability-registry-model`

### 7. Local-first, distributed-ready — adapters change, concepts don't

It runs on one machine with SQLite and the filesystem, and scales to Postgres, object stores, queues, and worker pools by swapping adapters. The run/job/capability concepts are identical in both modes.
*Enforced by:* `define-oktopus-platform-roadmap`, the upgrade-path section of every change

### 8. Adversarial review is first-class, with real isolation

Review is a workflow gate, not a casual prompt. Reviewer jobs run with controller-enforced context isolation — they see only the artifacts explicitly handed to them — and their findings are structured evidence that can block downstream work.
*Enforced by:* `define-core-run-job-event-model`

### 9. Everything is evented and auditable

Every state transition emits an append-only event. State tables answer "what is true now"; the event log is the audit and debug truth. You can always reconstruct what happened, who did it, and which capability versions were involved.
*Enforced by:* `define-core-run-job-event-model`

## Two axes of work

Because agents are interchangeable workers, "what kind of work an agent does" resolves to two orthogonal axes rather than a fixed product category:

- **Engagement lifecycle** — one-shot, interactive, long-running (start/pause/resume), scheduled, event-triggered.
- **Output modality** — artifact, stream, or effect.

A batch report is *one-shot × artifact*. But the axes are independent: an interactive session can stream a conversation *or* edit a codebase (*interactive × effect*), and an effect can be one-shot ("apply this fix") or long-running. Any lifecycle pairs with any modality, and the same spine serves them all.

## The stance

Own your context. Own and extend your harness. Plug in your proprietary world. Consume a governed marketplace. The agent is replaceable; everything that matters is yours.
