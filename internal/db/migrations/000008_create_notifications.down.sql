DROP INDEX IF EXISTS idx_fcm_tokens_user_id;
DROP TABLE IF EXISTS fcm_tokens;

DROP TRIGGER IF EXISTS set_notification_preferences_updated_at ON notification_preferences;
DROP TABLE IF EXISTS notification_preferences;

DROP INDEX IF EXISTS idx_notifications_user_id;
DROP TABLE IF EXISTS notifications;
