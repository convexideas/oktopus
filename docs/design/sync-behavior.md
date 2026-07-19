# Sync Behavior Design

## Core Model

The portal gateway is the only path for cloud LLMs. If you can reach
Anthropic/OpenAI, you can reach the portal — they require the same
connectivity. Data lands server-side in real-time. No sync needed.

```
Cloud LLM:   agent → portal gateway → captures + memory live server-side
Local LLM:   agent → local gateway → local SQLite → push when connected
```

The local gateway exists solely for local models (Ollama, vLLM).
This is niche — most sessions use cloud LLMs and need no sync.

## Symmetric Profiles

Personal and org-managed profiles use the same architecture:

| | Self-managed | Org-managed |
|--|--|--|
| Gateway | Portal | Portal |
| API keys | User provides | Org provides |
| Memory scope | User's workspaces | Org workspaces |
| Cost | User's bill | Org's bill |
| Visibility | Only the user | Team/org |
| Policy | User-defined | Admin-defined |

## Session Tagging

Every session carries metadata for the portal to file it, regardless
of where it ran:

```yaml
session:
  user: me@acme.com
  workspace: payments
  sandbox: { type: seatbelt, host: saurabh-macbook }
  agent: pi
  model: claude-sonnet-4-20250514
  cost_usd: 0.34
  tokens: { in: 12400, out: 3200 }
  status: completed
```

## Sync (local-model sessions only)

When a session ran against a local model and the user wants it
visible on the portal:

```
ok sync push --workspace name
  → POST /sessions/batch (metadata)
  → POST /memory/batch (extracted knowledge)
```

Pull (org knowledge + policy flowing down):
```
ok sync pull --workspace name
  → GET /memory?workspace=X&since=last_pull
  → GET /policy?workspace=X&since=last_pull
```

No bidirectional conflict. Sessions/memory flow up. Policy/knowledge flows down.

## Multi-Org

Single `portals.json` keyed by portal name. Each workspace affiliates
with one portal via `portal:` field in workspace.yaml.
Workspaces without `portal:` are purely local.
