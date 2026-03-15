-- Keep only the newest wallet row per user before enforcing uniqueness.
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

ALTER TABLE wallets
    ADD CONSTRAINT uq_wallets_user_id UNIQUE (user_id);
