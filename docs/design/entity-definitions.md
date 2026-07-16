# Entity Definitions

Canonical definitions for all Oktopus primitives. Strict boundaries.

## Session

A bounded unit of work between a user (or automation) and one or more agents. Immutable record after completion.

**Owns:** ID, timestamps, sandbox reference, harness(es) invoked, conversation transcript, outcome, cost metrics.

**Does NOT own:** memory (workspace-scoped), policy (sandbox-scoped), agent configuration (persona + profile, materialized before session).

**Lifecycle:** Created on run → appended during execution → sealed on end. Never modified after.

**Flavors:**

| Flavor | Interaction | Trigger |
|--------|-------------|---------|
| Interactive | Human at keyboard | Manual |
| One-shot | Task in, result out | Manual or API |
| Scheduled | Cron/timer | Automated |
| Event-driven | Webhook/event | Automated |
| Shared/live | Multiple humans | Manual (invite) |
| Forked | Cloned from another | Manual |

---

## Workspace

Declarative scope of work. Defines WHAT can be done. Namespace for memory and sandboxes.

**Owns:** ID (UUID, portal-addressable), name, source refs, I/O declarations, connector declarations, memory, sandboxes.

**Does NOT own:** compute (sandbox provider), policy (layered from org → workspace), user identity (profile).

**Lifecycle:** Created once, evolves over time. Durable. Synced between portal and local.

**Flavors:**

| Flavor | Owner | Sharing |
|--------|-------|---------|
| Personal | One user | Private |
| Team | A team | Members access |
| Template | Platform team | Fork to create from |
| Public | Org-wide | View/fork |

---

## Sandbox

Runtime environment where agents execute. The "machine + account." Persistent, re-enterable, governed.

**Owns:** Name (within workspace), HOME directory, provider type, bound policy, harness state (session history, extensions).

**Does NOT own:** project files (workspace source), memory (workspace-level), identity (profile), persona (injected at materialization — configuration, not identity).

**Lifecycle:** Created explicitly → persists → re-entered across sessions → destroyed by policy or user.

**File layout:**
```
~/.ok/<workspace>/sandboxes/<name>/home/
```

**HOME redirected.** Harness sees sandbox as its world. cwd = workspace source (real project path).

**Flavors:**

| Flavor | Lifecycle | Use case |
|--------|-----------|----------|
| Persistent | Until destroyed | Day-to-day dev |
| Ephemeral | Destroyed after session | CI, one-shot, security |
| Shared | Multiple users enter | Pair programming |
| Locked | Read-only after point | Audit freeze |

---

## Profile

User's portable identity. Who they are, what they prefer, what credentials they carry.

**Owns:** User ID, preferences (portable: model, extensions, tool exclusions), harness config (per-harness settings), credentials (referenced).

