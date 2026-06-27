## ADDED Requirements

### Requirement: Local Coding Session MVP

Oktopus SHALL first provide a local coding-session manager, not a general workflow engine.

#### Scenario: Start a local coding session

- **WHEN** a developer runs `oktopus session start <title> --workspace <workspace>`
- **THEN** Oktopus creates a durable session bound to that workspace
- **AND** records the session title, workspace reference, status, timestamps, and local profile context

#### Scenario: Inspect local coding sessions

- **WHEN** a developer lists or shows sessions
- **THEN** Oktopus displays the session, workspace, status, timeline, and artifact references from durable state

### Requirement: Profiles and Workspaces

Oktopus SHALL model local execution through profiles and workspaces.

#### Scenario: Initialize local profile

- **WHEN** a developer runs `oktopus profile init local`
- **THEN** Oktopus creates or records a local profile with registry path, database URL, runs/artifacts directory, and sandbox defaults

#### Scenario: Create workspace from local path

- **WHEN** a developer runs `oktopus workspace create <name> --path <path>`
- **THEN** Oktopus records a workspace pointing at the local coding directory
- **AND** later sandbox startup can copy, mount, or upload that workspace into an isolated runtime

### Requirement: Workspace Materialization

Oktopus SHALL materialize a workspace into a sandbox using the selected profile. Materialization is an implementation step, not a first-class user object.

#### Scenario: Prepare workspace for sandbox

- **WHEN** a developer starts or attaches a coding session that needs a sandbox
- **THEN** Oktopus prepares workspace files, profile metadata, session metadata, generated agent instructions, memory summary, runtime launch metadata, event log path, and artifact directory for the sandbox provider
- **AND** the prepared layout can be copied, mounted, uploaded, or image-backed depending on provider support

#### Scenario: Keep secrets out of materialized workspace

- **WHEN** Oktopus prepares workspace context for a sandbox
- **THEN** it stores secret references and policy intent only
- **AND** raw API keys, login tokens, and provider credentials are not written into workspace files or staging directories

### Requirement: OpenShell as Local Sandbox Provider

Oktopus SHALL treat OpenShell as the first local sandbox provider for interactive agent sessions.

#### Scenario: Run session in sandbox

- **WHEN** a session is run through the OpenShell provider
- **THEN** Oktopus creates or selects an OpenShell sandbox from the workspace and selected profile
- **AND** records the OpenShell sandbox identifier on the session or workspace

#### Scenario: Attach to session

- **WHEN** a developer runs `oktopus session attach <session>`
- **THEN** Oktopus attaches to the managed agent process or sandbox terminal using the provider's connect/exec/TTY mechanism
- **AND** user input and agent output can be mirrored into session events or logs

### Requirement: Local State Store

Oktopus SHALL use the storage boundary for local durable state.

#### Scenario: Initialize local database

- **WHEN** a developer runs `oktopus db init`
- **THEN** Oktopus creates the configured local state store if missing
- **AND** applies all pending migrations exactly once

#### Scenario: Apply migrations safely

- **WHEN** migrations are run multiple times
- **THEN** Oktopus applies each migration once
- **AND** records applied migration versions in `schema_migrations`

### Requirement: Local Session Events and Artifacts

Oktopus SHALL persist session timeline events and artifact metadata outside any agent runtime.

#### Scenario: Record session event

- **WHEN** a session is created, prepared, attached, run, paused, resumed, or closed
- **THEN** Oktopus records an append-only session event with timestamp, type, actor, and payload metadata

#### Scenario: Record session artifact

- **WHEN** an agent or developer adds an artifact to a session
- **THEN** Oktopus records artifact metadata including URI, type, size, digest, producer, and creation time

### Requirement: Installed CLI is on PATH

Oktopus SHALL be invokable as `oktopus` after supported installation.

#### Scenario: Installed CLI is on PATH

- **WHEN** a developer installs Oktopus through a supported installer or package manager
- **THEN** the installer makes the `oktopus` command available on `PATH`
- **AND** the developer does not need to invoke the binary by full filesystem path

### Requirement: Deferred Workflow Engine

Oktopus SHALL defer general workflow/run/job execution until the coding-session MVP is working.

#### Scenario: Keep run/job schema as future foundation

- **WHEN** existing run/job/worker tables or specs are present
- **THEN** they MAY remain as future foundation
- **BUT** the local MVP user flow is profiles, workspaces, sessions, sandbox attach, events, and artifacts
