-- When the verifier sets deal status to edit_pending, we record when the comment was edited
ALTER TABLE deals
    ADD COLUMN IF NOT EXISTS edited_at TIMESTAMPTZ;

-- Time-series of like_count and velocity for deal comments (for performance graphs)
CREATE TABLE IF NOT EXISTS deal_comment_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    captured_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    like_count INT NOT NULL,
    velocity DOUBLE PRECISION NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_deal_comment_metrics_deal_captured ON deal_comment_metrics(deal_id, captured_at);
