## Tasks

### Specification

- [x] Define the image and snapshot store as a first-class fourth storage class.
- [x] Define the image record with digest, environment_ref, provenance, and owner scope.
- [x] Define the snapshot record with digest, base image ref, reason, and retention policy.
- [x] Define content addressing and provenance rules (pin by digest, quarantine without provenance).
- [x] Define the revitalize flow against the snapshot store.
- [x] Define retention and garbage collection defaults.
- [x] Define tenant-scoped access using the canonical scope vocabulary.
- [x] Define local-first implementation (Docker cache + tarball snapshots) and distributed upgrade path (OCI + object store).
- [x] Define the store interface and events.
- [x] Wire `image_ref` / `snapshot_policy` from the Workspace profile to this store.
- [ ] Decide snapshot diffing strategy (full vs incremental over base image) for the first implementation.
- [ ] Decide signature verification policy at pull time for distributed mode.
- [ ] Validate OpenSpec with `openspec validate --all --strict`.

### Implementation Follow-up

- [ ] Add `images` and `snapshots` schema migration.
- [ ] Implement `ImageStore` interface with local Docker-daemon + tarball backend.
- [ ] Implement no-op behavior for `local`/`worktree` environments.
- [ ] Implement digest pinning and tag-to-digest resolution.
- [ ] Implement provenance recording and quarantine of provenance-less images.
- [ ] Implement snapshot create/restore and the revitalize flow.
- [ ] Implement retention GC with pinned-snapshot exemption.
- [ ] Implement tenant-scope checks with `policy.denied` emission.
- [ ] Add image/snapshot events to the event model.
- [ ] Add tests for digest pinning, provenance quarantine, retention GC, and cross-tenant denial.
