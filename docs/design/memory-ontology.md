# Memory Ontology

Defines the schema for Oktopus's knowledge graph — what kinds of entities exist,
how they relate, and what properties they carry.

Used today to structure LLM extraction prompts. Used later as the actual graph schema.

## Design Principles

- **Session-derived**: all knowledge originates from session observations
- **Scoped**: entities belong to a workspace, user, or are global
- **Temporal**: relationships have timestamps — knowledge evolves
- **Hierarchical**: org → team → user scoping (future)

## Node Types

### Project
A codebase or body of work.

| Property | Type | Example |
|----------|------|---------|
| name | string | "oktopus" |
| language | string[] | ["go", "python"] |
| workspace_id | ref | links to workspace entity |

### Component
A logical part of a project (module, service, feature area).

| Property | Type | Example |
|----------|------|---------|
| name | string | "auth", "CLI", "memory" |
| path | string | "internal/identity/" |

### File
A specific file in the project.

| Property | Type | Example |
|----------|------|---------|
| path | string | "internal/cli/run.go" |
| role | string | "entry point", "config", "test" |

### Decision
An architectural or design choice made during a session.

| Property | Type | Example |
|----------|------|---------|
| summary | string | "use cobra for CLI" |
| rationale | string | "ecosystem standard, shell completions" |
| session_id | ref | when it was made |

### Pattern
A recurring practice, preference, or convention.

| Property | Type | Example |
|----------|------|---------|
| name | string | "explicit error handling" |
| scope | string | "user" or "project" |

### Tool
An external tool, library, or dependency.

| Property | Type | Example |
|----------|------|---------|
| name | string | "squirrel", "koanf", "cobra" |
| purpose | string | "SQL query builder" |

### Person
A user or contributor.

| Property | Type | Example |
|----------|------|---------|
| id | string | "local" |
| preferences | ref[] | links to Pattern nodes |

### Issue
A problem identified during a session.

| Property | Type | Example |
|----------|------|---------|
| summary | string | "auth middleware doesn't validate tokens" |
| severity | string | "high", "medium", "low" |
| status | string | "open", "resolved" |

## Edge Types

| From | Edge | To | Meaning |
|------|------|----|---------|
| Project | HAS_COMPONENT | Component | logical containment |
| Component | IMPLEMENTED_IN | File | where the code lives |
| File | DEPENDS_ON | File | import/reference |
| Project | USES | Tool | dependency |
| Session | PRODUCED | Decision | what was decided |
| Session | IDENTIFIED | Issue | what was found |
| Decision | AFFECTS | Component | what it changes |
| Decision | SUPERSEDES | Decision | evolution over time |
| Person | PREFERS | Pattern | user preferences |
| Project | FOLLOWS | Pattern | project conventions |
| Issue | RELATES_TO | Component | where the problem is |
| Issue | RESOLVED_BY | Decision | how it was fixed |

## Scoping Rules

- **Project-level**: Components, Files, Decisions, Issues — belong to a workspace
- **User-level**: Patterns (preferences) — belong to a profile
- **Global**: Tools — shared knowledge (e.g., "squirrel is a Go SQL builder")

## Temporal Properties

Every edge carries:
- `created_at` — when the relationship was first observed
- `session_id` — which session established it
- `confidence` — how certain (explicit statement vs inference)

This allows:
- "What did we know about auth *before* the refactor?"
- "Which decisions were made in the last week?"
- "Has this pattern been consistent or is it evolving?"

## Extraction Prompt Template

When extracting knowledge from an episode, the LLM is prompted with this schema:

```
Given the following session transcript, extract entities and relationships.

Entity types: Project, Component, File, Decision, Pattern, Tool, Issue
Relationship types: HAS_COMPONENT, IMPLEMENTED_IN, DEPENDS_ON, USES, PRODUCED,
  IDENTIFIED, AFFECTS, SUPERSEDES, PREFERS, FOLLOWS, RELATES_TO, RESOLVED_BY

For each entity, provide: type, name, properties
For each relationship: from, edge_type, to, confidence (high/medium/low)

Only extract what is explicitly stated or strongly implied. Do not speculate.

Transcript:
---
{episode_content}
---
```

## Storage Mapping

### Phase 3 (now): Flat facts in SQLite

Extracted entities stored as facts with structured scope:
```
Fact: "oktopus uses cobra for CLI"
Scope: "workspace:<id>"
Source: "session:<id>"
```

The ontology guides extraction quality but storage is flat.

### Phase 5 (future): Graph DB

Migrate to KùzuDB or similar embedded graph. Entities become nodes, relationships
become edges. Full Cypher queries. The ontology becomes the literal schema.

---

## The Learning Loop

Inspired by Hermes Agent's self-improving skill loop, but scoped to the
organization rather than a single agent instance.

### Three Memory Layers

