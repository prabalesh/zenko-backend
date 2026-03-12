-- name: GetNotificationPreferences :one
SELECT * FROM notification_preferences WHERE user_id = $1 LIMIT 1;

-- name: UpsertNotificationPreferences :one
INSERT INTO notification_preferences (
    user_id, friend_request, friend_accepted, challenge_received, 
    challenge_declined, reengagement, weekly_reset, global_mute
) VALUES (
    $1, 
    COALESCE(sqlc.narg('friend_request'), true),
    COALESCE(sqlc.narg('friend_accepted'), true),
    COALESCE(sqlc.narg('challenge_received'), true),
    COALESCE(sqlc.narg('challenge_declined'), true),
    COALESCE(sqlc.narg('reengagement'), true),
    COALESCE(sqlc.narg('weekly_reset'), true),
    COALESCE(sqlc.narg('global_mute'), false)
)
ON CONFLICT (user_id) DO UPDATE SET
    friend_request = COALESCE(sqlc.narg('friend_request'), notification_preferences.friend_request),
    friend_accepted = COALESCE(sqlc.narg('friend_accepted'), notification_preferences.friend_accepted),
    challenge_received = COALESCE(sqlc.narg('challenge_received'), notification_preferences.challenge_received),
    challenge_declined = COALESCE(sqlc.narg('challenge_declined'), notification_preferences.challenge_declined),
    reengagement = COALESCE(sqlc.narg('reengagement'), notification_preferences.reengagement),
    weekly_reset = COALESCE(sqlc.narg('weekly_reset'), notification_preferences.weekly_reset),
    global_mute = COALESCE(sqlc.narg('global_mute'), notification_preferences.global_mute),
    updated_at = NOW()
RETURNING *;

-- name: RegisterFCMToken :exec
INSERT INTO fcm_tokens (user_id, fcm_token, platform)
VALUES ($1, $2, $3)
ON CONFLICT (fcm_token) DO UPDATE SET
    user_id = EXCLUDED.user_id,
    platform = EXCLUDED.platform,
    updated_at = NOW();

-- name: GetUserFCMTokens :many
SELECT fcm_token FROM fcm_tokens WHERE user_id = $1;

-- name: DeleteFCMToken :exec
DELETE FROM fcm_tokens WHERE fcm_token = $1;

-- name: ListNotificationsPaginated :many
SELECT * FROM notifications 
WHERE user_id = $1 AND (id < $2 OR $2 IS NULL)
ORDER BY created_at DESC
LIMIT $3;

-- name: MarkNotificationRead :exec
UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false;

-- name: InsertNotification :one
INSERT INTO notifications (user_id, type, title, body, data)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
