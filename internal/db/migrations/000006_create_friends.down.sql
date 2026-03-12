DROP INDEX IF EXISTS idx_friends_receiver_id;
DROP INDEX IF EXISTS idx_friends_sender_id;
DROP TRIGGER IF EXISTS set_friends_updated_at ON friends;
DROP TABLE IF EXISTS friends;
