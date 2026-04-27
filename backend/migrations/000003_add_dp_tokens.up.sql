ALTER TABLE youtube_channels
  ADD COLUMN IF NOT EXISTS dp_access_token TEXT,
  ADD COLUMN IF NOT EXISTS dp_refresh_token TEXT,
  ADD COLUMN IF NOT EXISTS dp_token_expiry TIMESTAMPTZ;
