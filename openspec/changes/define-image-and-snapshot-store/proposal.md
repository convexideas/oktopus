## Why

The `define-enterprise-admin-marketplace-sessions` change introduced interactive workspaces with a profile that references `image_ref`, `snapshot_policy`, and `persistence_policy`. Nothing in any spec defines the backend those fields point at. `image_ref` dangles: there is no store for base machine images, no store for warm workspace snapshots, and no rules for how either is built, versioned, addressed, retained, or scoped per tenant.

Oktopus already specifies three storage classes — relational state (SQLite/Postgres), run-output artifacts (filesystem/object store), and capability manifests (registry files + SQLite index). Machine images and workspace snapshots are a distinct fourth class and cannot reuse the artifact store: artifacts are per-run outputs, while images are shared execution inputs with OCI-registry semantics, and snapshots are point-in-time captures of running workspaces with their own retention lifecycle.

Without this backend, the Environment capability and Workspace profile cannot be implemented, and "revitalize the container" (recreate a workspace while keeping its history) has nowhere to store the warm state it restores from.

## What Changes

- Define an Image & Snapshot Store as a first-class storage backend, separate from the artifact store.
- Define content-addressed base images with provenance linking each image to the config that built it.
- Define workspace snapshots (warm captures of a running workspace) with retention and garbage collection.
- Define tenant-scoped access so one tenant cannot pull another tenant's images or snapshots.
- Define the local-first implementation (local Docker daemon / image cache + tarball snapshots under `runs/`) and the distributed upgrade path (OCI registry + object store).
- Wire the store to the existing `environment` capability and the Workspace profile `image_ref` / `snapshot_policy` fields.

## Out of Scope

- Building specific OCI registry backends (Harbor, ECR, GHCR, etc.); only the interface and local default.
- Image build pipelines beyond recording build provenance (the actual builder is an environment/tool capability).
- Cryptographic signing implementation; this change records signing metadata and pinning but does not implement a signer.
- VM/microVM disk image formats beyond treating them as opaque content-addressed blobs.
- Cross-region replication and CDN concerns.

## Impact

This change closes the dangling `image_ref` / `snapshot_policy` references and unblocks Environment and Workspace implementation. It is additive: the relational, artifact, and registry stores are unchanged. The local-first default requires only a local Docker daemon (or none, for `local`/`worktree` environments that need no image), preserving the project's no-external-services MVP constraint.
