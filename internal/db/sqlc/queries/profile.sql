-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET 
  username = COALESCE(sqlc.narg('username'), username),
  bio = COALESCE(sqlc.narg('bio'), bio),
  country = COALESCE(sqlc.narg('country'), country),
  dob = COALESCE(sqlc.narg('dob'), dob)
WHERE id = $1
RETURNING *;

-- name: GetSocialLinksByUserID :many
SELECT * FROM user_social_links WHERE user_id = $1;

-- name: UpsertSocialLink :one
INSERT INTO user_social_links (user_id, platform, url)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, platform) 
DO UPDATE SET url = EXCLUDED.url, updated_at = NOW()
RETURNING *;

-- name: DeleteSocialLink :exec
DELETE FROM user_social_links WHERE user_id = $1 AND platform = $2;

-- name: CountUsernameChangesPast30Days :one
SELECT COUNT(*) FROM username_changes 
WHERE user_id = $1 AND changed_at > NOW() - INTERVAL '30 days';

-- name: GetLatestUsernameChange :one
SELECT changed_at FROM username_changes
WHERE user_id = $1
ORDER BY changed_at DESC
LIMIT 1;

-- name: InsertUsernameChange :exec
INSERT INTO username_changes (user_id, old_username, new_username)
VALUES ($1, $2, $3);
