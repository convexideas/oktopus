## Why

Oktopus needs a durable execution spine before adapters, UI, CI/CD hooks, telemetry, or inference solution packs can be useful. The platform must record what was requested, which workflow/capabilities were resolved, what jobs were scheduled, which worker leased each job, what happened during execution, which artifacts were produced, and which verifier or approval gates accepted the result.

Without a canonical run/job/event model, agent work remains chat-shaped and hard to resume, audit, debug, distribute, or commercialize as client-specific harnesses.

## What Changes

- Define the first core domain model for Oktopus execution state.
- Introduce durable entities: `Project`, `Run`, `Job`, `Step`, `Attempt`, `Artifact`, `Event`, `Approval`, `Worker`, `Lease`, `Environment`, and `PolicyDecision`.
- Define lifecycle states for runs, jobs, attempts, approvals, and artifacts.
- Define append-only event semantics and minimum event types.
- Define the worker lease protocol at the data-model level.
- Define local-first persistence requirements using SQLite and filesystem artifacts, while preserving a path to Postgres/object-store/event-backend adapters.

## Out of Scope

- Implementing the worker runtime.
- Implementing Pi, shell, Graphify, OpenSpec, Beads, Archon, CI/CD, telemetry, or VLM adapters.
- Building the web UI.
- Choosing final wire protocol details for distributed workers.
- Designing full policy language beyond the initial decision/audit records.

## Impact

This change establishes the platform spine. Future changes can implement the schema, CLI, worker loop, artifact store, verifiers, and adapters against this model.
