# Oktopus

A governed operating system for agentic work — a control plane that gives agents shared capabilities, context, policy, memory, and audit across any runtime.

Models and agent runtimes come and go. The operating layer should belong to you.

## What it does

Oktopus sits between you and your AI agents. It:

- **Wraps any coding agent** (Claude Code, Kiro, Codex, Cursor, etc.) in a governed session
- **Tracks every session** with structured capture for audit, replay, and memory extraction
- **Enforces policy** — approval gates, spend caps, tool restrictions — in code, not prompts
- **Owns your context** — sessions, memory, and knowledge persist across runtimes and switch with you
- **Composes capabilities** — skills, tools, personas, workflows from a governed registry

The control plane owns state; agents plug in as disposable workers.

## Quick start

```bash
go build -o ok ./cmd/ok/

# Launch Claude Code in a tracked session
./ok run claude-code

# List your sessions
./ok sessions list
```

Environment variables:
- `OK_PROXY_ADDR` — MITM proxy address for structured API capture (optional)
- `OK_LOG_DIR` — directory for session log files (optional)
- `OK_DB_PATH` — SQLite database path (default: `~/.ok/oktopus.db`)

## Architecture

```
┌──────────────────────────────────────────────────────┐
│  ok run <agent>                                       │
└──────────────────┬───────────────────────────────────┘
                   ▼
┌──────────────────────────────────────────────────────┐
│  Policy Hooks (pre/post session)                      │
└──────────────────┬───────────────────────────────────┘
                   ▼
┌──────────────────────────────────────────────────────┐
│  Harness Adapter (PTY passthrough)                    │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐       │
│  │Claude Code │ │   Kiro     │ │  (future)  │       │
│  └────────────┘ └────────────┘ └────────────┘       │
└──────────────────┬───────────────────────────────────┘
                   ▼
┌──────────────────────────────────────────────────────┐
│  Session Manager (SQLite) + MITM Capture (optional)   │
└──────────────────────────────────────────────────────┘
```

## Capability Registry

Capabilities are versioned manifests under `registry/`:

- `adapters/` — runtime bridges (CI/CD, ticketing, data stores)
- `skillpacks/` — external skill/persona/workflow packs
- `tools/` — callable actions and integration bridges
- `personas/` — operating profiles with role + allowed tools + output contracts

Imported packs are read-only by default. Local overrides shadow them by policy.

## Design principles

1. **Control plane owns state; agents are disposable workers.** Switch the runtime, keep the context.
2. **Policy in code, not prompts.** Permissions, approvals, and budgets are enforced at execution time.
3. **Not coding-specific.** Any workflow, any domain — a news digest, a label processor, a code review.
4. **Enterprise auth is first-class.** Granular ACLs per entity (agent, tool, skill, session), scoped to user/group/role/org.
5. **Organizational memory.** A temporal knowledge graph with org → team → user propagation, captured from sessions, constantly updating.

## Roadmap

| Phase | What | Status |
|-------|------|--------|
| 1 | Harness orchestrator + session capture | ✅ Built |
| 2 | Memory capture + extraction pipeline (MITM → compactor → knowledge) | Next |
| 3 | Capability registry + agent builder (skills, tools, personas, MCP) | Planned |
| 4 | Web UI + collaboration (chat, agent builder GUI, team features) | Planned |
| 5 | Org memory graph with scoped propagation | Planned |

## Project structure

```
cmd/ok/              CLI entry point
internal/
  harness/           Adapter interface + implementations (claude/, kiro/, …)
  policy/            Pre/post session hooks
  session/           Session lifecycle (DB)
  cli/               Command routing
  db/                SQLite schema + migrations
proxy/               MITM proxy addon for structured API capture
registry/            Capability manifests (adapters, skillpacks, tools, personas)
docs/                Architecture docs, design notes
```

## Documentation

- [Concepts](docs/concepts.md) — mental model, entities, capabilities
- [Extending](docs/extending.md) — connect your systems, governed supply chain
- [IO Capture Design](docs/design/io-capture.md) — MITM-based session recording

## License

MIT