```
┌─────────────────────────────────────────────────────────┐
│ Layer 3: Organizational Knowledge                        │
│ Scope: org → team → workspace                           │
│ Propagates down to new workspaces/members automatically  │
│ Admin-gated for cross-team sharing                       │
└──────────────────────────┬──────────────────────────────┘
                           │ crystallizes from
┌──────────────────────────▼──────────────────────────────┐
│ Layer 2: Procedural Memory (Patterns + Facts)            │
│ Scope: workspace                                        │
│ Detected by pattern analysis across episodes             │
│ Confidence-scored, versioned, temporal                   │
└──────────────────────────┬──────────────────────────────┘
                           │ extracted from
┌──────────────────────────▼──────────────────────────────┐
│ Layer 1: Episodic Memory (Raw)                           │
│ Scope: session                                          │
│ Immutable. Append-only. What happened.                   │
│ Source: IO stream capture + gateway captures             │
└─────────────────────────────────────────────────────────┘
```

### Layer 1: Episodic (implemented)

What we have today. Raw session output stored in the episodes table.

- Source: stdout capture, Pi JSON mode, gateway request/response
- One episode per session (or per capture chunk)
- Immutable after creation
- Used for: audit, replay, LLM summarization

### Layer 2: Procedural (to build)

Extracted knowledge with confidence and temporal properties.

**What lives here:**
- Facts: "this project uses Stripe for payments"
- Patterns: "always run migrations before deploy"
- Preferences: "team prefers explicit error handling over panic"
- Decisions: "chose cobra because shell completions"
- Skills: "to deploy this service: helm upgrade payments-prod"

**How it's populated:**

```
Trigger: after every N sessions (configurable, default=3) OR on explicit
         `ok memory extract` command

Process:
  1. Collect recent episodes for this workspace
  2. Collect existing procedural knowledge (for deduplication)
  3. LLM prompt: "Given these sessions and what you already know,
     extract NEW facts, patterns, and decisions. Score confidence."
  4. Merge: new facts added, existing facts updated (confidence bumped
     if confirmed, lowered if contradicted)
  5. Supersession: if new decision contradicts old one, mark old as
     superseded (temporal evolution)
```

**Schema:**

```
knowledge_items:
  id           TEXT PRIMARY KEY
  workspace_id TEXT NOT NULL
  kind         TEXT NOT NULL  -- "fact", "pattern", "preference", "decision", "skill"
  content      TEXT NOT NULL  -- natural language description
  scope        TEXT NOT NULL  -- "workspace", "team", "org"
  confidence   REAL NOT NULL  -- 0.0 to 1.0
  source_sessions TEXT        -- JSON array of session IDs that contributed
  created_at   TEXT NOT NULL
  updated_at   TEXT NOT NULL
  superseded_by TEXT          -- NULL or ID of newer item that replaces this
```

### Layer 3: Propagation (to build)

Knowledge flows upward by policy:

```
workspace → team:  automatic when confidence > 0.8 AND seen in 3+ sessions
                   OR admin-approved via portal

team → org:        admin-approved only (prevents noise)

Downward inheritance:
  new workspace in team → inherits team knowledge at creation
  new team member → their agent sees team + org knowledge
```

**Propagation rules:**

```yaml
# Example: workspace-level policy (in workspace.yaml or portal)
memory:
  propagation:
    auto_promote_threshold: 0.8    # confidence to auto-promote to team
    min_sessions: 3                # must be seen in N sessions
    admin_review_for_org: true     # org-level always needs approval
    inherit_from_team: true        # new workspaces get team knowledge
    inherit_from_org: true         # new workspaces get org knowledge
```

### The Pattern Detector

Runs periodically (post-session or scheduled). Compares episodes across
sessions to find recurring themes.

**Signals that indicate a pattern:**
- Same concept mentioned in 3+ sessions
- Same tool/command used repeatedly
- Same error encountered and resolved the same way
- Same architectural decision referenced

**Implementation (phase 1 — LLM-based):**

```
Input:
  - Last N episodes for this workspace
  - Existing knowledge items (to avoid re-extracting)

Prompt:
  "Compare these sessions. What patterns emerge?
   What is consistently true across sessions that isn't yet captured
   in the existing knowledge base?
   
   Existing knowledge:
   {existing_items}
   
   Recent sessions:
   {episode_summaries}
   
   Extract new patterns. For each:
   - content: what the pattern is
   - kind: fact | pattern | preference | decision | skill
   - confidence: how certain (0.0–1.0)
   - evidence: which sessions demonstrate this"
```

**Implementation (phase 2 — hybrid):**
- Embedding similarity to cluster related episodes
- LLM only for the final crystallization step
- Cheaper, more scalable

### Knowledge Injection (Pre-Session)

Before a session starts, relevant knowledge is loaded into the agent's context.

