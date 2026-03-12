CREATE TYPE friend_status AS ENUM ('pending', 'accepted', 'blocked');
CREATE TYPE challenge_status AS ENUM ('pending', 'accepted', 'declined', 'expired');
CREATE TYPE notification_type AS ENUM ('friend_request','friend_accepted','challenge_in','challenge_declined','weekly_reset','reengagement');
CREATE TYPE game_mode AS ENUM ('arithmetic', 'memory', 'operator');
CREATE TYPE platform_type AS ENUM ('ios', 'android');
CREATE TYPE social_platform AS ENUM ('instagram', 'twitter', 'github', 'linkedin', 'youtube');

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
