-- name: CreateSession :exec
INSERT INTO sessions (id, sandbox_instance_id, title, status, created_by, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetSession :one
SELECT * FROM sessions WHERE id = ?;

-- name: ListSessionsBySandbox :many
SELECT * FROM sessions WHERE sandbox_instance_id = ? ORDER BY created_at DESC;

-- name: UpdateSessionStatus :exec
UPDATE sessions SET status = ?, updated_at = ? WHERE id = ?;

-- name: CreateMessage :exec
INSERT INTO messages (id, session_id, parent_id, role, sequence, created_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListMessages :many
SELECT * FROM messages WHERE session_id = ? ORDER BY sequence;

-- name: GetMessage :one
SELECT * FROM messages WHERE id = ?;

-- name: CreateMessagePart :exec
INSERT INTO message_parts (id, message_id, sequence, type, content_json)
VALUES (?, ?, ?, ?, ?);

-- name: ListMessageParts :many
SELECT * FROM message_parts WHERE message_id = ? ORDER BY sequence;

-- name: CreateMessageAttachment :exec
INSERT INTO message_attachments (id, message_id, direction, name, uri, media_type, size_bytes, sha256, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListMessageAttachments :many
SELECT * FROM message_attachments WHERE message_id = ?;
