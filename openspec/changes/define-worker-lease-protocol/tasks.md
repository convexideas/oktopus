## Tasks

### Specification

- [x] Define worker kinds and lifecycle statuses.
- [x] Define worker capability advertisement and scheduler matching fields.
- [x] Define request/accept/lease/heartbeat/complete/fail/cancel flow.
- [x] Define atomic lease acquisition semantics.
- [x] Define TTL, heartbeat, expiration, and retry behavior.
- [x] Define at-least-once execution expectation and stale completion rejection.
- [x] Define cancellation and draining behavior.
- [x] Define event/log/artifact responsibilities.
- [x] Define adversarial review worker defaults.
- [x] Define `retry_policy` field with `idempotent|at-most-once|manual` modes; controller enforces at requeue point.
- [x] Remove `current_job_id` from worker record — `leases` table is authoritative.
- [x] Define capability pinning rule — workers use `resolved_capabilities_json` from lease grant, not live registry.
- [x] Define adversarial review context isolation — controller creates clean working directory from explicit artifact list.
- [ ] Decide HTTP vs gRPC for first remote worker API.
- [x] Validate OpenSpec with `openspec validate --all`.

### Implementation Follow-up

- [ ] Add worker, lease, and attempt schema migrations.
- [ ] Add `retry_policy_json` column to jobs table.
- [ ] Add scheduler query for eligible jobs.
- [ ] Add atomic `accept_lease` store operation.
- [ ] Add heartbeat and lease expiration reconciler with `retry_policy.mode` enforcement.
- [ ] Add stale completion rejection tests.
- [ ] Add local shell worker using same lease protocol.
- [ ] Add worker event ingestion with severity validation.
- [ ] Add cancellation and draining commands.
- [ ] Add fixture workflow with adversarial review worker jobs.
- [ ] Add `oktopus jobs retry <job-id>` command (for `manual` retry_policy mode).
