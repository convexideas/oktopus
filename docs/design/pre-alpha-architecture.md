# Architecture Design — Pre-Alpha

This document captures architectural decisions that need to be right before
alpha. These are the seams — hard to change once users depend on them.

---

## Landscape: How Others Solve This

| Framework | Center of gravity | Gateway | Session model | Memory | Config |
|-----------|------------------|---------|---------------|--------|--------|
| **OpenShell** (NVIDIA) | Sandbox runtime | Reverse proxy (inference.local). Mandatory in container. Supervisor enforces. | Gateway owns lifecycle. Supervisor reports state. | Structured logs (OCSF). No learning. | YAML policy + gateway config. |
| **Omnigent** | Meta-harness orchestration | L7 egress proxy (bwrap/seatbelt). Advisory on Windows. | Server-managed. Real-time sync across devices. | Session messages in DB. No extraction. | config.yaml + agent YAML. Server is source of truth. |
| **CrabTrap** (Brex) | Security proxy | Forward MITM proxy. Static rules + LLM-as-judge. iptables enforced. | N/A (proxy only, bolts onto OpenClaw). | Audit trail in Postgres. Policy builder generates from traffic. | Natural-language policy + static rules. |
| **OpenClaw** | Multi-channel gateway | WebSocket control plane. Routes messages. Skill dispatch. | Gateway-managed. Canvas for multi-step persistence. | SQLite/Redis synchronous. Explicit config. Does NOT improve over time. | Skill marketplace + JSON bindings. |
| **Hermes Agent** (Nous) | Self-improving agent | Kanban state machine. Decentralized workers. | Per-worker process. Kanban board as coordination. | Multi-layer: episodic + procedural. Pattern detector → skill generator. Learns over time. | Markdown + YAML frontmatter. Per-workspace. |

### Key insight: center of gravity determines everything

- **OpenClaw** = gateway platform. Intelligence is in routing + skills.
- **Hermes** = learning agent. Intelligence accumulates in the agent itself.
- **OpenShell** = secure runtime. Intelligence is in policy enforcement.
- **Omnigent** = meta-harness. Intelligence is in orchestration.
- **Oktopus** = organizational memory. Intelligence accumulates across the org.

We are closest to Hermes in *ambition* (memory that improves over time) but
closest to Omnigent in *mechanism* (meta-harness that wraps existing agents).
Our unique position: organizational propagation. None of them go from
user → team → org. None of them own the knowledge graph.

---

## 1. Gateway: Advisory vs Mandatory

### The Problem

Our gateway sets `ANTHROPIC_BASE_URL` as an env var. If a harness ignores it,
the gateway is bypassed entirely.

### Decision: Two-tier enforcement

```
Tier 1 (always): Advisory gateway via env vars.
  - Works for harnesses that respect base URL overrides.
  - Claude Code, Codex, openai-agents all respect these.
  - Zero overhead, zero setup.

Tier 2 (opt-in, requires sandbox type=container or type=vm):
  - All egress blocked except through gateway.
  - iptables (Linux) / pf (macOS) rules in the container.
  - Only possible when we control the network namespace.
  - local/process driver CANNOT enforce this.
```

### Implication

For local/process, the gateway is advisory. For container/VM sandbox types,
the gateway is mandatory. We don't invest in bwrap/seatbelt for the process
case — if you want enforcement, use a container.

When the gateway is bypassed (advisory mode), we still have IO stream capture
as fallback. Both channels feed memory.

---

## 2. Session Lifecycle

### The Problem

Session identity is generated in `run.go` and state transitions are scattered.
No clear ownership, no state machine.

### Decision: Session is a state machine in runtime/

```
States:
  created → running → completed | failed | cancelled

Transitions:
  created:   ID assigned, record written, pre-checks pass
  running:   harness process started successfully
  completed: harness exits 0
  failed:    harness exits non-zero
  cancelled: user SIGINT / policy deny mid-session
```

Session owns: ID, scope (workspace:sandbox), agent, state, timestamps,
links to captures and episodes.

Session does NOT own: sandbox (outlives sessions), workspace (outlives
sessions), harness process (ephemeral child).

