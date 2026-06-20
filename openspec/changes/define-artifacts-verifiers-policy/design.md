## Design

### Goal

Define artifacts, verifiers, policies, and approvals as first-class Oktopus primitives.

The design must support:

- Evidence-backed agent workflows.
- Local filesystem artifacts in MVP.
- Object-store artifacts later.
- Verifier gates for tests, builds, diffs, reports, screenshots, evals, and review findings.
- Policy-enforced permissions outside prompts.
- Human approval gates for risky actions.
- Audit and compliance records.
- CI/CD, AIOps, documentation, and client inference solution packs.

### Conceptual Flow

```text
job executes
  → emits artifacts
  → verifier jobs inspect artifacts/external systems
  → policy evaluates requested actions
  → approval gates resolve risky decisions
  → workflow accepts/rejects/repairs based on evidence
```

### Artifact Model

Artifacts are durable typed outputs or evidence.

Artifact record:

```text
id
run_id
job_id
attempt_id
step_id
type
name
uri
media_type
size_bytes
sha256
status
producer_ref
provenance
retention_policy
redaction_level
created_at
finalized_at
metadata
```

Artifact statuses:

```text
created
uploading
finalized
quarantined
redacted
deleted
expired
```

Artifact types:

```text
spec
plan
task_list
patch
diff
log
trace
graph
report
screenshot
browser_recording
eval_result
model_output
verifier_result
approval_record
policy_decision
review_report
review_finding
incident_report
runbook
adr
release_note
postmortem
dataset_manifest
golden_set
confusion_matrix
```

### Artifact Immutability

Finalized artifacts are content-addressed evidence.

Rules:

1. Artifact bytes are immutable after `finalized`.
2. Metadata fields for retention, redaction, quarantine, and deletion may change through events.
3. Replacement creates a new artifact record.
4. Artifact references use ID and digest, not mutable path alone.
5. Every finalized artifact records producer capability and attempt.

### Local and Distributed Storage

Local MVP:

```text
runs/<run-id>/artifacts/<artifact-id>/<name>
runs/<run-id>/jobs/<job-key>/attempts/<n>/artifacts/<name>
runs/<run-id>/logs/<job-key>-attempt-<n>.log
```

SQLite stores metadata. Filesystem stores bytes.

Distributed future:

```text
object://<bucket>/<project>/<run>/<artifact-id>/<name>
```

Postgres stores metadata. Object store stores bytes. Event stream records creation/finalization/deletion.

### Provenance

Artifact provenance captures where evidence came from.

```text
producer_type: worker|controller|human|external_system
producer_id
capability_ref
command_or_tool_ref
input_artifact_ids
source_uri
source_revision
redaction_policy
```

Examples:

- `graphify` emits `graph` and `report` artifacts.
- `openspec validate --all` emits `verifier_result`.
- adversarial reviewer emits `review_report` and `review_finding`.
- CI webhook emits `log`, `trace`, and `test_output` artifacts.
- VLM evaluator emits `eval_result`, `confusion_matrix`, and `dataset_manifest`.

### Verifier Model

A verifier is a capability that produces evidence about whether a condition holds.

Verifier manifest:

```yaml
kind: verifier
name: command-exit-zero
version: 0.1.0
inputs:
  command: string
  cwd: string
outputs:
  artifact_type: verifier_result
policy:
  filesystem: scoped_workspace
  network: false
```

Verifier result record:

```text
id
run_id
job_id
attempt_id
verifier_ref
status
summary
started_at
finished_at
input_artifact_ids
output_artifact_id
exit_code
error_message
metadata
```

`verifier_results` is both a dedicated table (for structured query: "did go-test pass on run X?") and linked to a `verifier_result` artifact (for durable evidence attachment). They are linked via `verifier_results.output_artifact_id → artifacts.id`. Both are created; the table row is the queryable index, the artifact is the immutable evidence record.

Verifier statuses:

```text
pending
running
passed
failed
errored
skipped
waived
```

Verifier examples:

