ALTER TABLE notification_preferences
  ADD COLUMN friend_requests BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN challenges BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN marketing BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE notification_preferences
  DROP COLUMN friend_request,
  DROP COLUMN friend_accepted,
  DROP COLUMN challenge_received,
  DROP COLUMN challenge_declined,
  DROP COLUMN reengagement,
  DROP COLUMN weekly_reset,
  DROP COLUMN global_mute;

ALTER TABLE fcm_tokens
  RENAME COLUMN fcm_token TO token;

ALTER TABLE fcm_tokens
  RENAME COLUMN platform TO device_type;
