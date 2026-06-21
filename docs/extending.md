# Extending the Harness

This is the heart of the value proposition. Oktopus is not a fixed agent with a fixed toolset — it is a **governed socket** into which you plug your own and others' capabilities, and connect agents to the systems your organization actually runs on.

## Connect your proprietary world

Off-the-shelf agents reach internal systems only through per-tool hooks — an MCP server here, a plugin there — bolted on, ungoverned, and re-wired for every vendor you use. Oktopus makes that integration **first-class**: connect a system once as a governed capability, and every agent uses it under the same policy and scope.

Proprietary and internal systems plug in through:

- **Adapters** — bridges to internal systems (CI/CD, ticketing, telemetry, data stores).
- **Custom tools** — your CLIs, scripts, services, and containers as first-class callable capabilities.
- **MCP servers** — your proprietary MCP server is a governed entry point for private data and actions.
- **Data & knowledge sources** — internal (vaults, docs repos, vector stores, private APIs) or third-party (external APIs, SaaS data feeds).

The platform defines the open socket; the specific connectors are implementations you (or providers) build — versioned and governed by the registry.

→ Spec: `define-client-server-runtime-config-sources`

## The governed supply chain

Capabilities reach your runs from five sources, all through one registry under one trust-and-scope model:

```text
builtin            platform defaults
upstream_pack      open / community packs       (addyosmani/agent-skills, ponytail)
third-party        marketplace providers         (verified vendors)
org                proprietary / internal         (your private data + tools)
project / run      local customization
```

Unlike a vendor's closed plugin store, **every capability — open, third-party, or proprietary — passes the same gates**: trust tier, signature/provenance, sandbox dry-run, scoped activation, and policy. You consume a third-party tool with the exact controls you apply to your own.

→ Specs: `define-capability-registry-model`, `define-enterprise-admin-marketplace-sessions`

## Capabilities

The building blocks you add and compose — each a versioned, governed registry entry (a manifest with source, trust level, version, and policy):

- **Runtime** — the agent/model engine; off-the-shelf (Claude Code, Codex, shell) or custom.
- **Persona** — a single-role operating profile: role, allowed tools, skills, output contract, limits.
- **Tool / Adapter** — callable actions and integration bridges.
- **Skill** — a reusable procedure a persona follows.
- **Workflow** — composition of personas/jobs into multi-step work.
- **Verifier / Policy / Environment** — evidence checks, code-enforced rules, and execution boundaries.
- **Packs** — `skillpack` and `solution_pack` bundle the above for import and sharing.

You rarely build a runtime: take an off-the-shelf harness and customize it through these primitives — your personas, skills, tools, and policy apply on top of whatever runtime executes, so the harness is shaped by portable patterns instead of per-person config. Imported packs are read-only by default; local definitions can shadow them by policy.

→ Spec: `define-capability-registry-model`. See [Concepts](concepts.md) for the full primitive model.

## Configuration and the scope hierarchy

Who can add, override, or activate what is resolved by a single precedence ladder:

```text
builtin < upstream_pack < org < client < project < workflow < run_override
```

Higher scope can narrow permissions freely; **broadening** permissions (more tools, secrets, network, data access) requires policy evaluation and approval. The same ladder governs both capabilities and configuration.

→ Spec: `define-client-server-runtime-config-sources`

## Secrets, credentials, and data sources

Secrets and API keys are distributed centrally and referenced, never embedded — held by the control plane, scoped per org/team/project, and policy-gated on use (`secret_read`), so access is attributable in the event log and governable against budgets. Data and knowledge sources — internal (vaults, vector stores, private APIs) and third-party (external APIs, SaaS feeds) — attach with access policy and redaction, so sensitive data is governed before any worker sees it.

→ Specs: `define-client-server-runtime-config-sources`, `define-artifacts-verifiers-policy`

## Enterprise governance and the marketplace

In an organization, customization is bounded by control:

- **Capability catalog** — the approved set an org publishes internally.
- **Marketplace** — internal and, eventually, **third-party providers** publish capabilities; consumption is governed by trust and scope.
- **Standardized vs. user-custom modes** — admins can mandate fixed workflows/workspace profiles, or allow user customization within policy.
- **Governance controls** — provenance, signing, approval gates, and audit on every addition and activation.

→ Spec: `define-enterprise-admin-marketplace-sessions`