```text
command-exit-zero
artifact-exists
json-schema-valid
openspec-validate
go-test
npm-test
lint
typecheck
screenshot-match
security-scan
review-findings-empty
accuracy-threshold
latency-budget
telemetry-recovered
```

### Evidence Contracts

Workflow jobs can declare evidence contracts. The contract is interpreted according to the job's `output_kind` (`artifact | stream | effect`, defined in `define-core-run-job-event-model`), because not every job produces a storable blob — some stream a conversational reply, others act on an external system.

```yaml
output_kind: artifact          # artifact | stream | effect (default: artifact)
outputs:
  required_artifacts:          # used when output_kind = artifact
    - type: patch
    - type: verifier_result
      verifier: go-test
acceptance:
  verifiers:                   # consequence checks; primary signal for output_kind = effect
    - openspec-validate
    - artifact-exists
  delivery:                    # used when output_kind = stream
    destination: required      # message must be delivered to an approved destination
    transcript: optional       # transcript may be logged but is not a required artifact
  max_critical_review_findings: 0
  required_approvals:
    - production-deploy
```

**Output-kind acceptance:**

- `artifact` — the job must produce its `required_artifacts`, finalized for the attempt. This is the default and matches prior behavior.
- `stream` — the job must deliver its message to an approved output destination; a stored artifact is **not** required. A transcript may be captured as evidence but is optional. Acceptance is delivery, not a blob.
- `effect` — the job is accepted on its **consequence**, verified by `acceptance.verifiers` against the changed system (tests pass, file present, endpoint healthy, deploy reported). The diff/commit/command may be captured as evidence, but the acceptance criterion is the effect verifier, not a stored output.

**Enforcement:** The contract is evaluated by the controller's job-completion handler, not by a separate job. When a worker calls `Complete(lease_id, attempt_id, result)`:

```text
1. Controller verifies lease is active and attempt matches.
2. Branch on output_kind:
   - artifact: verify all required_artifacts exist and are finalized for this attempt.
   - stream:   verify delivery to an approved destination succeeded (delivery event recorded).
   - effect:   verify acceptance.verifiers all passed for this attempt (consequence check).
3. For each verifier in acceptance.verifiers: checks verifier_results WHERE
   verifier_ref = ? AND attempt_id = ? AND status = 'passed'.
4. If max_critical_review_findings set: counts review_finding artifacts attached to
   this job's run WHERE provenance.severity = 'error' (critical maps to error severity).
5. All applicable checks pass → job.status = succeeded, emit job.succeeded.
6. Any check fails → job.status = failed, emit job.failed with reason = evidence_contract_unmet.
```

Only the most recent attempt's artifacts and verifier results count toward the contract. Prior attempt artifacts are retained for audit but not evaluated.

`review_finding` artifacts must include a `severity` field in their provenance metadata (`debug|info|warn|error`) so the controller can evaluate `max_critical_review_findings`. Workers that emit `review_finding` artifacts without a severity in provenance will cause contract evaluation to fail with `evidence_contract_error`.

Rules:

1. A job cannot be marked succeeded unless its evidence contract — interpreted for its `output_kind` — is satisfied or all unmet items are explicitly waived.
2. A verifier failure blocks dependent jobs unless workflow policy allows failure.
3. Waivers require actor, reason, expiration, and event record.
4. Final run summary must list evidence and waivers.
5. A `stream` job with no delivery destination and an `effect` job with no consequence verifier are configuration errors caught at run creation, not at completion.

### Policy Decision Model

Policy decisions record whether an action is allowed, denied, or escalated.

Policy decision record:

```text
id
run_id
job_id
attempt_id
step_id
policy_ref
subject_type
subject_id
action
resource
outcome
reason
requires_approval
approval_id
created_at
metadata
```

Outcomes:

```text
allowed
denied
requires_approval
```

Policy action categories:

```text
tool_use
secret_read
network_egress
filesystem_read
filesystem_write
source_mutation
destructive_command
capability_activation
production_deploy
external_api_call
budget_exceeded
human_data_access
```

### Policy Enforcement Points

Policies are checked at:

