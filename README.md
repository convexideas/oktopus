# Oktopus

Single-machine first, distributed-ready: a vendor-agnostic control plane for agent orchestration and governance. Agent runtimes plug in as interchangeable workers.

## Step 1: Capability Registry

Capabilities are versioned manifests under `registry/`:

- `skillpacks/` — external skill/persona/workflow packs, e.g. `addyosmani/agent-skills`
- `tools/` — callable actions/adapters, e.g. Pi, Graphify, OpenShell
- `personas/` — local persona overrides
- `workflows/` — orchestrated state machines/DAGs
- `verifiers/` — evidence gates

Imported packs are read-only by default. Local overrides can shadow imported capabilities by policy.

## Current MVP

```bash
cd oktopus
go run ./cmd/oktopus version
go run ./cmd/oktopus db init
go run ./cmd/oktopus registry validate
go run ./cmd/oktopus registry list
go run ./cmd/oktopus registry show workflow:sdlc-default
```

## Design Rules

1. Persona = bounded role, leaf worker by default.
2. Workflow = composition layer.
3. Tool = explicit callable action with permissions and outputs.
4. Skill = reusable process with evidence requirements.
5. Harness controller = only component allowed to orchestrate fan-out/fan-in.
6. Agent-created capabilities become proposals, not live capabilities.

## Next Steps

1. Add run creation and workflow expansion.
2. Add worker lease protocol and local worker loop.
3. Add importer for `addyosmani/agent-skills` that expands `skills/*/SKILL.md` and `agents/*.md` into registry entries.
4. Add Pi adapter.
5. Add verifier execution.
6. Add CI/CD + telemetry adapters.
