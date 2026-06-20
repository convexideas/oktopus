## Why

Archon is already part of the Convex Ideas agent tooling stack and is intended for explicit, structured AI workflows. Oktopus should lean on Archon where it adds value, especially for workflow proposal, workflow execution as a bounded job, and repeatable process authoring.

However, workflows encode authority: tools, secrets, mutation rights, approval gates, triggers, and future automation. Archon or any agent must not silently install or activate workflows. Oktopus must treat Archon as a governed adapter whose outputs become artifacts and proposals unless explicitly approved.

## What Changes

- Define Archon as an Oktopus `adapter` and `tool` capability.
- Define supported Archon modes: proposal generation, explicit workflow execution, and workflow catalog import.
- Define no-auto-install rule for workflows and other capabilities.
- Define proposal artifact locations and activation gates.
- Define event, artifact, policy, and approval records for Archon use.
- Define how Archon-produced workflows become Oktopus workflow capabilities after validation and review.

## Out of Scope

- Implementing the Archon CLI adapter.
- Defining Archon's internal workflow syntax.
- Automatically translating every Archon workflow feature into Oktopus DAG semantics.
- Granting Archon direct registry write access.

## Impact

This change lets Oktopus benefit from Archon without surrendering governance. Archon can help create and run structured workflows, but Oktopus remains the system of record for active capabilities, run state, policies, approvals, and audit.
