# Roadmap

The implementation plan, sequenced by the principles: build the **source-of-truth spine** first, then **evidence**, then **governance**, then **sessions and extensibility**. This reconciles with the phasing in `define-oktopus-platform-roadmap` — that change owns the canonical phasing; this page is its readable view.

## Status at a glance

The spine's foundations exist today: the Go CLI, SQLite migrations, the capability registry loader with a SQLite index, and seed manifests. The execution layer (run creation, worker loop, verifiers) is the next build.

## Phases

### Phase 0 — Spine foundations ✅ (in place)
Go module + Cobra CLI, SQLite migrations (`0001`, `0002`), registry loader writing the capability index, `db init` / `db migrate`, seed capabilities and starter workflows.

### Phase 1 — Run & event spine
Run creation and workflow→job expansion, transactional state-transition helpers that emit events, project auto-creation with race-free run numbering.
*Specs:* `define-core-run-job-event-model`, `define-local-mvp-cli-storage`

### Phase 2 — Worker loop & leases
Eligible-job scheduler query, atomic lease acquisition, heartbeats and expiry reconciliation with retry-policy enforcement, a local shell/stub worker.
*Specs:* `define-worker-lease-protocol`

### Phase 3 — Evidence
Filesystem artifact store with sha256 finalization, `artifact-exists` and `command-exit-zero` verifiers, output-kind-aware evidence-contract evaluation (artifact / stream / effect).
*Specs:* `define-artifacts-verifiers-policy`

### Phase 4 — Governance
Policy decision recording at enforcement points, approval gates, capability proposal → validate → activate flow.
*Specs:* `define-capability-registry-model`, `define-artifacts-verifiers-policy`, `define-archon-adapter`

### Phase 5 — Sessions, workspaces & images
Workspace provisioning from profiles, image & snapshot store (local Docker + tarballs), pause/resume/revitalize, scoped memory.
*Specs:* `define-image-and-snapshot-store`, `define-enterprise-admin-marketplace-sessions`

### Phase 6 — Extensibility & sources
Configurable capability/knowledge sources, scope-hierarchy resolution, secrets references, MCP and adapter entry points.
*Specs:* `define-client-server-runtime-config-sources`

### Phase 7 — Marketplace & enterprise
Internal catalog, third-party providers, standardized vs. custom modes, enterprise governance controls.
*Specs:* `define-enterprise-admin-marketplace-sessions` (+ a future `define-marketplace-providers` for external publisher identity and verification)

### Phase 8 — Distributed mode
Swap adapters: Postgres, object store, queue/event backend, remote/containerized worker pools. Concepts unchanged.
*Specs:* upgrade-path sections across all changes

## Beyond
CI/CD, telemetry/AIOps, documentation, and client-inference solution packs — forward-looking requirements owned by `define-oktopus-platform-roadmap`, built on the proven spine.
