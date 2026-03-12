package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/microads/microads-backend/internal/models"
)

// --- Users ---

func (s *Store) CreateUser(ctx context.Context, email string, role models.UserRole) (*models.User, error) {
	u := &models.User{}
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO users (email, role) VALUES ($1, $2)
		 RETURNING id, email, role, created_at, updated_at`,
		email, role,
	).Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u := &models.User{}
	err := s.Pool.QueryRow(ctx,
		`SELECT id, email, role, created_at, updated_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	err := s.Pool.QueryRow(ctx,
		`SELECT id, email, role, created_at, updated_at FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// --- YouTube Channels ---

func (s *Store) UpsertYouTubeChannel(ctx context.Context, userID uuid.UUID, channelID, channelTitle, accessToken, refreshToken string, expiry time.Time) (*models.YouTubeChannel, error) {
	ch := &models.YouTubeChannel{}
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO youtube_channels (user_id, channel_id, channel_title, access_token, refresh_token, token_expiry)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (channel_id) DO UPDATE SET
		   access_token = EXCLUDED.access_token,
		   refresh_token = EXCLUDED.refresh_token,
		   token_expiry = EXCLUDED.token_expiry,
		   channel_title = EXCLUDED.channel_title,
		   updated_at = now()
		 RETURNING id, user_id, channel_id, channel_title, access_token, refresh_token, token_expiry, created_at, updated_at`,
		userID, channelID, channelTitle, accessToken, refreshToken, expiry,
	).Scan(&ch.ID, &ch.UserID, &ch.ChannelID, &ch.ChannelTitle, &ch.AccessToken, &ch.RefreshToken, &ch.TokenExpiry, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (s *Store) GetYouTubeChannelByUser(ctx context.Context, userID uuid.UUID) (*models.YouTubeChannel, error) {
	ch := &models.YouTubeChannel{}
	err := s.Pool.QueryRow(ctx,
		`SELECT id, user_id, channel_id, channel_title, access_token, refresh_token, token_expiry, created_at, updated_at
		 FROM youtube_channels WHERE user_id = $1`, userID,
	).Scan(&ch.ID, &ch.UserID, &ch.ChannelID, &ch.ChannelTitle, &ch.AccessToken, &ch.RefreshToken, &ch.TokenExpiry, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (s *Store) GetYouTubeChannelByChannelID(ctx context.Context, channelID string) (*models.YouTubeChannel, error) {
	ch := &models.YouTubeChannel{}
	err := s.Pool.QueryRow(ctx,
		`SELECT id, user_id, channel_id, channel_title, access_token, refresh_token, token_expiry, created_at, updated_at
		 FROM youtube_channels WHERE channel_id = $1`, channelID,
	).Scan(&ch.ID, &ch.UserID, &ch.ChannelID, &ch.ChannelTitle, &ch.AccessToken, &ch.RefreshToken, &ch.TokenExpiry, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

// --- Wallets ---

func (s *Store) UpsertWallet(ctx context.Context, userID uuid.UUID, address, chain string) (*models.Wallet, error) {
	w := &models.Wallet{}
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO wallets (user_id, address, chain) VALUES ($1, $2, $3)
		 ON CONFLICT (address) DO UPDATE SET user_id = EXCLUDED.user_id, chain = EXCLUDED.chain
		 RETURNING id, user_id, address, chain, created_at`,
		userID, address, chain,
	).Scan(&w.ID, &w.UserID, &w.Address, &w.Chain, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (s *Store) GetWalletByUser(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
	w := &models.Wallet{}
	err := s.Pool.QueryRow(ctx,
		`SELECT id, user_id, address, chain, created_at FROM wallets WHERE user_id = $1`, userID,
	).Scan(&w.ID, &w.UserID, &w.Address, &w.Chain, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	return w, nil
}

// --- Trending Videos ---

func (s *Store) UpsertTrendingVideo(ctx context.Context, videoID, title, channelTitle, thumbnailURL string, viewCount int64) (*models.TrendingVideo, error) {
	v := &models.TrendingVideo{}
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO trending_videos (video_id, title, channel_title, thumbnail_url, view_count)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (video_id) DO UPDATE SET
		   title = EXCLUDED.title,
		   channel_title = EXCLUDED.channel_title,
		   thumbnail_url = EXCLUDED.thumbnail_url,
		   view_count = EXCLUDED.view_count,
		   updated_at = now()
		 RETURNING id, video_id, title, channel_title, thumbnail_url, view_count, discovered_at, updated_at`,
		videoID, title, channelTitle, thumbnailURL, viewCount,
	).Scan(&v.ID, &v.VideoID, &v.Title, &v.ChannelTitle, &v.ThumbnailURL, &v.ViewCount, &v.DiscoveredAt, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (s *Store) ListTrendingVideos(ctx context.Context, limit int) ([]models.TrendingVideo, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, video_id, title, channel_title, thumbnail_url, view_count, discovered_at, updated_at
		 FROM trending_videos ORDER BY updated_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.TrendingVideo, error) {
		var v models.TrendingVideo
		err := row.Scan(&v.ID, &v.VideoID, &v.Title, &v.ChannelTitle, &v.ThumbnailURL, &v.ViewCount, &v.DiscoveredAt, &v.UpdatedAt)
		return v, err
	})
}

// --- Viral Comments ---

func (s *Store) UpsertViralComment(ctx context.Context, videoID uuid.UUID, commentID, authorChannelID, authorDisplayName, originalText string, likeCount int) (*models.ViralComment, error) {
	vc := &models.ViralComment{}
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO viral_comments (video_id, comment_id, author_channel_id, author_display_name, original_text, like_count)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (comment_id) DO UPDATE SET
		   prev_like_count = viral_comments.like_count,
		   like_count = EXCLUDED.like_count,
		   last_polled = now()
		 RETURNING id, video_id, comment_id, author_channel_id, author_display_name, original_text, like_count, prev_like_count, velocity, status, first_seen, last_polled`,
		videoID, commentID, authorChannelID, authorDisplayName, originalText, likeCount,
	).Scan(&vc.ID, &vc.VideoID, &vc.CommentID, &vc.AuthorChannelID, &vc.AuthorDisplayName, &vc.OriginalText,
		&vc.LikeCount, &vc.PrevLikeCount, &vc.Velocity, &vc.Status, &vc.FirstSeen, &vc.LastPolled)
	if err != nil {
		return nil, err
	}
	return vc, nil
}

func (s *Store) UpdateCommentVelocity(ctx context.Context, id uuid.UUID, velocity float64, status models.CommentStatus) error {
	_, err := s.Pool.Exec(ctx,
		`UPDATE viral_comments SET velocity = $1, status = $2 WHERE id = $3`,
		velocity, status, id)
	return err
}

func (s *Store) ListAvailableComments(ctx context.Context, limit, offset int) ([]models.ViralComment, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, video_id, comment_id, author_channel_id, author_display_name, original_text, like_count, prev_like_count, velocity, status, first_seen, last_polled
		 FROM viral_comments WHERE status = 'available' ORDER BY velocity DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.ViralComment, error) {
		var vc models.ViralComment
		err := row.Scan(&vc.ID, &vc.VideoID, &vc.CommentID, &vc.AuthorChannelID, &vc.AuthorDisplayName, &vc.OriginalText,
			&vc.LikeCount, &vc.PrevLikeCount, &vc.Velocity, &vc.Status, &vc.FirstSeen, &vc.LastPolled)
		return vc, err
	})
}

func (s *Store) GetViralCommentByID(ctx context.Context, id uuid.UUID) (*models.ViralComment, error) {
	vc := &models.ViralComment{}
	err := s.Pool.QueryRow(ctx,
		`SELECT id, video_id, comment_id, author_channel_id, author_display_name, original_text, like_count, prev_like_count, velocity, status, first_seen, last_polled
		 FROM viral_comments WHERE id = $1`, id,
	).Scan(&vc.ID, &vc.VideoID, &vc.CommentID, &vc.AuthorChannelID, &vc.AuthorDisplayName, &vc.OriginalText,
		&vc.LikeCount, &vc.PrevLikeCount, &vc.Velocity, &vc.Status, &vc.FirstSeen, &vc.LastPolled)
	if err != nil {
		return nil, err
	}
	return vc, nil
}

func (s *Store) ListCommentsByAuthorChannel(ctx context.Context, authorChannelID string) ([]models.ViralComment, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, video_id, comment_id, author_channel_id, author_display_name, original_text, like_count, prev_like_count, velocity, status, first_seen, last_polled
		 FROM viral_comments WHERE author_channel_id = $1 ORDER BY velocity DESC`, authorChannelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.ViralComment, error) {
		var vc models.ViralComment
		err := row.Scan(&vc.ID, &vc.VideoID, &vc.CommentID, &vc.AuthorChannelID, &vc.AuthorDisplayName, &vc.OriginalText,
			&vc.LikeCount, &vc.PrevLikeCount, &vc.Velocity, &vc.Status, &vc.FirstSeen, &vc.LastPolled)
		return vc, err
	})
}

// --- Campaigns ---

func (s *Store) CreateCampaign(ctx context.Context, advertiserID uuid.UUID, adText string, budgetCents, pricePerPlacement int) (*models.Campaign, error) {
	c := &models.Campaign{}
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO campaigns (advertiser_id, ad_text, budget_cents, price_per_placement)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, advertiser_id, ad_text, budget_cents, price_per_placement, status, escrow_tx_hash, created_at, updated_at`,
		advertiserID, adText, budgetCents, pricePerPlacement,
	).Scan(&c.ID, &c.AdvertiserID, &c.AdText, &c.BudgetCents, &c.PricePerPlacement, &c.Status, &c.EscrowTxHash, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) GetCampaignByID(ctx context.Context, id uuid.UUID) (*models.Campaign, error) {
	c := &models.Campaign{}
	err := s.Pool.QueryRow(ctx,
		`SELECT id, advertiser_id, ad_text, budget_cents, price_per_placement, status, escrow_tx_hash, created_at, updated_at
		 FROM campaigns WHERE id = $1`, id,
	).Scan(&c.ID, &c.AdvertiserID, &c.AdText, &c.BudgetCents, &c.PricePerPlacement, &c.Status, &c.EscrowTxHash, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) ListCampaignsByAdvertiser(ctx context.Context, advertiserID uuid.UUID) ([]models.Campaign, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, advertiser_id, ad_text, budget_cents, price_per_placement, status, escrow_tx_hash, created_at, updated_at
		 FROM campaigns WHERE advertiser_id = $1 ORDER BY created_at DESC`, advertiserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Campaign, error) {
		var c models.Campaign
		err := row.Scan(&c.ID, &c.AdvertiserID, &c.AdText, &c.BudgetCents, &c.PricePerPlacement, &c.Status, &c.EscrowTxHash, &c.CreatedAt, &c.UpdatedAt)
		return c, err
	})
}

func (s *Store) UpdateCampaignStatus(ctx context.Context, id uuid.UUID, status models.CampaignStatus, escrowTxHash *string) error {
	_, err := s.Pool.Exec(ctx,
		`UPDATE campaigns SET status = $1, escrow_tx_hash = COALESCE($2, escrow_tx_hash), updated_at = now() WHERE id = $3`,
		status, escrowTxHash, id)
	return err
}

func (s *Store) ListActiveCampaigns(ctx context.Context) ([]models.Campaign, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, advertiser_id, ad_text, budget_cents, price_per_placement, status, escrow_tx_hash, created_at, updated_at
		 FROM campaigns WHERE status IN ('funded', 'active') ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Campaign, error) {
		var c models.Campaign
		err := row.Scan(&c.ID, &c.AdvertiserID, &c.AdText, &c.BudgetCents, &c.PricePerPlacement, &c.Status, &c.EscrowTxHash, &c.CreatedAt, &c.UpdatedAt)
		return c, err
	})
}

