## Tasks

### Specification

- [x] Define Oktopus client/server architecture.
- [x] Define message, command, and trigger input surfaces.
- [x] Define configurable source categories for skills, tools, runtimes, adapters, verifiers, policies, and knowledge.
- [x] Define knowledge source/vault integration hooks.
- [x] Define configuration hierarchy.
- [x] Define workflow selection authority.
- [x] Define Source/Blueprint/Proposal/Workflow/Run/Job/Step taxonomy.
- [x] Define context scoping before worker handoff.
- [x] Define background agent workers and handoff contract.
- [ ] Validate OpenSpec with `openspec validate --all`.

### Implementation Follow-up

- [ ] Add config schema for sources and routing.
- [ ] Add source loader abstraction.
- [ ] Add skill source loader for filesystem and git.
- [ ] Add tool source loader for local YAML manifests.
- [ ] Add knowledge source interface and filesystem markdown stub.
- [ ] Add router stub for imperative commands.
- [ ] Add message ingestion stub that creates recommendation artifacts, not auto-runs.
- [ ] Add worker handoff context bundle type.
- [ ] Add registry support for workflow blueprints.
