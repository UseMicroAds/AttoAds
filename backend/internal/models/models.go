package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleCommenter  UserRole = "commenter"
	RoleAdvertiser UserRole = "advertiser"
)

type CommentStatus string

const (
	CommentAvailable CommentStatus = "available"
	CommentClaimed   CommentStatus = "claimed"
	CommentExpired   CommentStatus = "expired"
)

type CampaignStatus string

const (
	CampaignDraft     CampaignStatus = "draft"
	CampaignFunded    CampaignStatus = "funded"
	CampaignActive    CampaignStatus = "active"
	CampaignCompleted CampaignStatus = "completed"
)

type DealStatus string

const (
	DealPending     DealStatus = "pending"
	DealEditPending DealStatus = "edit_pending"
	DealVerified    DealStatus = "verified"
	DealPaid        DealStatus = "paid"
	DealFailed      DealStatus = "failed"
)

type TxStatus string

const (
	TxPending   TxStatus = "pending"
	TxConfirmed TxStatus = "confirmed"
	TxFailed    TxStatus = "failed"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type YouTubeChannel struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	ChannelID    string    `json:"channel_id"`
	ChannelTitle string    `json:"channel_title"`
	AccessToken  string    `json:"-"`
	RefreshToken string    `json:"-"`
	TokenExpiry  time.Time `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Wallet struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Address   string    `json:"address"`
	Chain     string    `json:"chain"`
	CreatedAt time.Time `json:"created_at"`
}

type TrendingVideo struct {
	ID            uuid.UUID `json:"id"`
	VideoID       string    `json:"video_id"`
	Title         string    `json:"title"`
	ChannelTitle  string    `json:"channel_title"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	ViewCount     int64     `json:"view_count"`
	VideoCategory *string   `json:"video_category,omitempty"`
	DiscoveredAt  time.Time `json:"discovered_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ViralComment struct {
	ID                uuid.UUID     `json:"id"`
	VideoID           uuid.UUID     `json:"video_id"`
	CommentID         string        `json:"comment_id"`
	AuthorChannelID   string        `json:"author_channel_id"`
	AuthorDisplayName string        `json:"author_display_name"`
	OriginalText      string        `json:"original_text"`
	LikeCount         int           `json:"like_count"`
	PrevLikeCount     int           `json:"prev_like_count"`
	Velocity          float64       `json:"velocity"`
	Status            CommentStatus `json:"status"`
	PublishedAt       *time.Time    `json:"published_at"`
	UpdatedAt         *time.Time    `json:"updated_at"`
	FirstSeen         time.Time     `json:"first_seen"`
	LastPolled        time.Time     `json:"last_polled"`
}

type Campaign struct {
	ID                uuid.UUID      `json:"id"`
	AdvertiserID      uuid.UUID      `json:"advertiser_id"`
	AdText            string         `json:"ad_text"`
	BudgetCents       int            `json:"budget_cents"`
	PricePerPlacement int            `json:"price_per_placement"`
	Status            CampaignStatus `json:"status"`
	EscrowTxHash      *string        `json:"escrow_tx_hash,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

type Bounty struct {
	ID                  uuid.UUID      `json:"id"`
	AdvertiserID        uuid.UUID      `json:"advertiser_id"`
	AdText              string         `json:"ad_text"`
	BudgetCents         int            `json:"budget_cents"`
	AmountPerClaimCents int            `json:"amount_per_claim_cents"`
	VideoCategory       *string        `json:"video_category,omitempty"`
	MinLikes            int            `json:"min_likes"`
	Status              CampaignStatus `json:"status"`
	EscrowTxHash        *string        `json:"escrow_tx_hash,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

type Deal struct {
	ID          uuid.UUID   `json:"id"`
	CampaignID  *uuid.UUID  `json:"campaign_id,omitempty"`
	BountyID    *uuid.UUID  `json:"bounty_id,omitempty"`
	CommentID   uuid.UUID   `json:"comment_id"`
	CommenterID uuid.UUID   `json:"commenter_id"`
	Status      DealStatus  `json:"status"`
	EditedAt    *time.Time  `json:"edited_at,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type DealCommentMetric struct {
	ID         uuid.UUID `json:"id"`
	DealID     uuid.UUID `json:"deal_id"`
	CapturedAt time.Time `json:"captured_at"`
	LikeCount  int       `json:"like_count"`
	Velocity   float64   `json:"velocity"`
}

type Transaction struct {
	ID         uuid.UUID `json:"id"`
	DealID     uuid.UUID `json:"deal_id"`
	TxHash     string    `json:"tx_hash"`
	AmountUSDC int       `json:"amount_usdc"`
	Status     TxStatus  `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
