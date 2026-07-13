# Agent Instructions

Standing instructions for AI agents working on this project.

## Reference Frameworks

These three open-source projects are our primary inspiration. Study them for
patterns, feature parity, and architectural decisions:

- **Omnigent** (github.com/omnigent-ai/omnigent) — meta-harness for coding
  agents. Reference for: harness orchestration, PTY/native bridges, policy
  engine, OS-level sandboxing, real-time collaboration, cloud sandbox
  provisioning, agent YAML spec.

- **LibreChat** (github.com/danny-avila/LibreChat) — multi-provider AI chat
  platform (acquired by ClickHouse). Reference for: no-code agent builder,
  granular per-entity ACLs (user/group/role), Skills framework (SKILL.md),
  MCP-native tooling with deferred/programmatic tool calling, admin panel,
  enterprise auth (OAuth2/SAML/LDAP/2FA).

- **Open WebUI** (github.com/open-webui/open-webui) — self-hosted AI platform.
  Reference for: plugin/pipeline architecture, RAG with 9 vector DB backends,
  community tool marketplace, agentic memory system (hierarchical paths,
  native tool calling for memory CRUD), RBAC, hybrid search.

Our goal is to absorb the best of all three into one governed platform with
an organizational memory graph.

## Database

This is a **greenfield project**. Do NOT create incremental migration files.
The schema lives in a single file: `internal/db/migrations/0001_schema.sql`.
When the schema changes, **edit that file directly** — all developers drop and
recreate their local database during development. Incremental migrations are
only introduced once we ship to production users.

Use **sqlc** for database access. Queries live in `internal/db/queries/*.sql`,
generated Go code lives in `internal/db/dbq/`. Regenerate with `sqlc generate`
after editing queries or schema.

## Versioning

The project version is `0.1.0` (constant in `internal/cli/root.go`). Do NOT
bump it. It stays at `0.1.0` until explicitly told otherwise.

## Language & Style

- Go is the primary language. Python is used only for the MITM proxy addon and docs toolchain.
- Keep things minimal. No abstractions that aren't needed yet.
- Follow the ponytail principle: lazy means efficient. Mark intentional simplifications with `// ponytail:` comments.

## Architecture Constraints

- The control plane owns state; agents are disposable workers.
- Enterprise auth is a first-class concern (design for it now, implement later).
- The platform is not coding-specific — any workflow, any domain.
- Harness adapters are PTY passthrough (Phase 1); message-level control comes later.
- MITM proxy is optional; if not configured, sessions still work with stdout capture only.

## Future Architecture (adopt when complexity justifies)

These patterns are correct but premature today. Adopt each when its trigger fires:

- **Layered config** (knadh/koanf) — merge defaults < config file < env < CLI flags.
  Trigger: adding `~/.ok/config.yaml` or server-mode flags.

- **Repository pattern** — define `Store` interfaces in domain packages (e.g.
  `session.Store`), implement in `internal/db/sqlite/`. Consumer never imports DB.
  Trigger: adding a second storage backend (Postgres for server mode) or
  in-memory backend for tests.

- **Structured logging** (`log/slog`, stdlib) — replace fmt.Fprintf for
  operational output. Trigger: adding server mode / observability.

- **HTTP server** (stdlib `net/http` with `chi` router, or Go 1.22+ ServeMux
  patterns) — Trigger: Phase 4 (web UI + API layer).

- **Dependency injection** — constructor functions are fine for now. Consider
  `google/wire` (compile-time DI) only if wiring grows past ~10 dependencies.

- **Testing** — stdlib `testing` + `testify/assert` for readable assertions.
  Add table-driven tests as coverage grows.

Keep things manual and explicit until the trigger fires. The Go ecosystem
favors "add it when you need it" over "set up the framework first."
