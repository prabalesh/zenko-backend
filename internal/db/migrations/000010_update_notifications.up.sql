ALTER TABLE notification_preferences
  DROP COLUMN friend_requests,
  DROP COLUMN challenges,
  DROP COLUMN marketing;

ALTER TABLE notification_preferences
  ADD COLUMN friend_request BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN friend_accepted BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN challenge_received BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN challenge_declined BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN reengagement BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN weekly_reset BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN global_mute BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE fcm_tokens
  RENAME COLUMN token TO fcm_token;

ALTER TABLE fcm_tokens
  RENAME COLUMN device_type TO platform;
