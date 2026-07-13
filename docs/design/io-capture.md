# IO Capture — Design Notes

## Problem

Agent sandbox sessions are ephemeral. When the sandbox dies, all working memory dies with it. We need durable capture of what happened so that memories can be extracted, aggregated across runtimes, and served back to future sessions.

## Architecture

```
ok run --agent codex --model gemini-2.0-flash
    │
    ├─ 1. Starts/connects to MITM proxy (configured, not built-in)
    │
    ├─ 2. Launches agent CLI in sandbox with:
    │     - HTTPS_PROXY → MITM proxy
    │     - CA cert injected into trust store
    │     - Workspace mounted
    │
    ├─ 3. Agent runs normally — all HTTPS traffic captured by proxy
    │
    ├─ 4. Proxy stores raw request/response logs (local or remote)
    │
    └─ 5. Session ends → metadata written

Async (later):
    Compactor process → reads raw logs → extracts episodic/semantic memories
    Memory store → serves memories back to future agent sessions
```

## Design Principles

1. **Transparent** — agent CLIs don't need modification; they just run in a configured env
2. **Proxy is external** — could be mitmproxy, LiteLLM, custom; `ok` just configures it
3. **Progressive isolation** — use whatever sandbox is available (OpenShell > Apple Containers > Podman > Docker > bare process)
4. **Capture everything** — not just inference; all HTTPS traffic (GitHub API, tool calls, etc.)
5. **Commodity compaction** — cheap models turn raw logs into useful memories async
6. **Cross-runtime aggregation** — memories keyed by meaningful identity (repo, user, project), not by sandbox

## Components

| Component | Responsibility | Build order |
|-----------|---------------|-------------|
| `ok run` | Launch agent in sandbox with proxy configured | 1st |
| MITM proxy | Capture HTTPS traffic, store raw logs | External (configured) |
| `ok sessions list` | Query captured sessions | 2nd |
| Compactor | Extract episodic/semantic memories from raw logs | Later |
| Memory store | Serve memories to future sessions | Later |

## Data Model

```sql
sessions
  id              TEXT PRIMARY KEY
  agent           TEXT        -- "codex", "claude-code", "opencode"
  model           TEXT
  started_at      TIMESTAMP
  ended_at        TIMESTAMP
  status          TEXT        -- "running", "completed", "failed"
  metadata        JSON        -- repo path, labels, etc.

turns
  id              TEXT PRIMARY KEY
  session_id      TEXT FK → sessions
  seq             INTEGER
  request         JSON        -- full request body
  response        JSON        -- full response body
  input_tokens    INTEGER
  output_tokens   INTEGER
  latency_ms      INTEGER
  created_at      TIMESTAMP
```

## MVP Scope (Milestone 1)

1. `ok run` — launches agent in sandbox with MITM proxy configured
2. Raw logs land in store (local SQLite), keyed by session/agent
3. `ok sessions list` — proves capture works

## MVP Non-Goals

- Memory extraction (compactor)
- Memory serving back to agents
- Building the proxy (it's external/configured)
- Multi-tenant auth
- Session continuation / replay

## Open Questions

- Identity anchor for cross-runtime aggregation: repo? user? project?
- How to serve memories back: workspace files? MCP tool? system prompt injection via proxy?
- Compactor trigger: time-based? session-end hook? manual?
- TLS pinning: which agents refuse proxy CA injection?

## Decisions Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-06-30 | Proxy is external, not embedded | Separation of concerns; can use mitmproxy, LiteLLM, or custom |
| 2026-06-30 | MITM model, not endpoint override | Works with all agents including OAuth/SSO ones (Kiro, Copilot) |
| 2026-06-30 | Store full JSON blobs | No lossy transformation; extract structure later via compactor |
| 2026-06-30 | No session continuation in MVP | Don't guess; focus on capture and extraction first |
| 2026-06-30 | Compactor uses commodity model | Async, cheap, can re-run as techniques improve |
| 2026-07-01 | ok CLI orchestrates; doesn't own proxy | Clean separation; proxy is a config concern |
