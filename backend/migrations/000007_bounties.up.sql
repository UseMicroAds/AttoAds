-- Video category for bounty filtering (optional; discovery can populate later)
ALTER TABLE trending_videos
    ADD COLUMN IF NOT EXISTS video_category TEXT;

-- Bounties: advertiser-funded pool with constraints (category, min_likes)
CREATE TABLE IF NOT EXISTS bounties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ad_text TEXT NOT NULL,
    budget_cents INT NOT NULL,
    amount_per_claim_cents INT NOT NULL,
    video_category TEXT,
    min_likes INT NOT NULL DEFAULT 0,
    status campaign_status NOT NULL DEFAULT 'draft',
    escrow_tx_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_bounties_advertiser ON bounties(advertiser_id);
CREATE INDEX idx_bounties_status ON bounties(status);

-- Deals can be for a campaign OR a bounty
ALTER TABLE deals
    DROP CONSTRAINT IF EXISTS deals_campaign_id_fkey;
ALTER TABLE deals
    ALTER COLUMN campaign_id DROP NOT NULL;
ALTER TABLE deals
    ADD COLUMN IF NOT EXISTS bounty_id UUID REFERENCES bounties(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_deals_bounty ON deals(bounty_id);
ALTER TABLE deals
    ADD CONSTRAINT chk_deal_campaign_or_bounty CHECK (
        (campaign_id IS NOT NULL AND bounty_id IS NULL) OR
        (campaign_id IS NULL AND bounty_id IS NOT NULL)
    );
-- Re-add FK for campaign_id (nullable; NULL allowed for bounty-only deals)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'deals_campaign_id_fkey') THEN
        ALTER TABLE deals ADD CONSTRAINT deals_campaign_id_fkey
            FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;
    END IF;
END $$;