1. Run creation.
2. Job eligibility.
3. Lease grant.
4. Tool/runtime invocation.
5. Artifact publication.
6. Capability activation.
7. Deployment/remediation action.
8. Waiver creation.

Prompt instructions can guide agents, but policy decisions must be controller/worker-enforced.

### Approval Model

Approvals gate actions that policy cannot auto-allow.

Approval record:

```text
id
run_id
job_id
step_id
type
status
requested_by
requested_at
resolved_by
resolved_at
expires_at
reason
payload
```

Approval statuses:

```text
requested
approved
rejected
expired
canceled
```

Approval types:

```text
destructive_command
secret_access
network_access
source_mutation
production_deploy
capability_activation
waiver
client_data_access
remediation_execute
```

Rules:

1. Approval must include requested action, resource, risk, and proposed command/tool details.
2. Approval grant is scoped to run/job/action/resource and optional expiration.
3. Approval does not grant broad permanent permissions unless explicitly modeled as capability/policy update.
4. Approval resolution emits events and policy decision linkage.

### Adversarial Review Integration

Adversarial review produces artifacts and gates.

Review artifacts:

```text
review_report
review_finding
```

Verifier examples:

```text
review-findings-empty
no-critical-review-findings
security-review-passed
architecture-review-accepted
```

Review policy:

- high-risk jobs require review reports before acceptance
- critical findings block downstream jobs unless waived
- waiver requires approval and rationale
- repair jobs link to findings they address

### CI/CD Integration

CI/CD workflows use artifacts/verifiers for receipts.

Examples:

```text
build.log → artifact: log
junit.xml → artifact: test_report
coverage.json → artifact: eval_result/report
openspec validate output → verifier_result
PR diff → artifact: diff
```

Failed-CI remediation run acceptance can require:

- reproduced failure artifact
- diagnosis report
- patch/diff artifact
- tests passing verifier result
- review findings resolved or waived

### Telemetry/AIOps Integration

Telemetry workflows use artifacts/verifiers for incident evidence.

Examples:

```text
alert payload → artifact: report
trace sample → artifact: trace
log excerpt → artifact: log
metric graph → artifact: screenshot/report
RCA → artifact: incident_report
runbook action → policy_decision + approval_record
recovery check → verifier_result
```

Remediation acceptance can require:

- blast-radius report
- proposed action policy decision
- approval for risky remediation
- recovery verifier from metrics/logs
- incident/postmortem artifact

### Client Inference Integration

Inference solution packs use artifacts/verifiers for product quality.

VLM examples:

```text
dataset_manifest
golden_set
eval_result
confusion_matrix
human_review_report
latency_report
bias_check_report
```

Deployment acceptance can require:

- accuracy threshold passed
- latency budget passed
- human review sample passed
- data boundary policy passed
- rollback plan artifact exists

### Redaction and Sensitive Data

Artifacts may contain sensitive data.

Fields:

```text
redaction_level: none|metadata_only|redacted|restricted
retention_policy
access_policy_ref
```

Rules:

1. Secret values must not be stored in artifacts unless explicitly allowed.
2. Sensitive artifacts require access policy.
3. Redacted artifacts preserve provenance link to restricted originals where allowed.
4. External sharing uses redacted copies by default.

### Events

Minimum artifact events:

```text
artifact.created
artifact.uploading
artifact.finalized
artifact.redacted
artifact.quarantined
artifact.deleted
artifact.expired
```

Minimum verifier events:

```text
verifier.started
verifier.passed
verifier.failed
verifier.errored
verifier.skipped
verifier.waived
```

Minimum policy/approval events:

```text
policy.evaluated
policy.allowed
policy.denied
policy.requires_approval
approval.requested
approval.approved
approval.rejected
approval.expired
approval.canceled
```

### Design Constraints

- No final success without evidence contract satisfaction or explicit waiver.
- Artifact bytes stay out of relational DB.
- Finalized artifact bytes are immutable.
- Waivers are auditable exceptions, not silent bypasses.
- Policy is enforced outside prompt text.
- Review findings become structured evidence, not prose lost in chat.
