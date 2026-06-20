## Tasks

### Specification

- [x] Define capability kinds.
- [x] Define common manifest envelope.
- [x] Define tool, skillpack, persona, workflow, and solution pack manifest examples.
- [x] Define lifecycle and activation gates.
- [x] Define override and precedence rules.
- [x] Define agent-created capability proposal flow.
- [x] Decide YAML vs TOML as default manifest format: YAML-first.
- [x] Decide whether local registry should be file-only or file + SQLite index in MVP. **Decision: file + SQLite index. `registry validate` writes to `capabilities` table. Workers use pinned refs from `resolved_capabilities_json`, not live registry.**
- [x] Define proposal enforcement boundary — workers write to `runs/` only; `registry/` is read-only to workers; controller mediates all activation.
- [x] Define `oktopus registry propose` and `oktopus registry activate` CLI commands.
- [x] Define `run_override` scope — represented in `resolved_capabilities_json` with `parent_capability_ref` and `reason`; set at run creation only.
- [x] Validate OpenSpec with `openspec validate --all`.

### Implementation Follow-up

- [ ] Create manifest schema package.
- [x] Implement local registry loader.
- [x] Implement capability validation command (writes to `capabilities` table on success; `--no-index` skips).
- [ ] Implement source/ref provenance capture.
- [ ] Implement scope precedence resolution.
- [ ] Implement proposal artifact directory convention under `runs/`.
- [ ] Add `oktopus registry propose --from <artifact-path>` command.
- [ ] Add `oktopus registry activate <ref> --approval <id>` command.
- [ ] Add seeded capabilities: `addyosmani-agent-skills`, `graphify`, `openspec`, `pi`, `shell`, `artifact-exists`, `sdlc-default`, and `guarded-build`.
