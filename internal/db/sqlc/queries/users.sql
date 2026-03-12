-- name: GetUserByGoogleID :one
SELECT * FROM users WHERE google_id = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (google_id, username, avatar_url)
VALUES ($1, $2, $3)
RETURNING *;