```go
type Session struct {
    ID        string
    Agent     string
    Workspace string
    Sandbox   string
    State     SessionState  // created, running, completed, failed, cancelled
    CreatedAt time.Time
    StartedAt time.Time
    EndedAt   time.Time
}

func NewSession(agent, workspace, sandbox string) *Session { ... }
func (s *Session) Start() { ... }
func (s *Session) Complete() { ... }
func (s *Session) Fail() { ... }
func (s *Session) Cancel() { ... }
```

CLI calls these transitions. It doesn't own the logic.

### Comparison

- Omnigent: server manages sessions. Real-time sync across devices.
- Hermes: Kanban board per task. Workers claim tasks atomically.
- Us: simpler — single-user CLI. Session is just an audit record with state.
  Portal adds sync later without changing the local model.

---

## 3. Workspace Config: Source of Truth

### The Problem

Workspace config lives in DB as a JSON blob. Where does it come from? How is
it edited? When the portal exists, who wins?

### Decision: File + DB, like git

```
Local-only mode:  workspace.yaml → DB (sync on read)
Portal mode:      portal → DB → workspace.yaml (portal authoritative)
```

### File location

```
~/.ok/<workspace>/workspace.yaml    — managed workspaces
.ok.yaml (in project root)          — project-embedded workspace (like .git)
```

### Workspace YAML shape

```yaml
name: myproject
source: /Users/me/code/myproject

sandbox:
  provider: local
  type: process

gateway:
  provider: anthropic
  # credentials resolved separately

defaults:
  harness: pi
  model: claude-sonnet-4-20250514
```

### Credentials: never in workspace.yaml

Stored separately:
- Environment variables (always checked, current behavior)
- `~/.ok/credentials.yaml` (future: local credential store)
- Portal credential store (future: fetched on demand)

Resolution: portal > credentials.yaml > env var

### Comparison

