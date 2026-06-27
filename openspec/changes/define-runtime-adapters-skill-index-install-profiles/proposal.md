## Why

Gentle-AI demonstrates useful implementation patterns for a multi-agent ecosystem: explicit runtime adapter contracts, capability matrices, exact-path skill registries, delegation triggers, staged install/rollback pipelines, and preset-based capability bundles. Oktopus needs equivalent concepts, but at the control-plane level rather than as a local agent configurator.

Oktopus should learn from those patterns while preserving its own boundary: the server remains source of truth for registry, run state, policy, leases, artifacts, approvals, and audit. Agent runtimes remain interchangeable workers that translate Oktopus jobs into native Pi, Claude Code, Codex, Kiro, OpenCode, shell, or other execution mechanisms.

## What Changes

- Define a runtime adapter contract for background agent workers.
- Define a runtime capability matrix for administrators and workflow selection.
- Define a skill source index that preserves exact skill paths and source provenance.
- Define delegation policy rules that decide when work should split into exploration, writer, review, verifier, or synthesis jobs.
- Define a staged capability install/activation pipeline with plan, snapshot, apply, verify, rollback, and audit phases.
- Define presets/profiles for marketplace and solution-pack installation.

## Out of Scope

- Implementing every runtime adapter.
- Mutating user-level agent configuration automatically.
- Replacing Oktopus workflow/job state with runtime-native task state.
- Building marketplace UI.
- Implementing rollback for every external system.

## Impact

This change makes Oktopus extension mechanics more concrete. It clarifies how configurable skills, agents, workflows, and packages become usable without losing source provenance, policy enforcement, or operational safety.
