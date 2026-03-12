CREATE TABLE leaderboard_snapshots (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  mode game_mode NOT NULL,
  week_start DATE NOT NULL,
  score INTEGER NOT NULL,
  rank INTEGER NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, mode, week_start)
);

CREATE INDEX idx_leaderboard_snapshots_week_mode ON leaderboard_snapshots(week_start, mode);
