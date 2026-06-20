## Why

Oktopus needs a small local MVP that proves the platform spine before distributed services, UI, CI/CD integrations, telemetry integrations, or inference solution packs are built. The MVP should exercise the same core concepts intended for distributed mode: capability registry, run/job/event state, worker leases, artifacts, verifiers, policy decisions, and approvals.

A local-first CLI and SQLite implementation gives Convex Ideas a practical development loop while preventing early architecture from becoming chat-shaped or Pi-specific.

## What Changes

- Define the local MVP packaging and CLI shape.
- Define Go as the core implementation language.
- Define SQLite as the local state store.
- Define filesystem paths for registry files, run artifacts, event mirrors, and logs.
- Define initial CLI commands.
- Define initial database tables and migration approach.
- Define initial local worker behavior.
- Define minimum seeded capabilities and workflows.
- Define validation required before moving to adapter-specific work.

## Out of Scope

- Web UI.
- Remote worker API.
- Postgres/object-store/queue backends.
- Pi/Graphify/OpenSpec adapter execution beyond stubs or shell commands.
- Production-grade auth/RBAC.
- Full policy engine.

## Impact

This change converts the OpenSpec roadmap into an implementable local milestone. After this MVP exists, later changes can add adapters, remote workers, CI/CD triggers, telemetry ingestion, and solution packs on top of proven primitives.
