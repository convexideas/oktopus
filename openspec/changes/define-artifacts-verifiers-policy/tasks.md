## Tasks

### Specification

- [x] Define artifact model, statuses, types, immutability, and storage rules.
- [x] Define artifact provenance and redaction metadata.
- [x] Define verifier model, verifier result records, and statuses.
- [x] Define evidence contracts and waiver rules.
- [x] Define policy decision model and enforcement points.
- [x] Define approval model and lifecycle.
- [x] Define adversarial review integration with artifacts/verifiers.
- [x] Define CI/CD, telemetry/AIOps, and inference evidence examples.
- [x] Decide if verifier results are stored as separate table, artifact metadata, or both in MVP. **Decision: both. `verifier_results` table for structured query; linked `verifier_result` artifact for durable evidence. Linked via `output_artifact_id`.**
- [x] Decide initial built-in verifier set. **Decision: `artifact-exists` and `command-exit-zero` for MVP.**
- [x] Define evidence contract enforcement path — evaluated by controller's job-completion handler, not a separate job.
- [x] Define `review_finding` severity requirement — artifacts must include `severity` in provenance_json for `max_critical_review_findings` evaluation.
- [x] Define output-kind-aware evidence contracts — acceptance branches on job `output_kind` (`artifact` requires blobs, `stream` requires delivery, `effect` requires consequence verifier). Output is not always a stored artifact.
- [x] Validate OpenSpec with `openspec validate --all`.

### Implementation Follow-up

- [ ] Add artifact metadata schema and filesystem storage layout.
- [ ] Add verifier result schema (table + linked artifact).
- [ ] Add policy decision schema with `attempt_id`.
- [ ] Add approval schema with `attempt_id`.
- [ ] Add artifact finalization with sha256 digest.
- [ ] Add `artifact-exists` verifier.
- [ ] Add `command-exit-zero` verifier.
- [ ] Add evidence contract evaluation in `CompleteAttempt` handler.
- [ ] Add waiver flow.
- [ ] Add event emission for artifacts, verifiers, policies, approvals, and waivers.
- [ ] Add tests for evidence contract acceptance, failure, and `max_critical_review_findings` enforcement.
