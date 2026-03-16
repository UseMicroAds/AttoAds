package discovery

import (
	"context"
	"log/slog"
	"time"

	"github.com/microads/microads-backend/internal/db"
	"github.com/microads/microads-backend/internal/models"
	"github.com/microads/microads-backend/internal/youtube"
)

type Engine struct {
	store             *db.Store
	ytClient          *youtube.Client
	regionCode        string
	maxVideos         int64
	maxComments       int64
	velocityThreshold float64
	pollInterval      time.Duration
}

type EngineConfig struct {
	RegionCode        string
	MaxVideos         int64
	MaxComments       int64
	VelocityThreshold float64
	PollInterval      time.Duration
}

func NewEngine(store *db.Store, ytClient *youtube.Client, cfg EngineConfig) *Engine {
	return &Engine{
		store:             store,
		ytClient:          ytClient,
		regionCode:        cfg.RegionCode,
		maxVideos:         cfg.MaxVideos,
		maxComments:       cfg.MaxComments,
		velocityThreshold: cfg.VelocityThreshold,
		pollInterval:      cfg.PollInterval,
	}
}

func (e *Engine) Run(ctx context.Context) {
	slog.Info("discovery engine starting",
		"interval", e.pollInterval,
		"threshold", e.velocityThreshold,
	)

	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()

	// Run once immediately
	e.poll(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("discovery engine shutting down")
			return
		case <-ticker.C:
			e.poll(ctx)
		}
	}
}

func (e *Engine) poll(ctx context.Context) {
	slog.Info("discovery: polling trending videos")

	videos, err := e.ytClient.FetchTrendingVideos(ctx, e.regionCode, e.maxVideos)
	if err != nil {
		slog.Error("discovery: failed to fetch trending videos", "error", err)
		return
	}

	categoryTitles, err := e.ytClient.FetchVideoCategoryTitles(ctx, e.regionCode)
	if err != nil {
		slog.Warn("discovery: failed to fetch category titles, videos will have no category", "error", err)
		categoryTitles = nil
	}

	for _, v := range videos {
		var videoCategory *string
		if v.CategoryID != "" {
			if categoryTitles != nil {
				if title, ok := categoryTitles[v.CategoryID]; ok && title != "" {
					videoCategory = &title
				}
			}
			// Fallback: API gave us a category ID but it wasn't in the region's list (or fetch failed)
			if videoCategory == nil {
				fallback := "Category " + v.CategoryID
				videoCategory = &fallback
			}
		}
		tv, err := e.store.UpsertTrendingVideo(ctx, v.VideoID, v.Title, v.ChannelTitle, v.ThumbnailURL, int64(v.ViewCount), videoCategory)
		if err != nil {
			slog.Error("discovery: failed to upsert video", "video_id", v.VideoID, "error", err)
			continue
		}

		comments, err := e.ytClient.FetchTopComments(ctx, v.VideoID, e.maxComments)
		if err != nil {
			slog.Error("discovery: failed to fetch comments", "video_id", v.VideoID, "error", err)
			continue
		}

		for _, c := range comments {
			vc, err := e.store.UpsertViralComment(ctx, tv.ID, c.CommentID, c.AuthorChannelID, c.AuthorDisplayName, c.TextDisplay, int(c.LikeCount))
			if err != nil {
				slog.Error("discovery: failed to upsert comment", "comment_id", c.CommentID, "error", err)
				continue
			}

			velocity := e.calculateVelocity(ctx, vc)
			// Record metrics for deals so advertisers can see performance over time
			dealIDs, _ := e.store.ListDealIDsByViralCommentID(ctx, vc.ID)
			for _, dealID := range dealIDs {
				_ = e.store.InsertDealCommentMetric(ctx, dealID, vc.LikeCount, velocity)
			}
		}
	}

	slog.Info("discovery: poll complete", "videos", len(videos))
}

func (e *Engine) calculateVelocity(ctx context.Context, vc *models.ViralComment) float64 {
	elapsed := time.Since(vc.FirstSeen).Minutes()
	if elapsed < 1 {
		elapsed = 1
	}

	deltaLikes := float64(vc.LikeCount - vc.PrevLikeCount)
	timeSinceLastPoll := time.Since(vc.LastPolled).Minutes()
	if timeSinceLastPoll < 1 {
		timeSinceLastPoll = 1
	}

	velocity := deltaLikes / timeSinceLastPoll

	status := vc.Status
	if velocity >= e.velocityThreshold && vc.Status != models.CommentClaimed {
		status = models.CommentAvailable
		slog.Info("discovery: comment flagged as available",
			"comment_id", vc.CommentID,
			"velocity", velocity,
		)
	}

	if err := e.store.UpdateCommentVelocity(ctx, vc.ID, velocity, status); err != nil {
		slog.Error("discovery: failed to update velocity", "comment_id", vc.CommentID, "error", err)
	}
	return velocity
}
