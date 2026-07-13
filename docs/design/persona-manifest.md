# Persona Manifest Schema

A persona is a portable, single-role operating profile: role + perspective +
behavior + output format. It answers "who is this agent?" — not "how is it
deployed." Deployment config (runtime, tools, constraints) lives in the run
command or workflow definition.

Inspired by [addyosmani/agent-skills](https://github.com/addyosmani/agent-skills):
personas do not call other personas; composition belongs to workflows or the user.

## Manifest format

Personas are YAML files. They map to the `capabilities` table with `kind = "persona"`.

```yaml
kind: persona
name: code-reviewer            # unique, kebab-case
version: "0.1.0"               # semver
description: >
  Senior code reviewer that evaluates changes across five
  dimensions: correctness, readability, architecture, security, performance.

# Optional metadata
author: saurabh
tags: [review, quality]

# The persona definition (stored in manifest_json)
persona:
  # --- Generic fields (portable across harnesses) ---

  # The role definition. This IS the system prompt body.
  system_prompt: |
    You are an experienced Staff Engineer conducting a thorough
    code review. Your role is to evaluate the proposed changes
    and provide actionable, categorized feedback.
    ...

  # How to inject the system prompt into the harness.
  # "append" (default): added to harness's native prompt.
  # "replace": replaces the harness's native prompt entirely.
  prompt_mode: append

  # Skills this persona follows (refs to other capabilities).
  skills:
    - code-review-and-quality
    - tdd-workflow

  # Expected output structure (informational, not enforced yet).
  output_format: |
    ## Review Summary
    **Verdict:** APPROVE | REQUEST CHANGES
    ...

  # --- Native overrides (harness-specific, optional) ---

  native:
    # Pi-specific config
    pi:
      # Files to materialize into .pi/ before launch
      files:
        "SYSTEM.md": |
          Custom system prompt for Pi specifically...
        "skills/review/SKILL.md": |
          ---
          name: review
          description: Structured code review skill
          ---
          Steps: ...
        "extensions/gate.ts": registry://extensions/approval-gate
      # CLI flags to pass
      flags:
        - "--tools"
        - "read,grep,find,ls"
        - "--thinking"
        - "high"
      # Settings overrides (merged into .pi/settings.json)
      settings:
        thinkingLevel: high

    # Codex-specific config
    codex:
      # Files to materialize
      files:
        "AGENTS.md": |
          Additional instructions for this persona...
        "skills/review.md": |
          ---
          description: Structured code review
          ---
          Review framework content here...
      # CLI flags
      flags: []
      # Config overrides
      config: {}

    # Kiro-specific config
    kiro:
      files:
        "steering/review.md": |
          ---
          title: Code Review Persona
          inclusion: always
          ---
          You are a senior code reviewer...
```

## Field reference

### Top-level (capability metadata)

| Field | Required | Description |
|-------|----------|-------------|
| `kind` | yes | Always `"persona"` |
| `name` | yes | Unique identifier, kebab-case |
| `version` | yes | Semver string |
| `description` | yes | One-line summary |
| `author` | no | Who created this |
| `tags` | no | Categorization labels |

### persona.system_prompt

The core identity. This text defines the role, perspective, rules, and output
format. It's the equivalent of the markdown body in an agent-skills persona file.

When `prompt_mode` is `append` (default), this is *added to* the harness's
native system prompt. When `replace`, it *becomes* the system prompt.

### persona.skills

References to capabilities of `kind: skill` in the registry. At resolution time,
skill contents are either:
- Appended to the system prompt (simple injection), or
- Materialized as native skill files (e.g., `.pi/skills/`, `.codex/skills/`)

### persona.output_format

Documents the expected output shape. Informational for Phase 2. In later phases,
post-session verifiers can check conformance.

### persona.native

Harness-specific configuration that takes precedence over generic fields.
Structure is per-harness:

```
native:
  <harness-name>:
    files: map[relative_path → content_or_registry_ref]
    flags: []string (CLI args appended to launch command)
    settings: map (merged into harness-specific config)
    config: map (harness-specific config overrides)
```

**files** are materialized into the workspace (or user config dir) before
the harness launches. Paths are relative to the project's harness config
directory (`.pi/`, `.codex/`, `.kiro/`).

**flags** are appended to the harness CLI invocation.

**settings/config** are merged into the harness's JSON/YAML config files.

## Resolution at runtime

When `ok run --persona code-reviewer` executes:

```
1. Load persona from capabilities table (kind=persona, name=code-reviewer)
2. Determine target harness:
   a. If --runtime flag given → use that
   b. If persona has native config for only one harness → use that
   c. If multiple native configs → error, require --runtime flag
   d. If no native configs → use default runtime from user profile
3. Apply generic fields:
   a. system_prompt → translate to harness's injection mechanism
   b. skills → resolve from registry, inject as system prompt or native files
4. Apply native overrides (if present for this harness):
   a. Materialize files into workspace
   b. Append flags to launch args
   c. Merge settings/config
5. Build harness.Config and launch
```

## Harness injection mechanisms

| Harness | System prompt injection | Tool control | Native config dir |
|---------|------------------------|--------------|-------------------|
| **Pi** | `--system-prompt` (replace) or `--append-system-prompt` (append). Also: `.pi/SYSTEM.md` or `.pi/APPEND_SYSTEM.md` | `--tools`, `--exclude-tools` | `.pi/` |
| **Codex** | `AGENTS.md` in project root or `~/.codex/` | Permission profiles, sandboxing config | `.codex/`, `AGENTS.md` |
| **Kiro** | `.kiro/steering/*.md` files | Via steering rules | `.kiro/` |

## Separation of concerns

```
┌─ Persona (the "who") ─────────────────────────────────┐
│ Portable. Single role. No runtime binding.             │
│ Same persona works on Pi, Codex, or Kiro.             │
│                                                        │
│ • system_prompt (the role definition)                  │
│ • skills (reusable procedures it follows)             │
│ • output_format (what it produces)                    │
│ • native overrides (optional per-harness tuning)      │
└────────────────────────────────────────────────────────┘

┌─ Run config (the "how") ──────────────────────────────┐
│ Specified at invocation time. Not stored in persona.  │
│                                                        │
│ • --runtime (which harness)                           │
│ • --workspace (which project/data)                    │
│ • --timeout, --max-cost (constraints)                 │
│ • --sandbox (execution environment)                   │
│ • Memory scopes to query                             │
└────────────────────────────────────────────────────────┘

┌─ Workflow (the "when") ───────────────────────────────┐
│ Composes multiple personas with per-step configs.     │
│                                                        │
│ • steps: [{persona, runtime, tools, constraints}]     │
│ • Fan-out, sequencing, artifact passing               │
└────────────────────────────────────────────────────────┘
```

## Mapping to capabilities table

| Manifest field | DB column |
|----------------|-----------|
| kind | kind |
| name | name |
| version | version |
| description | description |
| author | author |
| (file path) | manifest_path |
| persona.* (entire block as JSON) | manifest_json |
| (sha256 of manifest) | hash |
| "local" or "registry" | source_type |
| "user" / "team" / "org" | scope |
| "active" | status |
| persona.skills (list) | requirements_json |

## Examples

See `registry/personas/` for working examples.

## Compatibility with agent-skills

An `agent-skills` persona markdown file can be imported directly:
- The frontmatter `name` + `description` map to our top-level fields
- The markdown body becomes `persona.system_prompt`
- `kind: persona`, `version: "0.1.0"` are defaulted
- No native overrides (pure generic persona)

Import via: `ok capabilities add --from-agent-skills <path-to-md>`