// --- Deals ---

func (s *Store) CreateDeal(ctx context.Context, campaignID, commentID, commenterID uuid.UUID) (*models.Deal, error) {
	d := &models.Deal{}
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO deals (campaign_id, comment_id, commenter_id) VALUES ($1, $2, $3)
		 RETURNING id, campaign_id, comment_id, commenter_id, status, created_at, updated_at`,
		campaignID, commentID, commenterID,
	).Scan(&d.ID, &d.CampaignID, &d.CommentID, &d.CommenterID, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Store) UpdateDealStatus(ctx context.Context, id uuid.UUID, status models.DealStatus) error {
	_, err := s.Pool.Exec(ctx,
		`UPDATE deals SET status = $1, updated_at = now() WHERE id = $2`, status, id)
	return err
}

func (s *Store) ListDealsByCampaign(ctx context.Context, campaignID uuid.UUID) ([]models.Deal, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, campaign_id, comment_id, commenter_id, status, created_at, updated_at
		 FROM deals WHERE campaign_id = $1 ORDER BY created_at DESC`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Deal, error) {
		var d models.Deal
		err := row.Scan(&d.ID, &d.CampaignID, &d.CommentID, &d.CommenterID, &d.Status, &d.CreatedAt, &d.UpdatedAt)
		return d, err
	})
}

