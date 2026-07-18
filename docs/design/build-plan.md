# Build Plan

Updated: 2026-07-16

See also: [Entity Definitions](entity-definitions.md) | [Memory Ontology](memory-ontology.md)

## Design Principles

- **Go, idiomatic** — stdlib-first, interfaces at consumption boundaries
- **DDD** — bounded contexts own their entities, clean boundaries
- **Pi-first** — build deep for one harness, generalize later
- **Portal-native** — entities designed as if portal exists, stored locally for now
- **Memory is the moat** — every session makes the next one smarter

## Architecture

```
internal/
  identity/       Profile (preferences + harness config)
  registry/       Capabilities (personas, skills, tools) + manifest parsing
  execution/      Session, Workspace, Sandbox, Harness, Assembler
    pi/           Pi adapter (ArgsBuilder + ReadConversation)
    codex/        Codex adapter
    kiro/         Kiro adapter
    claude/       Claude adapter
  memory/         Episodes, summaries, future: semantic store
  policy/         Hooks (enforcement at every level)
  importer/       Onboarding from existing harnesses
  config/         App config (koanf)
  store/sqlite/   Persistence (portal API adapter later)
  cli/            Commands (thin delegates)
```

## Completed

| # | What |
|---|------|
| 1 | Persona manifest schema + parser |
| 2 | `ok capabilities add/list` |
| 3 | `ok workspace create/list/show` |
| 4 | `ok run` with --persona, --workspace, --runtime, --task |
| 5 | DDD refactor (identity, registry, execution, memory bounded contexts) |
| 6 | Profile with preferences + harness config |
| 7 | Pi adapter (ArgsBuilder, model, tools, non-interactive) |
| 8 | Memory capture (stdout, all modes) + LLM summaries + pre-session injection |
| 9 | `ok init` (import Pi settings + session history) |
| 10 | Assembler (layout-driven, generic across harnesses) |

## Current Phase: Sandbox + Native Capture

| # | What | Details |
|---|------|---------|
| 11 | **Sandbox redesign** | Stable paths (`~/.ok/<workspace>/sandboxes/<name>/home/`). HOME redirect. No workspace cloning. Persistent. |
| 12 | **`ok sandbox create/exec/list`** | User-facing commands. Create named sandbox, enter it (shell with HOME redirected), list existing. |
| 13 | **Harness reads sandbox HOME** | `ReadConversation(sandboxHome)` on Harness interface. Pi reads its JSONL from sandbox. Base returns stdout fallback. |
| 14 | **Update `ok run`** | Use new sandbox. Wire sandbox env (HOME) into harness. Call ReadConversation post-session. |
| 15 | **Bare `ok run`** | Profile defaults (default_runtime) make flags optional. `ok run` in a workspace just works. |

## Next Phase: Policy + Governance

| # | What | Details |
|---|------|---------|
| 16 | **Policy engine** | Hooks checked pre-session and per-tool-call. Declarative YAML. Stacking: org > workspace > sandbox > session. |
| 17 | **Spend cap builtin** | Track token/cost per session. Warn at threshold. Deny at cap. |
| 18 | **Tool approval builtin** | Pause before shell/file-write. Auto-approve within bounds. |
| 19 | **Policy on sandbox** | Sandbox carries policy. Automations constrained by policy (less privileged). |

## Then: Generalize + Portal Prep

| # | What | Details |
|---|------|---------|
| 20 | **Codex adapter** | AGENTS.md materialization, `codex -q`, ReadConversation. |
| 21 | **Kiro adapter** | `.kiro/steering/*.md`, ReadConversation. |
| 22 | **Workspace as full I/O spec** | Inputs, outputs, connectors, constraints — not just source ref. |
| 23 | **Session as portable record** | Full conversation transcript, serializable, forking support. |
| 24 | **Store interface for portal** | HTTP adapter that satisfies same interfaces as SQLite. CLI switches backend by config. |

## Future: Platform

| # | What |
|---|------|
| 25 | Portal MVP (web UI, API, auth, workspace sync) |
| 26 | Team workspaces + shared sandboxes |
| 27 | Session sharing (view, co-drive, fork) |
| 28 | Org memory graph (workspace → team → org propagation) |
| 29 | Scheduled/event-driven sessions |
| 30 | Cloud sandbox providers (Docker, OpenShell, Modal, Daytona) |
| 31 | Multi-agent workflows (coordinator + delegation) |
| 32 | Gateway capture (egress proxy, real-time structured capture) |
| 33 | Marketplace (personas, skills, connectors) |

## CLI Design Notes (revisit post-gateway)

Current `ok profile` conflates auth + config + extensions. Compare with GitHub CLI:
- `gh auth` = identity (login, switch, token)
- `gh config` = preferences (flat key-value)
- Extensions are a first-class top-level concept

Potential split when auth story gets complex (multiple remotes, OAuth):
```
ok auth login <portal-url>       # authenticate with platform
ok auth status                   # who am I, which remotes
ok auth switch                   # switch profiles

ok config set/show               # flat preferences
ok extensions add/list/remove    # MCP servers as first-class
```

Also: `ok sessions` should scope to workspace:sandbox by default (not global).
Hold until gateway work is done and we refine the full command surface.

## Key Decisions

- **Entities (12):** Org, Group, Role, Gateway, Credential, Profile, Workspace, Sandbox, Session, Persona, Policy, Memory. See [entity-definitions.md](entity-definitions.md).
- **Persona = agent configuration template.** Applied at sandbox materialization. Not a runtime constraint.
- **Policy = enforcement.** What the agent can/cannot do. Stacking, cannot be loosened by lower levels.
- **Sandbox = machine + account.** HOME redirected. User explicitly creates and enters.
- **Workspace = namespace.** I/O declarations, memory scope, sandbox container.
- **Memory = workspace-scoped.** Every session contributes. Propagation to group/org admin-controlled.
- **Session = immutable audit record.** Sealed after completion.
- **Profile = portable identity.** Automations run as the user, policy restricts.
- **Gateway = enforcement point.** Credentials, network, capture, routing converge here.
- **Role = permission bundle.** Platform roles + agent governance roles. Assigned scoped.
- **Group = recursive.** Teams contain teams. Policy cascades down, memory propagates up.
- **Credential = never in sandbox.** Sentinel/gateway pattern. Agent never sees real secrets.
- **No separate service account entity for personal automations.** Policy-on-sandbox achieves restriction.
- **CLI is a portal client.** Local SQLite is just the offline store. Same interfaces, different backend.
