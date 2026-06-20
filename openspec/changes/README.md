# OpenSpec Change Process

## Overview

This directory tracks active and archived OpenSpec changes for Oktopus.

Use OpenSpec's current spec-driven schema for new changes. Archived changes live under `archive/` and are historical context.

## Directory Structure

```text
openspec/changes/
├── archive/
└── <change-id>/
    ├── proposal.md
    ├── design.md
    ├── tasks.md
    └── specs/
        └── <spec-name>/
            └── spec.md
```

## Requirement Deltas

Each changed specification should include one or more delta sections:

```markdown
## ADDED Requirements
## MODIFIED Requirements
## REMOVED Requirements
```

Each requirement should include at least one concrete scenario with `**WHEN**` and `**THEN**` statements.

## Validation

Before finalizing a change, run:

```bash
openspec validate --all --strict
```

## Archive Ordering

All current changes add requirements to the single `oktopus-platform` spec. Because several changes touch overlapping domains, they MUST be archived in dependency order so the canonical spec stays coherent. Archive sequence:

1. `define-oktopus-platform-roadmap` — umbrella positioning, phasing, and delegation. Owns only high-level and forward-looking (CI/CD, telemetry, docs, inference) requirements; delegates spine behavior to the changes below.
2. `define-core-run-job-event-model` — durable entities, lifecycle states, event log, lease data-model.
3. `define-capability-registry-model` — capability kinds, manifests, scopes, lifecycle, proposals.
4. `define-worker-lease-protocol` — operational lease/heartbeat/retry/cancel protocol.
5. `define-artifacts-verifiers-policy` — artifacts, verifiers, evidence contracts, policy/approval enforcement.
6. `define-local-mvp-cli-storage` — local CLI + SQLite implementation of the above.
7. `define-archon-adapter` — Archon as a governed adapter on top of the spine.
8. `define-client-server-runtime-config-sources` — client/server control plane and configurable sources.
9. `define-image-and-snapshot-store` — image/snapshot storage backend; must precede the workspace/sessions change that references `image_ref` and `snapshot_policy`.
10. `define-enterprise-admin-marketplace-sessions` — enterprise catalog, marketplace, interactive workspaces, and resumable sessions; depends on #9 for workspace images and snapshots.

Rules:

- The roadmap (#1) is umbrella-only. It must not restate behavioral requirements owned by #2–#5. If a spine concept needs to change, modify the owning change, not the roadmap.
- Changes #2–#5 define their domain at different abstraction layers (data model vs operational protocol vs MVP usage). When archiving, a later change that refines an earlier requirement uses `## MODIFIED Requirements` against the same requirement name rather than `## ADDED` a parallel one.
- The canonical precedence ladder `builtin < upstream_pack < org < client < project < workflow < run_override` is shared by `define-capability-registry-model` and `define-client-server-runtime-config-sources`; it must not be redefined.

## Guidelines

- One change per directory.
- Keep changes focused and atomic.
- Use specs for durable decisions, not chat memory.
- Keep canonical requirements under `openspec/specs/` current with implemented behavior.
