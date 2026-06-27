## Tasks

### Specification

- [x] Define core execution entities: Project, Run, Job, Attempt, Step, Lease, Worker, Artifact, Event, Approval, PolicyDecision, CapabilityRef.
- [x] Define lifecycle states for runs, jobs, attempts, leases, approvals, and policy decisions.
- [x] Define event log rules and minimum event types.
- [x] Define worker lease data-model requirements.
- [x] Define adversarial review as a first-class workflow gate using isolated review jobs/subagents.
- [x] Review whether a dedicated `ReviewFinding` table is needed now or can remain artifact/event-based for MVP. **Decision: artifact-based for MVP. `review_finding` artifacts include `severity` in provenance_json.**
- [x] Validate OpenSpec with `openspec validate --all`.
- [x] Fix run state machine — remove `planning`, `queued`, `blocked`; add explicit transition table and missing events.
- [x] Resolve Worker `current_job_id` vs `current_jobs` — removed `current_job_id` from model; `leases` table is authoritative.
- [x] Add `attempt_id` to Approval and PolicyDecision minimum fields.
- [x] Define event severity enum: `debug|info|warn|error`.
- [x] Add adversarial review `context` block with isolation semantics.
- [x] Add `output_kind` (`artifact|stream|effect`) to the Job, orthogonal to run mode; acceptance branches on it.
- [x] Define storage adapter boundary so core execution logic is not coupled to SQLite.

### Implementation Follow-up

- [ ] Create Go module skeleton.
- [ ] Add SQLite migrations for core tables.
- [x] Add initial repository/store boundary for opening and migrating the relational store.
- [ ] Add state transition helpers that always emit events (including `ExpireAttempt` with retry_policy check and `CarryForwardApproval`).
- [ ] Add CLI commands for DB init, run creation/listing/showing, job listing, and event tailing.
- [ ] Add tests for state transitions, event emission, and lease uniqueness.
- [ ] Add fixture workflow that includes adversarial review gate.
- [ ] Add event severity validation at ingestion — reject events with unknown severity.
