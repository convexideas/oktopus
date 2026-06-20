## Why

Oktopus must not trust agent claims without receipts. Workflows need durable artifacts, verifier results, policy decisions, and approval records so operators can inspect what happened, prove whether work is complete, and safely gate risky actions.

Artifacts are the evidence layer. Verifiers are the proof layer. Policies and approvals are the governance layer. Together they turn agent output into auditable, resumable, and operationally safe work.

## What Changes

- Define the artifact model for durable evidence and outputs.
- Define artifact storage rules for local-first and distributed deployments.
- Define verifier capabilities and verifier result records.
- Define policy decision records for tools, secrets, network, filesystem, deployments, budgets, and capability activation.
- Define approval gate records and lifecycle.
- Define how artifacts, verifiers, policies, and approvals interact with workflow jobs and adversarial review.
- Define evidence requirements for final run acceptance.

## Out of Scope

- Implementing artifact storage.
- Implementing a full policy language.
- Implementing cryptographic signing or legal retention enforcement.
- Building UI for approvals.
- Integrating external policy engines such as OPA yet.

## Impact

This change establishes the safety and evidence model. Future implementation can add artifact stores, verifier workers, policy checks, approval APIs, and UI views without changing workflow semantics.
