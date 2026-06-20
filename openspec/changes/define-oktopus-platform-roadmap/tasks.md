## Tasks

### Phase 0 — OpenSpec Foundation

- [x] Create Oktopus OpenSpec change process documentation.
- [x] Draft platform roadmap proposal.
- [x] Draft high-level distributed system design.
- [x] Draft initial requirement deltas for the Oktopus platform.
- [ ] Validate OpenSpec with `openspec validate --all`.
- [ ] Review and approve platform boundary: Oktopus as control plane, not personal assistant or LLM runtime.

### Phase 1 — Core Spine

- [ ] Define canonical schemas for Capability, Run, Job, Step, Attempt, Artifact, Event, Approval, Worker, Environment, and Policy.
- [ ] Implement SQLite-backed run/job/event store.
- [ ] Implement capability registry loader and validator.
- [ ] Implement CLI commands for registry validation, run creation, run listing, job listing, and event tailing.
- [ ] Add basic run directory/artifact layout.

### Phase 2 — Worker Protocol

- [ ] Define worker request/accept/lease/heartbeat/complete/fail/cancel protocol.
- [ ] Implement local worker that polls and leases jobs.
- [ ] Add lease timeout and retry behavior.
- [ ] Add cancellation propagation.
- [ ] Add worker labels and platform matching.

### Phase 3 — Evidence and Policy

- [ ] Implement artifact store with digest/provenance metadata.
- [ ] Implement command verifier and artifact-exists verifier.
- [ ] Implement approval gates.
- [ ] Implement allowlisted secrets and tool permissions.
- [ ] Emit structured events for every important state transition.

### Phase 4 — Initial Adapters

- [ ] Add shell adapter.
- [ ] Add Pi worker adapter.
- [ ] Add Graphify tool adapter.
- [ ] Add OpenSpec tool adapter.
- [ ] Add Beads/Archon adapter stubs.
- [ ] Import `addyosmani/agent-skills` as a read-only skillpack.

### Phase 5 — Distribution

- [ ] Split all-in-one local mode from server/worker mode.
- [ ] Add Postgres option.
- [ ] Add object store option.
- [ ] Add queue/event backend option.
- [ ] Add Docker worker image.
- [ ] Add deployment manifests for Docker Compose or Nix.

### Phase 6 — Integrations and Solution Packs

- [ ] Add CI/CD webhook adapter.
- [ ] Add failed-CI remediation workflow.
- [ ] Add OpenTelemetry/log/metric ingest adapter.
- [ ] Add incident triage and remediation workflow.
- [ ] Add documentation workflow pack.
- [ ] Add VLM harness solution pack proposal.
