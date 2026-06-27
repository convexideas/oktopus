# Architecture

This document formalizes Oktopus's layered architecture, component boundaries, and build order. It is the structural contract that prevents drift during implementation.

## Layers

Oktopus is four layers, bottom to top. Each layer is independently useful and enriches the layers below it when added.

```
┌─────────────────────────────────────────────────────────┐
│  4. Portal                                               │
│     Web UI for browse/discover/manage (future)           │
├─────────────────────────────────────────────────────────┤
│  3. Governance                                           │
│     Orgs, teams, users, roles, policies, approvals       │
├─────────────────────────────────────────────────────────┤
│  2. Hub (Registry + Social)                              │
│     Install, publish, fork, search, versioning           │
├─────────────────────────────────────────────────────────┤
│  1. Runtime                                              │
│     Harness, workers, agent loop, runs, evidence         │
└─────────────────────────────────────────────────────────┘
```

### Layer 1 — Runtime

The execution engine. A harness expands workflows into runs, assigns jobs to workers, and collects evidence.

| Component | Responsibility |
|-----------|---------------|
| **Harness** | Workflow expansion, run creation, job scheduling, state transitions |
| **Worker** | Polls for jobs, executes via LLM + tools, reports results |
| **Hub interface** | Model gateway abstraction — one `Complete(messages, tools)` call, adapter per provider |
| **Tool bridge** | Registry tools become callable functions the LLM can invoke |
| **Evidence store** | Artifacts, events, step logs |

The first worker is a local agent loop: plan → tool-call → observe → repeat. It consumes capabilities from the registry and reports back to the harness. The harness/worker boundary is the seam that later enables remote workers, parallel execution, and policy gates — without rewriting the core.

### Layer 2 — Hub

The capability marketplace. Makes the local registry **federated** — local-first, pull from remotes.

| Component | Responsibility |
|-----------|---------------|
| **Remote index** | Searchable catalog of available capabilities (Git-backed initially) |
| **Resolver** | `hub://author/name@version` → fetch manifest + dependencies |
| **Lock file** | Pins installed capabilities, tracks provenance |
| **CLI commands** | `install`, `publish`, `fork`, `search`, `browse` |

Capabilities flow: remote → local registry → runtime consumes them. The hub adds discovery and sharing; it does not change how the runtime reads capabilities.

### Layer 3 — Governance

Policy enforcement at every seam. Decides *who can do what* across the system.

| Component | Responsibility |
|-----------|---------------|
| **Identity model** | Orgs → teams → users |
| **Role system** | owner, maintainer, developer, viewer (extensible) |
| **Policy engine** | Evaluates rules at enforcement points |
| **Approval gates** | Blocks actions that require human sign-off |
| **Scope hierarchy** | `builtin < upstream_pack < org < client < project < workflow < run_override` |

Governance is called by other layers, not a separate service:

| Caller | Question |
|--------|----------|
| `hub install` | Is this capability approved for this org? |
| `hub publish` | Does this user have publish rights? |
| `run workflow` | Is this user authorized to trigger this? |
| Tool call at runtime | Is this tool allowed in this context? |
| Promote capability | Does user have the role to move draft → approved? |

Early implementation: a policy file evaluated in-process. Later: a proper decision engine.

**Enterprise auth integration:** Oktopus never owns user identity. The identity model is a provider interface — OIDC, SAML, LDAP, or custom SSO adapters slot in behind it. Oktopus accepts a token, maps claims to a Subject (user + org + teams), resolves roles from its own policy layer (or imports group mappings from the IdP), and evaluates policy. No shadow user directory, no new login flows. Org/team membership can sync from SCIM or be managed locally — the Governor doesn't care where the identity came from, only what roles it resolves to.

### Layer 4 — Portal (future)

A web UI on top of the same data. Read-layer over the Git-backed hub index and governance state. Not built until layers 1–3 are solid.

## Language boundaries

Go and TypeScript split along the harness/worker seam:

| Concern | Language | Rationale |
|---------|----------|-----------|
| Harness, scheduling, state, DB | Go | Single binary, deterministic, infra-grade reliability |
| Governance, policy evaluation | Go | Hot-path, must be fast and correct |
| CLI entry point | Go | Single binary distribution, fast startup |
| Worker / agent loop | TypeScript | LLM SDKs, MCP client, tool ecosystem live here |
| Hub remote/index operations | TypeScript | npm-like patterns, JSON-native, shares types with portal |
| Portal (future) | TypeScript | Web-native |

**Integration protocol:** The harness (Go) spawns and communicates with workers (TS) over JSON via stdio or local HTTP. This is the same boundary that later enables remote/containerized workers — the protocol doesn't change, only the transport.

