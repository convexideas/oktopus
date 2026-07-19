# Sync Behavior Design

## Core Principles

1. **Portal is source of truth for org-scoped entities.** Workspaces, policy, org knowledge — created and managed in portal.
2. **Local is source of truth for execution artifacts.** Sessions, episodes, captures — born locally, pushed up.
3. **Credentials never sync.** Local keychain stays local. Portal-managed sandboxes get keys from the portal, not from users.
4. **Sync is workspace-scoped.** Each workspace belongs to at most one portal (org). No global sync.
5. **Admin preferences are immutable locally.** Pulled, enforced, never overwritten.

## Multi-Org Membership

A user can belong to multiple orgs. Each is a named portal connection:

```yaml
# ~/.ok/config.yaml
portals:
  work:
    server: https://portal.acme.com
    credential: portal-acme
  personal:
    server: https://portal.oktopus.dev
    credential: portal-personal

defaults:
  portal: work
```

A workspace declares its portal affiliation:

```yaml
# workspace.yaml
name: payments
portal: work
```

A workspace without `portal:` is purely local — never syncs.

## What Syncs (by entity)

### Workspaces

| Direction | Behavior |
|-----------|----------|
| Portal → Local | User with permission creates workspace in portal. `ok sync pull` brings it locally. |
| Local → Portal | User creates workspace locally with `portal: work`. `ok sync push` registers it in portal (if user has create permission). |
| Local-only | No `portal` field — never syncs. Personal experiments, throwaway work. |

Admin can lock workspace config fields. Locked fields pulled from portal cannot be overridden locally.

### Sandboxes

| What | Syncs? |
|------|--------|
| Sandbox config/defaults (in workspace.yaml) | Yes — part of workspace config sync |
| Sandbox instances (directories, running state) | Never — local ephemeral state |
| Sandbox policy ("must use container") | Portal → local (admin enforced) |

Local sandboxes don't appear in any global inventory. They're as ephemeral as a shell session.

### Sessions

| Direction | Behavior |
|-----------|----------|
| Local → Portal | Completed/failed sessions push up for team visibility + audit |
| Running sessions | Never sync — only finalized state pushes |
| Portal → Local | Session annotations (tags, reviews from teammates) pull down |

### Memory

| Entity | Direction | Conflict? |
|--------|-----------|-----------|
| Episodes (raw captures) | Local → Portal | No — append-only |
| Knowledge (user-scoped) | Local → Portal | No — user owns their scope |
| Knowledge (org-scoped) | Portal → Local | No — portal is truth |
| Knowledge (team-scoped) | Portal → Local | No — portal is truth |

Propagation direction: org → team → user (read), user → team → org (propose + admin approve).

### Policy

| Direction | Behavior |
|-----------|----------|
| Portal → Local | Always. Admin policies are pulled and enforced. |
| Local → Portal | Never. Users cannot push policy up. |
| Local override | Impossible. Locked policy cannot be relaxed locally. |

### Credentials

**Never sync. Period.**

- Local API keys: user's keychain, used by local gateway.
- Portal-managed sandbox keys: portal injects them into cloud sandboxes, never exposed to user.
- Portal auth token: stored in keychain under `ok:portal-<name>`, used for sync HTTP calls.

## Sync State Tracking

Each syncable record gets a `sync_status` column:

```
local_only    — created locally, not yet pushed (or no portal affiliation)
pushed        — pushed to portal, in sync
pull_pending  — exists in portal, not yet pulled locally
conflict      — both sides changed (workspace config only, rare)
```

Timestamps:
- `last_pushed_at` — when this record was last sent to portal
- `last_pulled_at` — when we last fetched from portal

Sync algorithm:
```
Push: SELECT * FROM sessions WHERE workspace IN (synced workspaces) AND sync_status = 'local_only' AND status IN ('completed', 'failed')
Pull: GET /api/v1/sync/changes?workspace=<id>&since=<last_pulled_at>
```

## Conflict Resolution

True conflicts only occur on workspace config (both sides edited):
- **Field-level merge** for non-overlapping changes (local changed model, portal changed policy → merge both)
- **Portal wins** for overlapping changes on admin-locked fields
- **Local wins** for overlapping changes on user-owned fields (with a warning)
- User is notified: `⚠ workspace 'payments' has remote changes — review with 'ok sync diff'`

## Permission Model

| Action | Required Permission |
|--------|-------------------|
| Create workspace in portal | `workspace:create` on the org |
| Push sessions | `workspace:write` on the workspace |
| Pull workspace | `workspace:read` on the workspace |
| Modify policy | `policy:admin` on the org/workspace |
| Push knowledge | `knowledge:write` on the scope |

Users without `workspace:create` can only pull workspaces created by others.

## CLI Commands

```
ok login [--portal name] [server-url]   — authenticate with a portal
ok logout [--portal name]               — disconnect
ok whoami                                — show auth state for all portals

ok sync status                           — what's pending (per workspace)
ok sync push [--workspace name]          — push local → remote
ok sync pull [--workspace name]          — pull remote → local
ok sync diff [--workspace name]          — show pending changes without applying
```

## Running Locally Without Sync

A user who:
- Creates a workspace without `portal:` field
- Has their own API keys in keychain
- Runs `ok run`

Gets full functionality: enforced sandbox, gateway capture, session tracking, memory extraction. Nothing ever leaves their machine. The portal adds: team visibility, org policy, knowledge propagation, audit. But it's not required.
