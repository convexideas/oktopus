-- name: GetCapability :one
SELECT * FROM capabilities WHERE id = ?;

-- name: ListCapabilities :many
SELECT * FROM capabilities WHERE status = ? ORDER BY kind, name;

-- name: ListCapabilitiesByKind :many
SELECT * FROM capabilities WHERE kind = ? AND status = ? ORDER BY name;

-- name: FindCapability :one
SELECT * FROM capabilities WHERE kind = ? AND name = ? AND version = ? LIMIT 1;

-- name: UpsertCapability :exec
INSERT INTO capabilities
  (id, kind, name, version, description, author, source_type, source_uri, source_ref,
   trust_level, scope, status, requirements_json, manifest_path, manifest_json, hash, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  description       = excluded.description,
  author            = excluded.author,
  source_type       = excluded.source_type,
  source_uri        = excluded.source_uri,
  source_ref        = excluded.source_ref,
  trust_level       = excluded.trust_level,
  status            = excluded.status,
  requirements_json = excluded.requirements_json,
  manifest_path     = excluded.manifest_path,
  manifest_json     = excluded.manifest_json,
  hash              = excluded.hash,
  updated_at        = excluded.updated_at;

-- name: DeleteCapability :exec
DELETE FROM capabilities WHERE id = ?;
