-- Ensure one wallet per user: dedupe then add unique constraint if missing.
-- Idempotent for when 000002 was skipped due to duplicate migration version.
WITH ranked_wallets AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY user_id
            ORDER BY created_at DESC, id DESC
        ) AS rn
    FROM wallets
)
DELETE FROM wallets w
USING ranked_wallets r
WHERE w.id = r.id
  AND r.rn > 1;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'wallets'::regclass
          AND conname = 'uq_wallets_user_id'
    ) THEN
        ALTER TABLE wallets
            ADD CONSTRAINT uq_wallets_user_id UNIQUE (user_id);
    END IF;
END $$;
