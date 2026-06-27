# Application Design Notes

This is the living design document for Oktopus while the implementation evolves. OpenSpec remains the normative source of truth; this page records the working application shape, navigation strategy, and maintenance rules so implementation decisions do not drift.

## Product frame

Oktopus is a framework for building enterprise agent systems — an operating layer and control plane for composing, governing, monitoring, and resuming agentic work across runtimes.

The application must stay useful in three modes:

1. **Local MVP** — coding sessions only: profile, immutable workspace definition, user-bound sandbox instance, shareable session, local store, local artifacts.
2. **Team server** — API server, background workers, shared DB, object store, configured sources.
3. **Enterprise platform** — marketplace, governance, identity, policies, workspaces, distributed workers.

## Current application spine

See [Actors and Interactions](actors-and-interactions.md) for the detailed actor map: config items, registry, resolver, installer, sandbox, runtime adapter, worker, stores, and verifier flow.

```text
CLI / API
  → profile resolution
  → immutable workspace resolution
  → user-bound sandbox instance
  → session creation
  → workspace materialization
  → OpenShell sandbox provider
  → agent process attach/logs
  → session events + artifacts
```

The control plane owns durable truth. For MVP, agent execution is attached to coding sessions; general workflows/runs/jobs are deferred.

## Current implementation map

```text
cmd/oktopus/              CLI entrypoint
internal/cli/             Cobra commands
internal/store/           storage boundary (SQLite now, Postgres later)
internal/db/              SQLite migration implementation
internal/registry/        YAML capability registry loader + relational index sync
internal/runtimeadapter/  runtime adapter contract
internal/skillindex/      exact-path Agent Skills indexer
registry/                 seed capabilities
openspec/                 normative requirements and tasks
```

## Design invariants

1. **Registry is source of truth for capabilities.** Runtime code must not hard-code the available tools, skills, personas, or workflows.
2. **Profiles choose defaults and limits.** Profiles may be user/org scoped and define defaults, provider choices, and policy limits.
3. **Workspaces are immutable shared definitions.** Workspaces are org/platform scoped templates for source, environment, and global resource refs. Change by forking, not mutating.
4. **Sandboxes are user-bound private instances.** Sandboxes are created from workspace + profile + session context and may contain user identity/secrets; sharing a session does not share the sandbox.
5. **Sessions are task capsules.** Sessions bind a user/task to a workspace and may be shared by policy; another user resumes into their own sandbox.
6. **Workspaces are materialized into sandboxes.** Materialization prepares context, generated instructions, memory summaries, runtime launch metadata, event paths, and artifact paths.
7. **Secrets do not enter workspace files or staging.** Prepared files contain references and policy intent; sandbox/provider layers inject credentials.
8. **State transitions emit events.** Durable state without an event is design drift.
9. **Policy is code-enforced.** Prompts can guide, but enforcement happens in controller/provider/sandbox.
10. **Local-first, distributed-ready.** SQLite/filesystem implementations must sit behind store/artifact/provider boundaries that can later swap to Postgres/object store/remote workers.

## Graphify and SCIP usage

Use both, but for different layers.

### Graphify

Use Graphify for macro architecture questions:

- component relationships
- file/module community structure
- “what owns what?” questions
- drift between docs and code
- onboarding a new contributor or agent
- periodic architecture snapshots

Suggested cadence:

```text
after each major phase
before larger refactors
before publishing architecture docs
when file relationships become hard to hold in working memory
```

Do not run Graphify on every small edit. It is a map-building tool, not the default code navigation tool.

### SCIP

Use SCIP for symbol-level implementation work:

- definitions
- references
- call sites
- package structure
- safe refactors
- public API checks

Prefer SCIP/LSP over ad-hoc text scanning when navigating code. If the SCIP index is stale or missing, regenerate it before relying on results.

## Design document maintenance rules

Update this document when:

- a new package or major component is added
- a boundary changes
- a design invariant changes
- a local-only shortcut is intentionally accepted
- an OpenSpec change introduces new implementation shape
- Graphify or SCIP reveals code/docs drift

Do not duplicate OpenSpec requirements here. Link to specs for normative behavior; summarize only the implementation shape and decisions.

## Storage boundary

`internal/store` is the application entrypoint for durable relational state. SQLite is the only implemented engine today, but callers should depend on store-level open/migration behavior rather than constructing SQLite connections directly.

```text
local MVP:   sqlite://.oktopus/oktopus.db
team/server: postgres://...       # adapter not implemented yet
```

Rule: new scheduler, run, lease, policy, and artifact metadata code must not rely on SQLite-only locking or PRAGMA behavior. If it needs storage semantics, put them behind the store/repository layer first.

## OpenShell boundary

OpenShell is the first local sandbox provider. Oktopus should wrap its gateway/sandbox model instead of replacing it:

```text
Oktopus profile/workspace/session
  → materialize workspace for sandbox
  → OpenShell sandbox create/connect/exec/logs
  → agent process inside sandbox
```

Oktopus owns immutable workspace definitions, sessions, memory, artifacts, profile metadata, sandbox references, and future hub sync. OpenShell owns sandbox lifecycle, supervisor enforcement, filesystem/network policy, credential injection, and terminal/connect plumbing.

## Near-term build order

1. Keep registry + store foundation green.
2. Add immutable workspace fields and a sandbox instance table.
3. Add `profile init`, `workspace create/fork`, and `session start/list/show/log`.
4. Add workspace materialization for sandbox startup.
5. Add generated session `AGENTS.md` and artifact directory.
6. Add OpenShell provider wrapper for sandbox create/connect/logs.
7. Add `session attach`.

Deferred: runs, jobs, workflows, leases, workers, verifiers, marketplace, remote sync.

## Open questions

- What is the minimum useful OpenShell workspace layout?
- Should workspace source be copied, mounted, uploaded, or symlinked for the first local sandbox?
- When should Graphify output become a first-class session artifact?
