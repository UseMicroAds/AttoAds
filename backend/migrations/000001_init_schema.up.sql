CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE user_role AS ENUM ('commenter', 'advertiser');
CREATE TYPE comment_status AS ENUM ('available', 'claimed', 'expired');
CREATE TYPE campaign_status AS ENUM ('draft', 'funded', 'active', 'completed');
CREATE TYPE deal_status AS ENUM ('pending', 'edit_pending', 'verified', 'paid', 'failed');
CREATE TYPE tx_status AS ENUM ('pending', 'confirmed', 'failed');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    role user_role NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE youtube_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel_id TEXT NOT NULL UNIQUE,
    channel_title TEXT NOT NULL DEFAULT '',
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expiry TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_youtube_channels_user ON youtube_channels(user_id);

CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address TEXT NOT NULL UNIQUE,
    chain TEXT NOT NULL DEFAULT 'base',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_wallets_user ON wallets(user_id);

CREATE TABLE trending_videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    channel_title TEXT NOT NULL DEFAULT '',
    thumbnail_url TEXT NOT NULL DEFAULT '',
    view_count BIGINT NOT NULL DEFAULT 0,
    discovered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE viral_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES trending_videos(id) ON DELETE CASCADE,
    comment_id TEXT NOT NULL UNIQUE,
    author_channel_id TEXT NOT NULL,
    author_display_name TEXT NOT NULL DEFAULT '',
    original_text TEXT NOT NULL,
    like_count INT NOT NULL DEFAULT 0,
    prev_like_count INT NOT NULL DEFAULT 0,
    velocity DOUBLE PRECISION NOT NULL DEFAULT 0,
    status comment_status NOT NULL DEFAULT 'available',
    first_seen TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_polled TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_viral_comments_video ON viral_comments(video_id);
CREATE INDEX idx_viral_comments_status ON viral_comments(status);
CREATE INDEX idx_viral_comments_author ON viral_comments(author_channel_id);

CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ad_text TEXT NOT NULL,
    budget_cents INT NOT NULL,
    price_per_placement INT NOT NULL,
    status campaign_status NOT NULL DEFAULT 'draft',
    escrow_tx_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_campaigns_advertiser ON campaigns(advertiser_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);

CREATE TABLE deals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    comment_id UUID NOT NULL REFERENCES viral_comments(id),
    commenter_id UUID NOT NULL REFERENCES users(id),
    status deal_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_deals_campaign ON deals(campaign_id);
CREATE INDEX idx_deals_commenter ON deals(commenter_id);
CREATE INDEX idx_deals_status ON deals(status);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE UNIQUE,
    tx_hash TEXT NOT NULL,
    amount_usdc INT NOT NULL,
    status tx_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
