# The Oktopus Manifesto

> The best harness is the one you own.

## The problem

Your agent conversations, your accumulated context, and your tools are trapped inside whichever vendor you used. ChatGPT holds your threads. Claude holds its own. Cursor holds another. Switch tools and you lose the history, the memory, and any say over how the work was governed. Worse, each vendor integrates with the world *they* choose — public APIs, popular SaaS. None of them will ever connect to your internal data lake, your bespoke tools, or your private systems.

In the agentic era this is backwards. The agent is the interchangeable part. Your conversations, your memory, your capabilities, and your governance are the durable assets — and they should belong to you, not to a model vendor.

## The inversion

Oktopus is a control plane that owns the durable layer and treats agent runtimes as interchangeable workers. One root conviction drives everything:

> **The control plane owns capabilities and memory; agents plug in as workers.**

From that root, four things become yours:

1. **Own your context.** Runs, threads, memory, artifacts, and evidence live in the control plane — agent-agnostic, portable, auditable. Change the agent without losing the work.

2. **Own and extend your harness.** Your skills, tools, commands, and secrets are first-class capabilities you add and compose — not a fixed menu chosen by a vendor.

3. **Plug in your proprietary world.** Internal data sources and tooling that no vendor agent supports connect through governed adapters, custom tools, MCP servers, and knowledge sources.

4. **Consume a governed supply chain.** Capabilities flow from platform builtins, open packs, third-party marketplace providers, and your own internal systems — all through one registry with the same trust, provenance, scope, and policy gates.

And in an organization, every one of these is **scope-governed**: who can add, override, or activate what is controlled per org, client, project, and run.

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

A batch report is *one-shot × artifact*. A chat assistant is *interactive × stream*. A coding agent on a workspace you pause and resume is *long-running × effect*. The same spine serves all of them.

## The stance

Own your context. Own and extend your harness. Plug in your proprietary world. Consume a governed marketplace. The agent is replaceable; everything that matters is yours.
