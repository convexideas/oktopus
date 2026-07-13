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
