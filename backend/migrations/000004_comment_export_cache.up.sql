-- Cache for daily Data Portability / Takeout comment export per channel.
CREATE TABLE IF NOT EXISTS comment_export_sync (
    channel_id TEXT PRIMARY KEY REFERENCES youtube_channels(channel_id) ON DELETE CASCADE,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    comment_count INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS comment_export_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id TEXT NOT NULL REFERENCES youtube_channels(channel_id) ON DELETE CASCADE,
    comment_id TEXT NOT NULL,
    video_id TEXT NOT NULL,
    text_display TEXT NOT NULL DEFAULT '',
    like_count BIGINT NOT NULL DEFAULT 0,
    published_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(channel_id, comment_id)
);
CREATE INDEX IF NOT EXISTS idx_comment_export_comments_channel ON comment_export_comments(channel_id);
