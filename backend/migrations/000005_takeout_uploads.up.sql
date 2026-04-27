-- User-uploaded Takeout archives (when Data Portability API is not available).
CREATE TABLE IF NOT EXISTS takeout_uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel_id TEXT NOT NULL REFERENCES youtube_channels(channel_id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processed', 'failed')),
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_takeout_uploads_status ON takeout_uploads(status);
CREATE INDEX IF NOT EXISTS idx_takeout_uploads_channel ON takeout_uploads(channel_id);