- Omnigent: `config.yaml` at repo root + agent YAML files. Server is truth.
- Hermes: Markdown + YAML frontmatter. Per-workspace. File-first.
- OpenClaw: Gateway daemon owns config. JSON + skill manifests.
- Us: file-first (like Hermes), syncs to portal (like Omnigent's server).

---

## 4. The Store Interface

### The Problem

One concrete SQLite store doing everything. No interface boundary. When we
add portal (HTTP API as backend), we need a clean seam.

### Decision: Interfaces in domain packages, concrete in store/

```go
// Each domain package defines its own store interface:

// runtime/store.go
type SessionStore interface {
    CreateSession(ctx, *Session) error
    GetSession(ctx, id) (*Session, error)
    ListSessions(ctx, limit) ([]Session, error)
    CompleteSession(ctx, id, status) error
}

type WorkspaceStore interface {
    CreateWorkspace(ctx, *Workspace) error
    GetWorkspaceByName(ctx, name) (*Workspace, error)
    ListWorkspaces(ctx) ([]Workspace, error)
}

// gateway/store.go
type CaptureStore interface {
    SaveCaptures(ctx, []Capture) error
    ListCaptures(ctx, sessionID) ([]Capture, error)
}

// memory/store.go
type EpisodeStore interface {
    SaveEpisode(ctx, *Episode) error
    ListEpisodes(ctx, workspaceID, limit) ([]Episode, error)
    ListSummariesByWorkspace(ctx, wsID, limit) ([]Episode, error)
}
```

Implementations:
- `store/sqlite/` — local (what we have)
- `store/portal/` — future HTTP client
- `store/composite/` — future: reads portal, writes both

### App struct holds interfaces

```go
type App struct {
    Sessions     runtime.SessionStore
    Workspaces   runtime.WorkspaceStore
    Captures     gateway.CaptureStore
    Memory       memory.EpisodeStore
    Profiles     identity.ProfileStore
    Capabilities registry.CapabilityStore
    Harness      runtime.HarnessRegistry
    Config       *config.Config
    Log          *slog.Logger
}
```

Today all point to `*sqlite.Store`. Tomorrow any can point to portal.

### Comparison

- Omnigent: Alembic-managed Postgres/SQLite. Server is the persistence layer.
  Hosts report to server. Clean client/server boundary.
- Hermes: SQLite for Kanban + skills. File-based for memory. No interface
  abstraction (monolithic Python).
- OpenShell: Gateway owns state. Drivers adapt to platform (K8s secrets,
  Docker volumes, etc.)
- Us: interface per domain. SQLite locally. Portal remotely. App wires them.

---

## 5. Gateway: Optional vs Always-On

### The Problem

Should every `ok run` start a gateway? What if no credentials? Internal sessions?

### Decision: Start when useful, skip gracefully

```
Gateway starts when:
  1. NOT an internal session (cfg.Internal = true → skip)
  2. Provider credentials can be resolved
  3. User hasn't passed --no-gateway

Gateway skipped when:
  - No API key found (user runs local Ollama via harness directly)
  - Internal session (memory summarization)
  - Explicit --no-gateway flag

When skipped:
  - No warning (expected case)
  - IO stream capture still works
  - No metering for this session
```

The gateway never blocks a session from starting. It degrades gracefully.

---

## Summary: Build Order

| # | What | Why first | Effort |
|---|------|-----------|--------|
| 1 | Store interfaces | Unlocks testing, portal, composability. Every other change touches the store. | Medium |
| 2 | Session state machine | Clean ownership, CLI simplification. Needed before we add more session features. | Small |
| 3 | Workspace YAML + resolver | User-facing truth. Blocks provider resolution from config (not just env vars). | Medium |
| 4 | Gateway graceful degradation | Small fix. Makes gateway non-breaking for users without API keys. | Small |
| 5 | Document two-tier enforcement | No code. Clarity for contributors and users. | Trivial |

---

## Patterns adopted from reference projects

| Pattern | Source | Our adaptation |
|---------|--------|----------------|
| Reverse proxy gateway | OpenShell inference.local | Same — agent talks to local endpoint |
| Credential never in sandbox | OpenShell phantom tokens | Same — gateway injects real creds |
| Policy stacking (server > agent > session) | Omnigent | org > workspace > sandbox > session |
| Agent defined in YAML | Omnigent | Persona YAML (already have this) |
| Server + host separation | Omnigent | Portal (server) + CLI (host/runner) |
| File-based config + DB cache | git | workspace.yaml + DB mirror |
| Two-tier enforcement | OpenShell (mandatory) | Advisory (process) + Mandatory (container/VM) |
| IO stream + network capture | Omnigent (PTY + L7 proxy) | stdout/JSON + gateway reverse proxy |
| Self-improving learning loop | Hermes Agent | Pattern detector + knowledge crystallizer, scoped to org |
| Procedural memory | Hermes Agent | Knowledge items with confidence, temporal decay, supersession |
| Static rules + LLM judge | CrabTrap | Policy eval + memory extraction (details below) |
| Policy from observation | CrabTrap | `ok policy suggest` — generate rules from captured traffic |
| Replay eval | CrabTrap | `ok policy test` — dry-run policy against historical sessions |

---

## CrabTrap-Inspired: Two-Tier Evaluation

CrabTrap's core insight: 97% of decisions match static rules (fast,
deterministic, microseconds). 3% go to an LLM judge (slow, nuanced). This
pattern applies to three areas of our system.

### Application 1: Policy Evaluation

```
Static rules (fast, checked first):
  allow:
    - file writes in workspace source directory
    - known-safe MCP tools
    - API calls to configured provider
  deny:
    - file writes outside workspace
    - network calls to unapproved hosts
  ask:
    - shell commands (always pause for approval)

LLM judge (long tail — only when no static rule matches):
  Input: the action, the task context, the workspace policy
  Output: ALLOW | DENY | ASK + reason
  Example: "Agent wants to run `curl | bash` — is this consistent
            with the stated task?"
```

### Application 2: Memory Extraction

Not every session produces novel knowledge. Most are routine.

```
Static rules (skip extraction):
  - Session output < 50 chars → trivial, skip
  - Session flagged internal (summarization) → skip
  - Session identical to recent session → skip

Static rules (update only):
  - Content matches existing knowledge item → bump confidence
  - Same tools/patterns used → no new extraction needed

LLM extraction (only for novel content):
  - New concepts not in existing knowledge → full extraction
  - Contradiction with existing facts → resolve + supersede
  - Cross-session pattern threshold reached → crystallize
```

This saves 80%+ of extraction LLM cost. Most sessions confirm what we
already know rather than teaching something new.

### Application 3: Gateway Traffic Policy