**Does NOT own:** workspaces (has access, doesn't own definition), policy (applied TO profile by admin), memory (workspace-scoped).

**Lifecycle:** Created on first use/import. Evolves with preferences. Portable across machines.

**Flavors:**

| Flavor | Governed by | Example |
|--------|-------------|---------|
| Personal | Self | Individual developer |
| Managed | Admin | Enterprise employee with org overlays |
| Service | Policy-only | Automation (no human, delegated permissions) |

**Automation under a profile:** Scheduled/triggered sessions run AS the user, constrained by sandbox policy. No separate service account entity needed for personal automations. Policy restricts permissions for automated contexts.

---

## Persona

Agent configuration template. Defines how to set up an agent. Applied at sandbox materialization. Stateless.

**Owns:** System prompt, skills, output format, native config per harness.

**Does NOT own:** permissions (policy), identity (profile), state (stateless template), runtime behavior (agent is free after configuration).

**Lifecycle:** Authored, versioned, registered in capability registry. Immutable per version.

**Usage:** Configures the primary agent in a sandbox. At runtime, the agent may delegate to sub-agents with dynamic instructions — sub-agents don't need pre-defined personas.

**Flavors:**

| Flavor | Scope | Example |
|--------|-------|---------|
| Bundled | Ships with platform | "code-reviewer" |
| User-authored | Personal registry | "my-opinionated-reviewer" |
| Org-curated | Org registry | "company-security-auditor" |
| Dynamic | Runtime (coordinator-generated) | Sub-agent ad-hoc instructions |

---

## Policy

Enforcement rules. Cannot be bypassed by agent.

**Owns:** Rules (filesystem, network, tools, spend), scope, enforcement mode (allow/deny/ask), stacking logic.

**Does NOT own:** configuration (persona + profile), identity (profile), audit trail (session records policy decisions).

**Lifecycle:** Defined by admin or user. Version-controlled. Some rules hot-reloadable (network, spend); some locked at sandbox creation (filesystem, process).

**Stacking:** org → team → workspace → sandbox → session. Stricter wins. Lower levels cannot loosen upper levels.

**Flavors:**

| Flavor | Applied at | Set by |
|--------|-----------|--------|
| Org-wide | Everything | Admin only |
| Workspace-level | All sandboxes within | Workspace owner |
| Sandbox-level | This sandbox | User (within bounds) |
| Session-level | This session | User (most restrictive wins) |

---

## Memory

Organizational intelligence extracted from sessions. Scoped to workspace.

**Owns:** Episodes (raw captures), summaries (LLM-distilled), future: semantic facts, knowledge graph.

**Does NOT own:** session transcripts (memory derived from them), policy.

**Lifecycle:** Accumulates. Episodes immutable. Summaries can be superseded. Propagation controlled by admin.

**Flavors:**

| Flavor | Scope | Propagation |
|--------|-------|-------------|
| Session-local | In-context only | Not persisted separately |
| Workspace-scoped | All sessions in workspace | Default |
| Team-scoped | Across team workspaces | Admin-controlled |
| Org-scoped | Global intelligence | Admin-curated |

---

## Group (Team)

A named collection of users and/or sub-groups. Recursive — groups can contain groups (Engineering → Backend → Payments). The unit for collaboration, policy, and memory scoping.

**Owns:** ID, name, parent group (nullable), membership (users + sub-groups), group-level policy, group-scoped memory, shared workspaces.

**Does NOT own:** individual user preferences (profile), org-level policy.

**Lifecycle:** Created by admin. Users/sub-groups added/removed. Durable.

**Role:** The collaboration/governance unit. Policy cascades down the tree (tightens only). Memory propagates up the tree (admin-controlled). Workspaces can be owned at any level.

**Structure:** Recursive tree under Org.
```
Org
├── Group: Engineering
│   ├── Group: Backend
│   │   ├── Group: Payments
│   │   └── Group: Identity
│   └── Group: Frontend
└── Group: Security
```

---

## Gateway

The enforcement and observation point between the sandbox and the outside world. Where policy, credentials, capture, and routing converge.

**Owns:** ID, configuration (per sandbox or shared), routing rules, credential injection mappings, traffic logs, TLS handling.

**Does NOT own:** credentials themselves (references credential store), policy definitions (enforces policies defined elsewhere), session logic.

**Responsibilities:**

| Concern | What the gateway does |
|---------|----------------------|
| Credential injection | Replaces sentinel tokens with real secrets on outbound requests |
| Network policy | Allows/denies egress to specific destinations |
| Inference routing | Directs model API calls to approved endpoints (region, provider) |
| Structured capture | Logs full API request/response for conversation capture |
| Cost metering | Counts tokens/cost from API response headers |

**Lifecycle:** Configured per sandbox (or shared across sandboxes in a workspace). Starts with sandbox, stops with sandbox. Logs persist for audit.

**Deployment modes:**

| Mode | Where it runs | When |
|------|---------------|------|
| None | No proxy, env passthrough | Phase 1 (now) |
| Local sidecar | Localhost proxy, sandbox routes traffic through it | Phase 2 |
| Host-level | One gateway per machine, all sandboxes route through | Phase 3 |
| Cloud | Managed gateway (Cloudflare Workers, edge proxy) | Portal phase |

**Connection to other entities:**
- Sandbox → routes all egress through its assigned Gateway
- Policy → Gateway enforces network/credential rules
- Credential → Gateway injects real values, logs usage
- Memory → Gateway produces structured conversation captures (highest fidelity)
- Audit → Gateway logs every external interaction

---

## Credential

A reference to a secret (API key, token, certificate) that the platform manages on behalf of the user/org. The agent never sees the real value.

**Owns:** ID, name (human-readable, e.g. "github-token"), scope (org/group/workspace/user), provider type, allowed destinations (which hosts/services this credential is for), rotation policy.

**Does NOT own:** the secret value itself in domain logic (stored in credential store — keychain, vault, env). Never serialized, never in DB, never in sandbox.

**Lifecycle:** Created by user/admin. Bound to scope. Rotatable. Revocable. Audited on every use.

**Provider types:**

| Provider | Where the real secret lives |
|----------|-----------------------------|
| env | Host environment variable (Phase 1 — simplest, least secure) |
| keychain | OS keychain (macOS Keychain, Linux secret-service) |
| vault | External secrets manager (HashiCorp Vault, AWS Secrets Manager, 1Password) |
| oidc | Identity federation — no static secret, short-lived tokens |

**Injection modes:**

| Mode | Security | When |
|------|----------|------|
| env-passthrough | Low — agent sees real value | Phase 1 (now) |
| resolved-at-launch | Medium — injected into sandbox env at start | Phase 2 |
| gateway-injection | High — agent sees sentinel, proxy injects real value on egress | Phase 3 |

**How it connects:**

- Workspace declares: "needs credentials: [github-token, openai-key]"
- Profile/Group/Org provides: "my github-token resolves via keychain entry X"
- Sandbox receives: sentinel env vars (`GITHUB_TOKEN=ok:sentinel:github-token`)
- Gateway rewrites: outbound request to github.com gets real token injected in auth header
- Policy governs: "credential X only allowed for destination Y"
- Audit logs: "credential github-token used by session Z to call github.com/api at timestamp T"

**Policy integration:**

```yaml
policy:
  credentials:
    github-token:
      destinations: [api.github.com]
      deny: ["*"]  # cannot be used for any other host
    openai-key:
      destinations: [api.openai.com]
      max_daily_cost_usd: 100
```

---

## Role

A named bundle of permissions that can be assigned to users or groups. Scoped to a level (org, group, workspace). The bridge between identity and policy.

**Two axes:**

### Platform Roles (what you can do in the system)

| Role | Scope | Grants |
|------|-------|--------|
| org-admin | Org | Full control: policy, groups, workspaces, users, billing |
| group-admin | Group | Manage group's workspaces, members, sub-groups, policy |
| workspace-owner | Workspace | Manage sandboxes, personas, connectors, policy for this workspace |
| member | Workspace/Group | Run sessions, create sandboxes (within policy), view memory |
| viewer | Workspace/Group | Read-only: sessions, memory, config. Cannot execute. |

### Agent Governance Roles (how agents behave under this identity)

| Role | Meaning |
|------|---------|
| full-auto | Agent executes without human approval (within policy bounds) |
| supervised | Agent pauses for approval on actions classified as risky |
| restricted | Agent can only suggest/draft. Never executes tools. |
| audited | Like full-auto, but every action logged with elevated detail |

**Owns:** ID, name, description, scope level (org/group/workspace), permission set, governance level.

**Does NOT own:** users (users are assigned TO roles), policy rules (policy references roles, not vice versa).

**Assignment:**
- User → Role (scoped): "Priya is member of payments-api workspace"
- Group → Role (scoped): "Backend group has workspace-owner on all backend workspaces"
- Inherited: sub-group members inherit parent group roles (can be tightened, not loosened)

**Policy references roles:**
```yaml
policy:
  rules:
    - when: role = "member"
      deny: sandbox.create(type=ephemeral)
    - when: role = "viewer"
      deny: session.create
    - when: governance = "supervised"
      require_approval: [shell, file_write, git_push]
    - when: governance = "restricted"
      deny: [shell, file_write, git_push, network_request]
```

**Lifecycle:** Defined by admin. Assigned to users/groups with scope. Revocable. Platform comes with built-in defaults. Custom roles creatable by org-admin.

**Hierarchy resolution:** When a user has multiple roles (via group membership at different levels), the most specific scope wins. For governance, the most restrictive wins.

---

## Org

The top-level entity. The enterprise boundary. Where IP protection and global policy live.

**Owns:** Teams, org-level policy, org-scoped memory (global consciousness), billing, admin controls, capability registry (approved personas/tools).

**Does NOT own:** individual team operations, user preferences.

**Lifecycle:** Created once. Durable. The root of all trust and policy.

**Role:** Sets the floor for governance. Org policy cannot be loosened by anyone below. Org memory is curated intelligence that flows down to all teams. IP boundary — what stays inside.

**Flavors:**

| Flavor | Scale | Example |
|--------|-------|---------|
| Solo | One user, no teams | Personal use, local-only |
| Startup | Flat, everyone on one team | Small company |
| Enterprise | Multiple teams, admin hierarchy | Large company with departments |

---

## Relationships

```
Org ──contains──→ Group(s) (recursive tree)
Org ──defines──→ Role(s) (platform + governance)
Org ──stores──→ Credential(s) (org-scoped secrets)
Group ──contains──→ Group(s) (sub-groups)
Group ──contains──→ User(s) (via membership)
Group ──owns──→ Workspace(s)
Group ──stores──→ Credential(s) (group-scoped secrets)
User/Group ──assigned──→ Role (scoped to org/group/workspace)
Role ──referenced-by──→ Policy (rules target by role)
Profile ──uses──→ Workspace
Profile ──binds──→ Credential(s) (personal secrets)
Workspace ──declares──→ Credential requirements (what secrets it needs)
Workspace ──contains──→ Sandbox(es)
Workspace ──contains──→ Memory
Sandbox ──configured-by──→ Persona (at materialization)
Sandbox ──governed-by──→ Policy (layered)
Sandbox ──receives──→ Credential sentinels (not real values)
Session ──runs-in──→ Sandbox
Session ──uses-identity──→ Profile (with role context)
Session ──produces──→ Memory episodes
Session ──audits──→ Credential usage
Policy ──cascades-down──→ org > group > ... > group > workspace > sandbox > session
Policy ──governs──→ Credential destinations + limits
Memory ──propagates-up──→ workspace > group > ... > group > org (admin-controlled)
Gateway ──enforces──→ Policy (network rules, credential destinations)
Gateway ──injects──→ real Credential values on egress (agent never sees them)
Gateway ──produces──→ structured conversation capture (highest fidelity)
Gateway ──meters──→ cost/tokens per Session
```

## Access Control (portal phase)

For sharing to work:
- Workspace must be accessible to all participants
- Sandbox must allow shared entry (explicit grant)
- Session is the shareable unit (link gives access)
- Memory stays workspace-scoped (shared sessions contribute to shared memory)
- Policy still applies per-participant (different spend caps in same session)

For local-first phase: everything implicitly personal. Model doesn't prevent sharing from being added.
