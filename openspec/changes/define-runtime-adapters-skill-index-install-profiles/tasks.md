## Tasks

### Specification

- [x] Define runtime adapter contract.
- [x] Define runtime capability matrix dimensions.
- [x] Define skill source index fields and loading rules.
- [x] Define delegation policy rules.
- [x] Define staged capability install/activation pipeline.
- [x] Define package profiles/presets.
- [x] Define configurator vs control-plane boundary.
- [x] Validate OpenSpec with `openspec validate --all --strict`.

### Implementation Follow-up

- [x] Add `runtime_adapter` interface package.
- [x] Add runtime capability matrix schema and storage table.
- [x] Add skill source index table and file indexer.
- [x] Add exact-path skill handoff type.
- [ ] Add delegation rules policy schema.
- [ ] Add capability install plan command.
- [ ] Add capability activation pipeline records.
- [ ] Add profile/preset manifest support.
- [ ] Add initial profile: `repo-governance/basic`.
