# Specifications

Specifications are the normative, testable contracts that sit between the concepts above and the implementation tickets below. They are authored and maintained as **OpenSpec changes** in this repository — this page indexes them but does not restate them. The OpenSpec files are the single source of truth.

Each change is validated with `openspec validate --all --strict`.

## Reading order

Changes add requirements to the single `oktopus-platform` spec and are listed in dependency order:

| # | Change | Owns |
|---|---|---|
| 1 | [define-oktopus-platform-roadmap](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-oktopus-platform-roadmap) | Umbrella positioning, phasing, delegation |
| 2 | [define-core-run-job-event-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-core-run-job-event-model) | Entities, lifecycle states, event log, lease data-model |
| 3 | [define-capability-registry-model](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-capability-registry-model) | Capability kinds, manifests, scopes, lifecycle, proposals |
| 4 | [define-worker-lease-protocol](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-worker-lease-protocol) | Operational lease / heartbeat / retry / cancel protocol |
| 5 | [define-artifacts-verifiers-policy](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-artifacts-verifiers-policy) | Artifacts, verifiers, evidence contracts, policy / approval |
| 6 | [define-local-mvp-cli-storage](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-local-mvp-cli-storage) | Local CLI + SQLite implementation of the spine |
| 7 | [define-archon-adapter](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-archon-adapter) | Archon as a governed adapter |
| 8 | [define-client-server-runtime-config-sources](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-client-server-runtime-config-sources) | Client/server control plane and configurable sources |
| 9 | [define-image-and-snapshot-store](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-image-and-snapshot-store) | Image / snapshot storage backend |
| 10 | [define-enterprise-admin-marketplace-sessions](https://github.com/convexideas/oktopus/blob/main/openspec/changes/define-enterprise-admin-marketplace-sessions) | Enterprise catalog, marketplace, interactive workspaces, sessions |

## How specs map down to work

Each change contains:

- `proposal.md` — why, what changes, scope boundaries
- `design.md` — the design decisions
- `specs/oktopus-platform/spec.md` — the requirement deltas (WHEN/THEN scenarios)
- `tasks.md` — the implementation backlog (these are the **tickets**)

So the chain is: **pillar → concept → spec requirement → task → code.** The `tasks.md` files are the ticket layer that the [Roadmap](roadmap.md) sequences.
