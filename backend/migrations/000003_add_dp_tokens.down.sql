ALTER TABLE youtube_channels
  DROP COLUMN IF EXISTS dp_access_token,
  DROP COLUMN IF EXISTS dp_refresh_token,
  DROP COLUMN IF EXISTS dp_token_expiry;
