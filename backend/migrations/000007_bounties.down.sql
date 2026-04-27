ALTER TABLE deals DROP CONSTRAINT IF EXISTS chk_deal_campaign_or_bounty;
DELETE FROM deals WHERE bounty_id IS NOT NULL;
ALTER TABLE deals DROP CONSTRAINT IF EXISTS deals_campaign_id_fkey;
ALTER TABLE deals DROP COLUMN IF EXISTS bounty_id;
ALTER TABLE deals ALTER COLUMN campaign_id SET NOT NULL;
ALTER TABLE deals ADD CONSTRAINT deals_campaign_id_fkey
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE;

DROP TABLE IF EXISTS bounties;

ALTER TABLE trending_videos DROP COLUMN IF EXISTS video_category;
