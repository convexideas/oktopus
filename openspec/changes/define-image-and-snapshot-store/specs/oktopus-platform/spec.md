## ADDED Requirements

### Requirement: Image and Snapshot Store

Oktopus SHALL provide an image and snapshot store as a first-class storage backend, distinct from the relational state store, the run-output artifact store, and the capability registry.

#### Scenario: Store an image separately from artifacts

- **WHEN** an environment base image is built or imported
- **THEN** Oktopus records it in the image store with a content digest and metadata
- **AND** does not store it through the run-output artifact store, because images are execution inputs, not run evidence

#### Scenario: Local-first default needs no external services

- **WHEN** a run uses a `local` or `worktree` environment
- **THEN** Oktopus requires no image and the image store is a no-op
- **AND** when a Docker environment is used, images are cached via the local Docker daemon and mirrored in the SQLite `images` table

### Requirement: Content-Addressed Images with Provenance

Oktopus SHALL address images by content digest and SHALL record build provenance for every non-builtin image.

#### Scenario: Pin by digest, not tag

- **WHEN** a workspace or run references an image
- **THEN** Oktopus pins the exact image digest
- **AND** a mutable tag change does not alter what an existing workspace or run resolves to

#### Scenario: Reject images without provenance

- **WHEN** an image has no recorded source config reference or build provenance
- **THEN** Oktopus marks it quarantined
- **AND** refuses to provision a workspace from it until provenance is recorded

### Requirement: Workspace Snapshots with Retention

Oktopus SHALL capture warm workspace state as snapshots with content addressing, a reference to the base image, and a retention policy.

#### Scenario: Revitalize a workspace

- **WHEN** a user revitalizes an interactive workspace
- **THEN** Oktopus snapshots the current workspace state, tears down the running environment instance, and re-provisions from the recorded base image digest
- **AND** the workspace identity and history persist while only the running instance is replaced

#### Scenario: Garbage-collect expired snapshots

- **WHEN** snapshots exceed the retention policy (default keep last 3)
- **THEN** Oktopus deletes expired snapshots and emits `snapshot.deleted`
- **AND** snapshots marked pinned are exempt from garbage collection

#### Scenario: Restore against a missing base image

- **WHEN** a snapshot is restored but its base image digest is deleted or quarantined
- **THEN** Oktopus fails the restore with an explicit error
- **AND** does not silently start the workspace from a clean image

### Requirement: Tenant-Scoped Image and Snapshot Access

Oktopus SHALL scope every image and snapshot to an owner scope and SHALL deny cross-tenant access.

#### Scenario: Deny cross-tenant pull

- **WHEN** a caller requests an image or snapshot whose owner scope the caller does not belong to
- **THEN** Oktopus denies the request and emits a `policy.denied` event
- **AND** the denial is recorded as a policy decision

### Requirement: Distributed Upgrade Path for Images and Snapshots

Oktopus SHALL allow the image and snapshot store to move from local backends to distributed backends without changing image or snapshot records or the store interface.

#### Scenario: Swap local store for distributed store

- **WHEN** Oktopus is deployed in distributed mode
- **THEN** images resolve from an OCI registry by digest and snapshots resolve from an object store
- **AND** image and snapshot metadata, digests, and the store interface are unchanged from local mode
