DROP INDEX IF EXISTS idx_challenges_challenged_id;
DROP INDEX IF EXISTS idx_challenges_challenger_id;
DROP TRIGGER IF EXISTS set_challenges_updated_at ON challenges;
DROP TABLE IF EXISTS challenges;
