-- name: GetUserByEmail :one
SELECT id, email, display_name, password_hash, created_at
FROM users
WHERE email = ?;

-- name: GetUserByID :one
SELECT id, email, display_name, password_hash, created_at
FROM users
WHERE id = ?;

-- name: CreateUser :one
INSERT INTO users (email, display_name, password_hash)
VALUES (sqlc.arg(email), sqlc.arg(display_name), sqlc.arg(password_hash))
RETURNING id, email, display_name, password_hash, created_at;
