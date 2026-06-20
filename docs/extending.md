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

The first-class, versioned building blocks you add and compose:

`tool` · `skill` · `skillpack` · `persona` · `workflow` · `verifier` · `environment` · `policy` · `runtime` · `adapter` · `solution_pack`

Each carries a manifest with source, trust level, version, and policy. Imported packs are read-only by default; local definitions can shadow them by policy.

→ Spec: `define-capability-registry-model`

## Configuration and the scope hierarchy

Who can add, override, or activate what is resolved by a single precedence ladder:

```text
builtin < upstream_pack < org < client < project < workflow < run_override
```

Higher scope can narrow permissions freely; **broadening** permissions (more tools, secrets, network, data access) requires policy evaluation and approval. The same ladder governs both capabilities and configuration.

→ Spec: `define-client-server-runtime-config-sources`

## Secrets and data sources

Secrets are referenced, never embedded — workspace profiles carry secret references, and `secret_read` is a policy-gated action. Data and knowledge sources — internal (vaults, vector stores, private APIs) and third-party (external APIs, SaaS feeds) — attach with access policy and redaction, so sensitive data is governed before any worker sees it.

→ Specs: `define-client-server-runtime-config-sources`, `define-artifacts-verifiers-policy`

## Enterprise governance and the marketplace

In an organization, customization is bounded by control:

- **Capability catalog** — the approved set an org publishes internally.
- **Marketplace** — internal and **third-party providers** publish capabilities; consumption is governed by trust and scope.
- **Standardized vs. user-custom modes** — admins can mandate fixed workflows/workspace profiles, or allow user customization within policy.
- **Governance controls** — provenance, signing, approval gates, and audit on every addition and activation.

→ Spec: `define-enterprise-admin-marketplace-sessions`
