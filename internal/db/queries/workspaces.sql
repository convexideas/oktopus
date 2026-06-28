-- name: CreateWorkspace :exec
INSERT INTO workspaces (id, name, created_at) VALUES (?, ?, ?);

-- name: GetWorkspace :one
SELECT * FROM workspaces WHERE id = ?;

-- name: GetWorkspaceByName :one
SELECT * FROM workspaces WHERE name = ?;

-- name: ListWorkspaces :many
SELECT * FROM workspaces ORDER BY name;

-- name: CreateWorkspaceRef :exec
INSERT INTO workspace_refs (workspace_id, kind, slot, ref, created_at)
VALUES (?, ?, ?, ?, ?);

-- name: ListWorkspaceRefs :many
SELECT * FROM workspace_refs WHERE workspace_id = ? ORDER BY kind, slot;

-- name: CreateSandboxDef :exec
INSERT INTO sandbox_defs (id, name, provider, base_image_ref, created_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetSandboxDef :one
SELECT * FROM sandbox_defs WHERE id = ?;

-- name: ListSandboxDefs :many
SELECT * FROM sandbox_defs ORDER BY name;

-- name: CreateSandboxInstance :exec
INSERT INTO sandbox_instances
  (id, sandbox_def_id, workspace_id, owner_subject, execution_principal,
   resource_limits_json, grants_json, environment_json, provider_sandbox_id,
   status, created_at, updated_at, last_seen_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetSandboxInstance :one
SELECT * FROM sandbox_instances WHERE id = ?;

-- name: ListSandboxInstancesByOwner :many
SELECT * FROM sandbox_instances WHERE owner_subject = ? AND status = ? ORDER BY created_at DESC;

-- name: UpdateSandboxInstanceStatus :exec
UPDATE sandbox_instances SET status = ?, updated_at = ?, last_seen_at = ? WHERE id = ?;
