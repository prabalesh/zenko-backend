-- name: CountUserFriends :one
SELECT COUNT(*) FROM friends 
WHERE (sender_id = $1 OR receiver_id = $1) AND status = 'accepted';

-- name: GetFriendship :one
SELECT * FROM friends 
WHERE (sender_id = $1 AND receiver_id = $2) 
   OR (sender_id = $2 AND receiver_id = $1)
LIMIT 1;

-- name: SendFriendRequest :one
INSERT INTO friends (sender_id, receiver_id, status)
VALUES ($1, $2, 'pending')
RETURNING *;

-- name: UpdateFriendStatus :exec
UPDATE friends
SET status = $3, updated_at = NOW()
WHERE (sender_id = $1 AND receiver_id = $2)
   OR (sender_id = $2 AND receiver_id = $1);

-- name: DeleteFriendship :exec
DELETE FROM friends
WHERE (sender_id = $1 AND receiver_id = $2)
   OR (sender_id = $2 AND receiver_id = $1);

-- name: ListFriendsPaginated :many
SELECT u.id, u.username, u.avatar_url, u.elo, u.wins, u.matches_played, u.online_status, f.status
FROM friends f
JOIN users u ON
  (f.sender_id = $1 AND u.id = f.receiver_id) OR
  (f.receiver_id = $1 AND u.id = f.sender_id)
WHERE f.status = 'accepted' AND (u.id > $2 OR $2 IS NULL)
ORDER BY u.online_status DESC, u.elo DESC, u.id ASC
LIMIT $3;

-- name: GetFriendRequestsList :many
SELECT f.id as request_id, f.sender_id, f.receiver_id, f.status, f.created_at,
       u.id as user_id, u.username, u.avatar_url, u.elo, u.wins, u.matches_played, u.online_status
FROM friends f
JOIN users u ON 
  (f.sender_id = $1 AND u.id = f.receiver_id) OR
  (f.receiver_id = $1 AND u.id = f.sender_id)
WHERE f.status = 'pending' AND (f.sender_id = $1 OR f.receiver_id = $1);
