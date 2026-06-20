## Tasks

### Specification

- [x] Define Archon as adapter/tool capability.
- [x] Define supported modes: proposal generation, explicit workflow execution, catalog import.
- [x] Define no-auto-install rule.
- [x] Define proposal artifact locations.
- [x] Define activation gate sequence.
- [x] Define policy checks for Archon runs.
- [x] Define Archon events and artifacts.
- [x] Define adversarial review triggers for Archon-generated proposals.
- [x] Confirm proposal enforcement boundary — Archon writes proposals to `runs/<run-id>/` only; activation requires explicit `oktopus registry propose` + `oktopus registry activate` commands.
- [ ] Validate OpenSpec with `openspec validate --all`.

### Implementation Follow-up

- [ ] Add `adapter:archon` and `tool:archon` registry manifests.
- [ ] Add Archon shell-wrapper worker/tool stub.
- [ ] Capture Archon stdout/stderr as artifacts.
- [ ] Add proposal artifact extraction convention.
- [ ] Add `capability proposal validate` command.
- [ ] Add activation workflow proposal.
- [ ] Add adversarial review fixture for workflow proposals.
