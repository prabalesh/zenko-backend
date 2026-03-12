CREATE TABLE friends (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  receiver_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status friend_status NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT no_self_friend CHECK (sender_id != receiver_id),
  UNIQUE(sender_id, receiver_id)
);

CREATE TRIGGER set_friends_updated_at
BEFORE UPDATE ON friends
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_friends_sender_id ON friends(sender_id);
CREATE INDEX idx_friends_receiver_id ON friends(receiver_id);