**Shared contracts:** Types shared across languages are defined as OpenAPI schemas (for REST/HTTP surfaces) or Protocol Buffers (for internal RPC and high-throughput paths), stored in `openspec/`. Code generation produces Go structs and TypeScript types — no hand-synced duplicates.

**Build systems:**

| System | Purpose | Production? |
|--------|---------|-------------|
| `go.mod` | Harness, CLI, governance | Yes |
| `package.json` | Worker, hub, portal | Yes |
| `pyproject.toml` | MkDocs Material docs build | No — CI/dev only |

Python is not a runtime language in this project. It exists solely for documentation tooling.

**Repo layout:**

```
cmd/oktopus/        Go — CLI + harness binary
internal/           Go — harness, governance, db, scheduling
worker/             TypeScript — agent loop, LLM calls, tool bridge, MCP
hub/                TypeScript — remote index, publish
openspec/           JSON Schema contracts (language-agnostic)
registry/           YAML manifests (consumed by both)
```

## Component map

```
cmd/oktopus/          CLI entry point
internal/
  harness/            Run creation, workflow expansion, job scheduling
  worker/             Local agent loop (the openclaw-like core)
  hub/                Model gateway interface + provider adapters
  registry/           Capability loading, validation, indexing
  remote/             Hub: fetch, publish, resolve from remote sources
  governance/         Policy evaluation, org/role model
  db/                 Storage (SQLite local, Postgres later)
  evidence/           Artifacts, events, verifiers
```

## Key interfaces

```go
// Hub — model gateway (Layer 1)
type Hub interface {
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
}

// Governor — policy decisions (Layer 3)
type Governor interface {
    Evaluate(ctx context.Context, action Action, subject Subject) Decision
}

// Remote — hub operations (Layer 2)
type Remote interface {
    Search(query string, opts SearchOpts) ([]Capability, error)
    Fetch(ref CapabilityRef) (*Manifest, error)
    Publish(manifest *Manifest, opts PublishOpts) error
}
```

## Build order

Priority: **be productive now** (working agent loop) while preserving architecture seams for later layers.

### Phase A — Agent loop (Runtime core)

Get `oktopus run workflow:sdlc-default` executing steps through a local LLM worker.

1. Run creation — expand workflow into jobs, persist state
2. Hub interface — `Complete()` with OpenAI-compatible adapter (covers OpenRouter, LiteLLM, local)
3. Worker loop — poll jobs, call LLM with tools, mark done
4. Tool bridge — registry tools become callable functions
5. Step reporting — outputs/evidence written back to run

**Skip for now:** multi-worker coordination, verifier execution, policy enforcement, importer.

### Phase B — Hub (federated registry)

Get `oktopus hub install @author/capability` pulling from remotes.

1. Remote index format — define manifest + catalog schema (Git-backed)
2. `hub install` — fetch from remote, merge into local registry
3. `hub publish` — push local manifest to remote index
4. `hub search` — query the catalog
5. Lock file — pin versions, track what's installed vs. local

### Phase C — Governance

Get `policy.yaml` enforced at runtime and hub operations.

1. Identity model — org/team/user in config or DB
2. Role system — simple RBAC
3. Policy file format — declarative rules
4. Enforcement points — wire `Governor.Evaluate()` into harness, hub, CLI
5. Approval gates — block + notify on restricted actions

### Phase D — Portal

Web UI. Depends on all above being stable. Not scoped yet.

## Invariants

These hold across all phases and must not be violated:

1. **Registry is the source of truth for capabilities.** The runtime never hard-codes what tools/skills exist.
2. **Harness/worker boundary is always a function call, never shared state.** This is the seam for distribution.
3. **Governance is a call, not a layer dependency.** Layers 1 and 2 work without governance present — they just skip policy checks.
4. **Hub adds discovery, not new capability formats.** A remote capability and a local capability have the same manifest schema.
5. **Config picks adapters; code uses interfaces.** Provider choice (model, storage, hub backend) is always configuration, never conditional logic in business code.

## What this replaces in the roadmap

The original roadmap (Phases 0–8) remains valid as a *spec-completeness* sequence. This architecture doc adds a *productivity-first build order* (Phases A–D) that can be executed in parallel with spec work:

| Architecture phase | Roadmap phases covered |
|--------------------|----------------------|
| A (agent loop) | 1 (Run spine) + 2 (Worker loop) — minimal |
| B (hub) | 7 (Marketplace) — CLI subset |
| C (governance) | 4 (Governance) |
| D (portal) | New — not in current roadmap |

The roadmap phases for evidence (3), sessions (5), extensibility (6), and distributed (8) layer on top once the A→B→C skeleton is running.
