# Build Plan

Captured: 2026-07-12

## Design Principles

- **Idiomatic Go** — stdlib-first, interfaces at consumption boundaries, small packages
- **Clean Architecture** — domain types are pure structs, business logic in app layer, infrastructure adapts
- **Pi-first** — build deep for one harness, generalize later
- **Small increments** — each change is reviewable, testable, independently useful

## Core Mental Model

```
Profile (who I am)  +  Persona (what role)  +  Workspace (what code)  →  Session (in a sandbox)
```

- **Profile** — user's complete runtime identity: extensions, model prefs, tool permissions, harness-specific config. Deeply personal. Enterprise adds mandatory overrides later.
- **Persona** — a portable role definition: system prompt, skills, output format. Not bound to a user.
- **Workspace** — declares what resources exist (source refs, connectors, knowledge). Never mutated.
- **Session** — execution instance. Creates a sandbox, materializes profile + persona + workspace into it, launches harness, captures output.
- **Sandbox** — isolated directory where everything lands. User's workspace is never clobbered.

## Architecture

```
internal/
  domain/          pure structs (zero deps)
  store/           persistence interfaces
  store/sqlite/    SQLite adapter (sqlx + squirrel)
  config/          app config (koanf: defaults → file → env)
  parser/          YAML manifest → domain types (koanf + validator)
  harness/         Harness interface + ProcessAdapter base
  harness/pi/      Pi-specific adapter (full config translation)
  harness/codex/   (future)
  harness/kiro/    (future)
  policy/          Hook interface
  cli/             cobra commands + App struct (DI)
```

## Completed

| # | What |
|---|------|
| 1 | Persona manifest schema + parser |
| 2 | `ok capabilities add/list` |
| 3 | `ok workspace create/list/show` |
| 4 | `ok run --persona --workspace --runtime` + resolution chain |
| 5 | Clean architecture refactor (domain, store, config, parser, App DI) |
| 6 | Cobra CLI + harness adapters registered (pi, codex, kiro, claude-code) |

## Phase 2: Profile + Sandbox + Pi Deep Customization

Focus: make `ok run -r pi` the best possible governed Pi experience.

| # | What | Details |
|---|------|---------|
| 7 | **Profile entity** | Schema: extensions (MCP servers), model prefs, tool permissions, harness-specific sections. Stored in DB. `ok profile show`, `ok profile set pi.extensions [...]` |
| 8 | **Sandbox abstraction** | Interface: `Create(workspace, profile, persona) → SandboxDir`. Simplest provider: temp dir with workspace source linked/copied in. Session launches harness pointing at sandbox. |
| 9 | **Session assembly** | Merge: profile + persona + workspace → materialized sandbox. For Pi: `.pi/extensions.json` (profile), `.pi/SYSTEM.md` (persona), `.pi/skills/` (persona), workspace source (workspace). |
| 10 | **Pi adapter: full config translation** | Translate Config → Pi flags: `--model`, `--system-prompt`/`--append-system-prompt`, `--tools`/`--exclude-tools`, env vars. Materialize files into sandbox (not workspace). |
| 11 | **Pi adapter: non-interactive mode** | `--print -p <prompt>` for headless runs. Exposed via `ok run -r pi --task "review this code"`. |
| 12 | **End-to-end verification** | `ok run -p code-reviewer -r pi -w my-project` creates sandbox, materializes persona files, launches Pi with correct flags, records session. |

## Phase 3: Memory

Focus: sessions produce durable knowledge that improves future sessions.

| # | What | Details |
|---|------|---------|
| 13 | **Episodic capture** | Post-session: capture raw output (stdout/log) as an episode bound to the session. |
| 14 | **Memory extraction** | Periodic or on-demand: distill episodes → semantic facts (Mem0 API or local extraction). |
| 15 | **Pre-session injection** | Before launch: query semantic memory for relevant context, inject into persona system prompt or `.pi/SYSTEM.md`. |
| 16 | **Memory scoping** | Per-workspace, per-persona, or global. User controls what context carries over where. |

## Phase 4: Generalize + Workflows

Focus: apply the Pi patterns to other harnesses, add multi-step orchestration.

| # | What | Details |
|---|------|---------|
| 17 | **Codex adapter** | `AGENTS.md` materialization, `codex -q` non-interactive. |
| 18 | **Kiro adapter** | `.kiro/steering/*.md` materialization. |
| 19 | **Workflow capability** | Sequential sessions with memory propagation. User defines: "plan → implement → review". Each step = persona + runtime. Memory flows between steps. |
| 20 | **Enterprise profile overrides** | Admin policies layered on top of user profiles. Mandatory extensions, model restrictions, tool blocklists. |

## Phase 5: Advanced

| # | What |
|---|------|
| 21 | Multi-profile (remote registries, org contexts) |
| 22 | Temporal knowledge graph (Graphiti evaluation) |
| 23 | Gateway capture mode (MITM proxy) |
| 24 | Native capture mode for open harnesses (Pi SDK/RPC) |
| 25 | A2A interop |

## Session Assembly (detailed)

```
1. User runs: ok run -p code-reviewer -r pi -w my-project

2. Resolution:
   - Profile → load user's profile from DB (extensions, model prefs, Pi config)
   - Persona → load code-reviewer from capabilities registry (system prompt, skills, native)
   - Workspace → load my-project source ref

3. Sandbox creation:
   - Create temp dir (or worktree, container — provider-dependent)
   - Link/copy workspace source into sandbox
   - Materialize profile files:    .pi/extensions.json
   - Materialize persona files:    .pi/SYSTEM.md, .pi/skills/*
   - Materialize native overrides: persona.native.pi.files → sandbox

4. Harness launch:
   - Pi adapter builds CLI args from Config (model, tools, etc.)
   - cmd.Dir = sandbox dir
   - Start process, record session

5. Post-session:
   - Capture output → episodic store
   - Clean up sandbox (or preserve for debugging)
```

## Profile Schema (draft)

```yaml
# Stored in DB, editable via ok profile set
profile:
  # Global preferences
  default_model: claude-sonnet-4
  default_runtime: pi

  # Per-harness config
  pi:
    extensions:
      - name: github-mcp
        command: npx
        args: ["-y", "@modelcontextprotocol/server-github"]
        env:
          GITHUB_TOKEN: "${GITHUB_TOKEN}"
      - name: filesystem-mcp
        command: npx
        args: ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
    model: claude-sonnet-4
    tools_exclude: [computer_use]

  codex:
    model: gpt-4.1
    approval_mode: suggest

  kiro:
    steering_files:
      - ~/shared-steering/always-include.md
```

## Key Decisions

- **Profile is per-user, local-first.** Multi-profile (org, team) comes later.
- **Sandbox is mandatory.** Even the simplest "directory" provider creates a temp dir. No clobbering workspaces.
- **Pi first.** Build deep for Pi, patterns will generalize to others.
- **Memory connects sessions.** Without memory, sessions are isolated. Memory is the value multiplier.
- **Workflows are sequential sessions with memory.** No separate DAG engine needed initially — user runs steps manually, memory carries over. Formal workflow engine is a convenience layer added when the pattern proves out.
- **Extensions are profile, not persona.** Personas are portable across users. Extensions are personal tooling config.
