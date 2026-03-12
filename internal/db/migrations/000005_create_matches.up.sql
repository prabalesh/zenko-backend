CREATE TABLE matches (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  player1_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  player2_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  mode game_mode NOT NULL,
  winner_id UUID REFERENCES users(id) ON DELETE SET NULL,
  is_bot_match BOOLEAN NOT NULL DEFAULT FALSE,
  player1_answers JSONB,
  player2_answers JSONB,
  forfeit_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
  played_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT no_self_match CHECK (player1_id != player2_id)
);

CREATE INDEX idx_matches_player1_id ON matches(player1_id);
CREATE INDEX idx_matches_player2_id ON matches(player2_id);