```
1. Resolve scope: workspace → team → org (inheritance chain)
2. Filter: only items with confidence > 0.5, not superseded
3. Rank: by relevance to task (if --task provided), recency, confidence
4. Format: structured context block injected into system prompt
5. Cap: max N items to avoid context bloat (configurable)
```

**Injection template:**

```
## Context from organizational memory

### Project facts
- This project uses Stripe for payments (confidence: 0.95)
- Deploys via helm to EKS (confidence: 0.9)

### Team patterns
- Always run migrations before deploy
- Use explicit error handling, not panic
- PRs require 2 approvals

### Recent decisions
- [2026-07-15] Chose cobra for CLI (rationale: shell completions)
- [2026-07-10] Moved from sqlx to sqlc (rationale: type safety)
```

### Comparison with Hermes

| Aspect | Hermes Agent | Oktopus |
|--------|-------------|---------|
| What's learned | Executable Python skills | Natural language knowledge + structured facts |
| Who benefits | The single agent instance | Everyone in the org |
| Persistence | ~/.hermes/skills/ (files) | Knowledge graph in DB (syncs to portal) |
| Trigger | After every task | After N sessions or on-demand |
| Scope | Per-user, per-agent | workspace → team → org |
| Portability | Dies with the user | Survives employee turnover |
| New member experience | Starts cold | Inherits team + org knowledge day 1 |

### What Hermes does better (and we should adopt)

1. **Executable skills** — Hermes generates actual code (Python functions) that
   run faster than re-prompting the LLM. We should support this too: a "skill"
   is a knowledge item with `kind: skill` that contains executable instructions
   or even a script.

2. **Automatic trigger** — Hermes detects patterns without being asked. Our
   pattern detector should also be automatic (post-session hook), not just
   manual `ok memory extract`.

3. **Confidence decay** — if a pattern hasn't been observed in recent sessions,
   its confidence should decay over time. Hermes handles this implicitly (old
   skills get overwritten). We need explicit temporal decay.

---

## Example: Full Learning Loop

```
Session 1 (Engineer A):
  "Set up Stripe webhook handler. Used stripe-go SDK. 
   Tested with stripe listen --forward-to localhost:4242"

Session 3 (Engineer A):
  "Fixed webhook signature verification. The secret is
   in STRIPE_WEBHOOK_SECRET env var."

Session 5 (Engineer B, same workspace):
  "Adding new payment method. Using stripe-go.
   Webhook endpoint at /api/webhooks/stripe"

→ Pattern Detector runs after session 5:

  New knowledge extracted:
    fact: "This project uses stripe-go SDK for Stripe integration" (0.95)
    fact: "Webhook secret stored in STRIPE_WEBHOOK_SECRET env var" (0.85)
    skill: "To test webhooks locally: stripe listen --forward-to localhost:4242" (0.9)
    fact: "Webhook endpoint is /api/webhooks/stripe" (0.8)

→ Confidence > 0.8, seen in 3+ sessions → auto-promoted to team level

Session 8 (Engineer C, NEW workspace in same team):
  Agent's system prompt includes:
    "## Team knowledge
     - Projects in this team use stripe-go for Stripe integration
     - Webhook secrets stored in STRIPE_WEBHOOK_SECRET
     - Test webhooks with: stripe listen --forward-to localhost:<port>"

  Engineer C asks: "How do I set up Stripe here?"
  Agent answers accurately from team knowledge, not from scratch.
```

---

## Build Sequence for Learning Loop

| # | What | Depends on | Effort |
|---|------|-----------|--------|
| 1 | knowledge_items table + store methods | Store interfaces | Small |
| 2 | Extraction prompt + `ok memory extract` command | Episodes exist | Medium |
| 3 | Auto-trigger: run extraction after every session | Session lifecycle | Small |
| 4 | Knowledge injection: pre-session loading | Workspace config | Medium |
| 5 | Pattern detector: cross-session analysis | Multiple episodes | Medium |
| 6 | Propagation: workspace → team → org | Portal / multi-user | Large (portal phase) |
| 7 | Confidence decay + supersession | Knowledge items exist | Small |
| 8 | Executable skills (code generation) | Trust + sandbox | Future |

## Example: What a session produces

Session transcript: "Reviewed auth middleware. Found that tokens aren't validated
on the refresh endpoint. Suggested adding validation. Also noticed the project
uses bcrypt for password hashing."

Extracted:
```
Nodes:
  Component(name="auth middleware", path="middleware/auth.go")
  Issue(summary="tokens not validated on refresh endpoint", severity="high", status="open")
  Tool(name="bcrypt", purpose="password hashing")
  Decision(summary="add token validation to refresh endpoint")

Edges:
  Issue RELATES_TO Component("auth middleware")
  Issue RESOLVED_BY Decision("add token validation")
  Decision AFFECTS Component("auth middleware")
  Project USES Tool("bcrypt")
```
