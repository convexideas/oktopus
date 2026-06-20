## Why

Oktopus is intended to be a configurable client/server agentic operations framework, not only a local CLI or one coding-agent wrapper. Users should be able to trigger tasks through messages, imperative commands, API calls, CI/CD events, telemetry alerts, and web UI actions. Background agents such as Pi, Claude Code, Codex CLI, Antigravity CLI, OpenCode, Kiro, Aider, Goose, shell workers, tool workers, verifier workers, and inference workers should execute bounded jobs according to server-side configuration and policy.

The platform also needs configurable sources for skills, tools, personas, workflows, runtimes, adapters, and knowledge bases such as vaults, documentation repositories, vector stores, and external APIs. These sources must be governable, versioned, policy-controlled, and scoped per organization, client, project, and run.

## What Changes

- Define Oktopus as a client/server system with server-owned control plane and background worker agents.
- Define input surfaces for message-driven and imperative command-driven task creation.
- Define configurable source registries for skills, tools, personas, workflows, runtimes, adapters, verifiers, policies, and knowledge bases.
- Define knowledge source and vault integration hooks.
- Define configuration hierarchy from built-in defaults through org, client, project, and run overrides.
- Define router behavior for converting messages or commands into workflow runs.
- Define context scoping before handing work to workers.
- Clarify that active workflows are selected by server policy, user command, or configured triggers; workers only execute bounded jobs.

## Out of Scope

- Implementing the HTTP API, web UI, or message gateway.
- Implementing specific connectors such as Obsidian, Confluence, Qdrant, GitHub, Slack, or PagerDuty.
- Implementing authentication/RBAC beyond acknowledging future need.
- Choosing final config file format beyond YAML-first manifests.
- Building agent runtime adapters beyond registering them as worker kinds/capabilities.

## Impact

This change updates the platform direction before deeper MVP work. Local CLI remains useful, but it is now clearly the first client of a broader Oktopus server. Background agents become workers that register capabilities and lease jobs. Skill/tool/knowledge sources become configurable inputs to the registry and context system rather than hardcoded code paths.
