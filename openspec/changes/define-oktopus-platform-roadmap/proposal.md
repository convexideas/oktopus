## Why

Convex Ideas needs a reusable harness platform, not a one-off coding agent wrapper. The target system should coordinate agent runtimes, tools, workflows, sandboxes, policies, evidence, and observability across local development, CI/CD, telemetry-driven operations, documentation, and client-specific inference offerings.

Existing systems solve adjacent problems:

- Harness OSS provides useful patterns for durable executions, scheduler/worker polling, logs, artifacts, and CI-style orchestration.
- OpenClaw is oriented around personal assistant/channel gateway UX.
- Self-improving agent-runtime systems are oriented around agent runtime, memory, and self-improvement.
- Pi, Claude Code, Codex CLI, Antigravity CLI, OpenCode, Kiro, Aider, Goose, OpenShell, Graphify, OpenSpec, Archon, Beads, and Agent Skills are capabilities or runtimes Oktopus should orchestrate, not replace.

Oktopus should become the control plane above these capabilities: workflow DAGs, run/job state, leases, policy gates, artifacts, events, verifiers, and solution packs.

## What Changes

- Define Oktopus as a standalone agentic orchestration control plane.
- Treat Pi and other agent CLIs as worker runtime adapters, not the platform core.
- Define a local-first architecture that can evolve into a distributed controller/worker system.
- Define first-class capabilities: tools, skills, personas, workflows, verifiers, environments, policies, and solution packs.
- Define roadmap phases for core platform, engineering ops, CI/CD, telemetry/AIOps, documentation, and client inference harnesses.
- Establish OpenSpec as the durable planning and change-management mechanism before implementation.

## Out of Scope

- Building a personal assistant/channel gateway.
- Building a new LLM runtime from scratch.
- Building all CI/CD, telemetry, or VLM integrations in the first milestone.
- Making agent-authored capabilities live without review, validation, and activation gates.

## Impact

This change sets the product and architecture direction before deeper implementation. Future changes should refine individual platform slices such as registry schema, worker lease protocol, event model, artifact store, verifier gates, and CI/CD integration.
