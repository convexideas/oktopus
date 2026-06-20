## Why

Oktopus should evolve into an enterprise-grade agent operations platform, not only a coding harness. Administrators need to define and govern standard personas, workflows, skills, MCP servers, tools, commands, data sources, secrets, models, runtimes, policies, approvals, schedules, output destinations, and solution packs for repeatable organizational tasks. Users also need room to customize their own agents and sessions when they are not operating inside a locked standardized workflow.

Enterprise usage also requires resumable long-running threads. Each thread should behave like a session container: it has identity, memory, state, artifacts, permissions, connected inputs, output destinations, and resume capability. For interactive work, users may need preconfigured VMs/workspaces with all tools, data connectors, runtimes, credentials, and policies already installed, so they can run rich agent sessions without local setup drift.

## What Changes

- Define enterprise administrator-managed capabilities and marketplace distribution.
- Define user-customizable agents and personal/team workspaces outside locked standardized workflows.
- Define standardized workflow enforcement for governed tasks.
- Define threads/session capsules as resumable execution and conversation containers.
- Define one-shot, scheduled, event-triggered, and long-running interactive run modes.
- Define input channels and output destinations beyond stdin/stdout.
- Define preconfigured interactive VM/workspace environments.
- Define enterprise governance: RBAC/ABAC, secrets, audit, retention, data boundaries, budgets, approvals, capability provenance, and marketplace lifecycle.

## Out of Scope

- Implementing marketplace UI.
- Implementing VM provisioning or Kubernetes integration.
- Implementing concrete Slack/Teams/Email/GitHub/Confluence connectors.
- Implementing full RBAC/ABAC policy language.
- Selecting final VM provider or sandbox technology.

## Impact

This change clarifies the enterprise product direction. Local MVP remains the first implementation path, but the architecture must preserve room for governed marketplaces, user-customized agents, resumable threads, configurable input/output surfaces, and preconfigured interactive workspaces.
