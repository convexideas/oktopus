-- name: CreateUser :exec
INSERT INTO users (id, name, email, external_id, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: GetUserByExternalID :one
SELECT * FROM users WHERE external_id = ?;

-- name: ListUsers :many
SELECT * FROM users ORDER BY name;

-- name: UpdateUser :exec
UPDATE users SET name = ?, email = ?, status = ?, updated_at = ? WHERE id = ?;