func (s *Store) ListDealsByCommenter(ctx context.Context, commenterID uuid.UUID) ([]models.Deal, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, campaign_id, comment_id, commenter_id, status, created_at, updated_at
		 FROM deals WHERE commenter_id = $1 ORDER BY created_at DESC`, commenterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Deal, error) {
		var d models.Deal
		err := row.Scan(&d.ID, &d.CampaignID, &d.CommentID, &d.CommenterID, &d.Status, &d.CreatedAt, &d.UpdatedAt)
		return d, err
	})
}

func (s *Store) ListPendingDeals(ctx context.Context) ([]models.Deal, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, campaign_id, comment_id, commenter_id, status, created_at, updated_at
		 FROM deals WHERE status IN ('pending', 'edit_pending') ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Deal, error) {
		var d models.Deal
		err := row.Scan(&d.ID, &d.CampaignID, &d.CommentID, &d.CommenterID, &d.Status, &d.CreatedAt, &d.UpdatedAt)
		return d, err
	})
}

func (s *Store) GetDealByID(ctx context.Context, id uuid.UUID) (*models.Deal, error) {
	d := &models.Deal{}
	err := s.Pool.QueryRow(ctx,
		`SELECT id, campaign_id, comment_id, commenter_id, status, created_at, updated_at
		 FROM deals WHERE id = $1`, id,
	).Scan(&d.ID, &d.CampaignID, &d.CommentID, &d.CommenterID, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// --- Transactions ---

func (s *Store) CreateTransaction(ctx context.Context, dealID uuid.UUID, txHash string, amountUSDC int) (*models.Transaction, error) {
	t := &models.Transaction{}
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO transactions (deal_id, tx_hash, amount_usdc) VALUES ($1, $2, $3)
		 RETURNING id, deal_id, tx_hash, amount_usdc, status, created_at, updated_at`,
		dealID, txHash, amountUSDC,
	).Scan(&t.ID, &t.DealID, &t.TxHash, &t.AmountUSDC, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Store) UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status models.TxStatus) error {
	_, err := s.Pool.Exec(ctx,
		`UPDATE transactions SET status = $1, updated_at = now() WHERE id = $2`, status, id)
	return err
}

func (s *Store) GetTransactionByDeal(ctx context.Context, dealID uuid.UUID) (*models.Transaction, error) {
	t := &models.Transaction{}
	err := s.Pool.QueryRow(ctx,
		`SELECT id, deal_id, tx_hash, amount_usdc, status, created_at, updated_at
		 FROM transactions WHERE deal_id = $1`, dealID,
	).Scan(&t.ID, &t.DealID, &t.TxHash, &t.AmountUSDC, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Store) ListTransactionsByCommenter(ctx context.Context, commenterID uuid.UUID) ([]models.Transaction, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT t.id, t.deal_id, t.tx_hash, t.amount_usdc, t.status, t.created_at, t.updated_at
		 FROM transactions t
		 JOIN deals d ON d.id = t.deal_id
		 WHERE d.commenter_id = $1
		 ORDER BY t.created_at DESC`, commenterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Transaction, error) {
		var t models.Transaction
		err := row.Scan(&t.ID, &t.DealID, &t.TxHash, &t.AmountUSDC, &t.Status, &t.CreatedAt, &t.UpdatedAt)
		return t, err
	})
}
