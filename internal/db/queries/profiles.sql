-- name: CreateProfile :exec
INSERT INTO profiles (id, name, created_at, updated_at) VALUES (?, ?, ?, ?);

-- name: GetProfile :one
SELECT * FROM profiles WHERE id = ?;

-- name: GetProfileByName :one
SELECT * FROM profiles WHERE name = ?;

-- name: ListProfiles :many
SELECT * FROM profiles ORDER BY name;

-- name: UpdateProfileName :exec
UPDATE profiles SET name = ?, updated_at = ? WHERE id = ?;

-- name: DeleteProfile :exec
DELETE FROM profiles WHERE id = ?;
