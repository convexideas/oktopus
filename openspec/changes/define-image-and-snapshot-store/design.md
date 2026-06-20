## Design

### Goal

Define the storage backend for machine images and workspace snapshots — the fourth storage class in Oktopus, distinct from relational state, run-output artifacts, and capability manifests.

The design must support:

- Base images that materialize an `environment` capability.
- Warm workspace snapshots that let a Workspace be paused, resumed, or revitalized.
- Content-addressing and provenance for supply-chain auditability.
- Retention and garbage collection (snapshots accumulate fast).
- Tenant-scoped access.
- Local-first operation (local Docker / tarballs) with a distributed upgrade path (OCI registry + object store).

### Why Not the Artifact Store

| | Artifact | Image | Snapshot |
|---|---|---|---|
| Direction | output of a run | input to execution | captured state of a running workspace |
| Scope | one run | shared across workspaces/tenants | one workspace, many over its life |
| Addressing | run/job/attempt path + sha256 | digest, OCI tag/layers | digest + parent image ref |
| Lifecycle | finalized then immutable | built/imported, tagged, pinned | created on pause/checkpoint, GC'd by retention |
| Backend (distributed) | object store | OCI registry | object store |

Images and snapshots are execution substrate, not evidence. They get their own store with its own metadata table and lifecycle.

### Three Logical Stores, One Component

```text
config  (Dockerfile / env profile)  → registry            (already exists)
image   (built or imported base)    → image store          (this change)
snapshot(warm workspace state)      → snapshot store        (this change)
```

The image store and snapshot store are presented as one backend component (`ImageStore`) with two record types, because they share addressing, provenance, retention, and tenant-scoping machinery. Snapshots reference the image they were taken from.

### Image Record

```text
id
digest                 -- content address (sha256 of manifest/config)
ref                    -- human ref, e.g. oktopus/coding-agent:1.4 or local tag
environment_ref        -- the environment capability this image materializes
source_type            -- built | imported | builtin
source_config_ref      -- registry ref of the Dockerfile/profile that produced it
build_provenance_json  -- builder identity, base image, build args, timestamp
size_bytes
media_type             -- application/vnd.oci.image.manifest.v1+json | vm-disk/qcow2 | ...
signed                 -- bool
signature_ref          -- pointer to signature metadata (not the signer impl)
owner_scope            -- tenant/org/client/project scope that owns the image
status                 -- building | available | quarantined | deprecated | deleted
created_at
metadata_json
```

Rules:

1. Images are addressed by digest. A `ref` (tag) is mutable; a digest is not. Workspaces and runs pin by digest, not tag.
2. Every non-builtin image records `source_config_ref` and `build_provenance_json`. An image with no provenance is `quarantined` and cannot back a workspace.
3. Image bytes never live in the relational DB. Metadata lives in the `images` table; bytes live in the local image cache or OCI registry.
4. A deprecated image stays resolvable for existing workspaces but is unavailable for new workspace creation unless explicitly allowed.

### Snapshot Record

```text
id
digest
workspace_id           -- the workspace this snapshot was taken from
base_image_digest      -- the image the workspace was provisioned from
parent_snapshot_id     -- previous snapshot, if incremental
reason                 -- pause | checkpoint | pre_revitalize | manual
size_bytes
uri                    -- local: runs/<...>/snapshots/<digest>.tar  | object://...
owner_scope
retention_policy       -- keep_last_n | ttl | pinned
expires_at
status                 -- creating | available | restoring | expired | deleted
created_at
metadata_json
```

Rules:

1. A snapshot captures warm workspace state (filesystem deltas over the base image; not the base image itself).
2. Revitalize flow: snapshot current workspace (`reason: pre_revitalize`) → tear down `EnvironmentInstance` → re-provision from `base_image_digest` → optionally restore the pre-revitalize snapshot or start clean. The Workspace and its history persist; only the running instance is replaced.
3. Snapshots are retained per `retention_policy`. Default `keep_last_n = 3`. Expired snapshots are GC'd; pinned snapshots are exempt.
4. Restoring a snapshot whose `base_image_digest` is `deleted`/`quarantined` fails with a clear error rather than silently starting clean.

### Local-First Implementation

```text
images:    local Docker daemon image cache, addressed by digest.
           environments of type local/worktree need no image (no-op store).
           image metadata mirrored in SQLite `images` table.
snapshots: tar of the workspace working directory under
           runs/<run-or-workspace-id>/snapshots/<digest>.tar
           snapshot metadata in SQLite `snapshots` table.
```

No external services required. `local`/`worktree` environments — the MVP default — need no image at all, so the store is a no-op until a Docker/VM environment is used.

### Distributed Upgrade Path

```text
images:    OCI registry (Harbor/ECR/GHCR/...). Pull by digest. Signature verification at pull.
snapshots: object store. Content-addressed keys. Lifecycle rules for TTL/GC.
metadata:  Postgres `images` and `snapshots` tables.
```

Adapters change; the `ImageStore` interface and the records above do not.

### Store Interface (MVP)

```text
PutImage(meta, bytesOrRef) -> digest
GetImage(digest) -> meta + local ref
ResolveImage(ref) -> digest            -- tag to digest
CreateSnapshot(workspace_id, reason) -> snapshot
RestoreSnapshot(snapshot_id) -> EnvironmentInstance handle
GCSnapshots(owner_scope) -> deleted[]
```

All operations are tenant-scoped: `owner_scope` is checked against the caller's scope before any read or write.

### Tenant Scoping

`owner_scope` on both records uses the canonical precedence ladder vocabulary (`org`/`client`/`project`). A pull or restore for a scope the caller does not belong to is denied and emits a `policy.denied` event. This is the storage-layer half of the per-tenant isolation the workspace model needs.

### Events

```text
image.built
image.imported
image.quarantined
image.deprecated
image.deleted
snapshot.created
snapshot.restored
snapshot.expired
snapshot.deleted
```

### SQLite Tables (MVP)

Two new tables, `images` and `snapshots`, with the fields above. Bytes stay out of the DB (image cache / tar files / object store). Added in a new migration; existing tables unchanged.

### Design Constraints

- Image and snapshot bytes never enter the relational DB.
- Images are pinned by digest; tags are conveniences only.
- No image backs a workspace without provenance.
- Snapshots are tenant-scoped and retention-bounded by default.
- Local default needs no external services; `local`/`worktree` environments need no image at all.
