ALTER TABLE viral_comments
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS published_at;
