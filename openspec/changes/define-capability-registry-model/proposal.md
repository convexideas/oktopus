## Why

Oktopus must be armed with many evolving capabilities: tools like Graphify, OpenSpec, Archon, Beads, OpenShell, CI/CD adapters, telemetry adapters, inference evaluators, Agent Skills packs, personas, and client-specific workflows. Hardcoding these integrations would make the platform brittle and prevent organizations from extending it safely.

A versioned capability registry gives Oktopus a stable way to discover, validate, govern, compose, override, and audit capabilities across local and distributed deployments.

## What Changes

- Define the capability registry as a first-class platform subsystem.
- Define capability kinds: `tool`, `skill`, `skillpack`, `persona`, `workflow`, `verifier`, `environment`, `policy`, `runtime`, `solution_pack`, and `adapter`.
- Define manifest requirements, lifecycle states, trust boundaries, override rules, and activation gates.
- Define how imported packs such as `addyosmani/agent-skills` are represented.
- Define how agent-created capabilities are handled as proposals, not live runtime assets.
- Define how solution packs compose capabilities for engineering ops, CI remediation, AIOps, documentation, and client inference harnesses.

## Out of Scope

- Full JSON Schema definitions for every manifest kind.
- Implementing package installation or signature verification.
- Building a marketplace.
- Implementing all listed adapters.
- Selecting final remote registry backend.

## Impact

This change makes Oktopus extensible by design. Future implementation can load local manifests first, then support signed org registries, package sources, marketplace-like distribution, and client-specific solution packs without changing the core execution model.