```
Static rules:
  - Requests to configured provider (api.anthropic.com) → allow
  - Requests to workspace-approved hosts (github.com) → allow
  - Requests to localhost/internal → allow
  - Everything else → deny (or judge)

LLM judge (when enabled):
  - "The agent is calling an unfamiliar API endpoint. Given the
    current task and workspace policy, should this be allowed?"
  - Returns: ALLOW/DENY + reason + suggested static rule for next time
```

### Policy Builder: Generate from Observation

Instead of writing policy by hand, observe real traffic and generate:

```
Phase 1: Run N sessions with gateway in observe mode (log, don't block)
Phase 2: Analyze captured traffic + tool calls
Phase 3: Generate suggested policy:

  ok policy suggest
  "Based on 12 sessions in workspace 'payments':
   Your agent typically:
     - Calls api.anthropic.com (inference)
     - Calls api.github.com (PR creation, code search)
     - Runs shell: go build, go test, git commit
     - Writes files in: internal/, cmd/, docs/

   Suggested policy:
     network_allow: [api.anthropic.com, api.github.com]
     shell_allow_patterns: [go *, git *]
     file_write_allow: [internal/**, cmd/**, docs/**]
     deny_all_else: true

   Apply? [y/n/edit]"
```

### Replay Eval: Test Before Deploy

Before a policy goes live, replay it against historical sessions:

```
ok policy test --replay 20 --policy ./strict-policy.yaml

  Results against last 20 sessions:
  ┌──────────┬────────┬─────────┬───────────────────────────────┐
  │ Session  │ Action │ Verdict │ Detail                        │
  ├──────────┼────────┼─────────┼───────────────────────────────┤
  │ abc123   │ shell  │ BLOCKED │ "rm -rf node_modules" at 14:32│
  │ def456   │ net    │ BLOCKED │ curl to pastebin.com at 09:15 │
  │ (18 more)│ ...    │ OK      │ no change                     │
  └──────────┴────────┴─────────┴───────────────────────────────┘

  2 of 20 sessions would have been affected.
  Review blocked actions? [y/n]
```

This prevents policy misconfiguration from breaking real workflows.

---

## Revised Build Sequence (Full Roadmap)

Incorporating all patterns from the landscape:

| # | What | Source | Phase |
|---|------|--------|-------|
| 1 | Store interfaces | (architectural hygiene) | Pre-alpha |
| 2 | Session state machine | Omnigent server model | Pre-alpha |
| 3 | Workspace YAML + resolver | git / Hermes file-first | Pre-alpha |
| 4 | Gateway graceful degradation | (fix) | Pre-alpha |
| 5 | knowledge_items table | Hermes procedural memory | Alpha |
| 6 | LLM extraction + `ok memory extract` | Hermes pattern detector | Alpha |
| 7 | Auto-trigger post-session | Hermes automatic loop | Alpha |
| 8 | Knowledge injection pre-session | Hermes skill loading | Alpha |
| 9 | Static policy rules | OpenShell policy YAML | Alpha |
| 10 | Policy from observation (`ok policy suggest`) | CrabTrap policy builder | Post-alpha |
| 11 | LLM-as-judge for long tail | CrabTrap | Post-alpha |
| 12 | Replay eval (`ok policy test`) | CrabTrap eval system | Post-alpha |
| 13 | Propagation (workspace → team → org) | Our unique addition | Portal phase |
| 14 | Confidence decay + supersession | Hermes skill evolution | Portal phase |

---

## Our moat (what none of them do)

| Capability | OpenClaw | Hermes | Omnigent | OpenShell | CrabTrap | Us |
|-----------|----------|--------|----------|-----------|----------|-----|
| Session capture | ✅ | ✅ | ✅ | ✅ | ✅ (audit) | ✅ |
| Memory that improves | ❌ | ✅ (per-agent) | ❌ | ❌ | ❌ | ✅ (per-org) |
| Knowledge graph | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Org propagation | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Multi-harness | ✅ (skills) | ❌ (own model) | ✅ | ✅ (any) | N/A | ✅ |
| Team collaboration | ❌ | ❌ | ✅ (real-time) | ❌ | ❌ | ✅ (portal) |
| Credential isolation | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ |
| Network policy | ❌ | ❌ | ✅ (L7 proxy) | ✅ (mandatory) | ✅ (LLM judge) | ✅ (advisory + mandatory) |
| Policy from traffic | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ (planned) |
| Policy replay/eval | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ (planned) |

Hermes learns per-agent. We learn per-org. That's the gap.
