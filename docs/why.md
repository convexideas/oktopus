# Why a control plane

Agents are easy to adopt and hard to govern. Here are the problems organizations hit — how each is handled today, and what a control plane changes — moving from composing work, to governing it, to accounting for it. The hardest and most overlooked is process & traceability.

## Standardization & shared knowledge

**Problem.** Every person runs agents their own way, and what one produces stays trapped in their tool. Nothing accumulates org-wide.

**Today.** Prompt copy-paste, tribal knowledge, ad hoc setups.

**With a control plane.** Workflows and capabilities are versioned registry entries, shared and reused across the org. A task done well once becomes a capability everyone can run.

## Proprietary & internal integration

**Problem.** Agents need to reach internal systems — data lakes, bespoke tools, private APIs.

**Today.** Each vendor offers its own hook — an MCP server, a plugin — so it's possible, but bolted on per tool, ungoverned, and re-wired for every agent.

**With a control plane.** Internal systems plug in as governed capabilities — adapters, custom tools, MCP servers, knowledge sources — connected once and reused across agents, versioned and scoped like any other.

## Process & traceability gates

**Problem.** Work should trace to a sufficiently-detailed user story, and steps should run in the right order. These are behavioral rules, the hardest kind to enforce.

**Today.** Enforced by human discipline and review — skipped under deadline pressure, and slow when enforced by hand. The choice between discipline and speed is made invisibly, per person.

**With a control plane.** Ordering is structural: a job isn't eligible until its prerequisites complete. Traceability is a precondition: a task requires a linked story to start. "Detailed enough" is a verifier, not a reviewer's mood. And gates are configurable — hard, soft, or waivable — so bypasses under pressure are recorded exceptions, not silent ones.

## Policy & permission enforcement

**Problem.** Guardrails on tools, secrets, network, and destructive actions need to actually stop unsafe operations.

**Today.** Written in prompts and wikis — advisory, and ignored when inconvenient.

**With a control plane.** Permissions and approvals are evaluated in code at execution time, outside the prompt. An action outside policy is denied or escalated before it runs.

## Credentials & usage

**Problem.** Teams need API keys and licenses to run agents, and the organization needs to control that access and account for the spend.

**Today.** Keys live in scattered `.env` files and personal accounts — no central control, no attribution, no view of usage or cost.

**With a control plane.** Keys and licenses are held and distributed centrally, scoped per team and project and policy-gated on use. Access is attributable through the event log, and spend is governed against budgets.

## Compliance & audit

**Problem.** You need to know what ran, on whose data, under which controls — to audit, debug, or prove compliance.

**Today.** No durable record; reconstruction from scattered logs, if at all.

**With a control plane.** Every state transition emits an append-only event, and outputs are captured as evidence. The timeline, the data touched, and the capability versions used are all on the record.

## Vendor lock-in & portability

**Problem.** Context and memory, individual and team, are trapped inside one vendor; switching loses history and control.

**Today.** Each tool owns its own threads; there's no portable layer.

**With a control plane.** Runs, artifacts, evidence, and scoped memory (user, team, org) live in the control plane, so context isn't tied to any one model — switch the model and keep the context.

## The boundary

A control plane governs what flows through it. Work done entirely outside the harness, it can't see — and a verifier is only as good as its check. State the limit plainly: this enforces process for work that runs on the platform, not human behavior in general.
