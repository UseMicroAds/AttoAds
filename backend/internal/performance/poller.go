package performance

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/microads/microads-backend/internal/db"
	ytclient "github.com/microads/microads-backend/internal/youtube"
)

type Poller struct {
	store        *db.Store
	ytClient     *ytclient.Client
	pollInterval time.Duration
}

func NewPoller(store *db.Store, ytClient *ytclient.Client, pollInterval time.Duration) *Poller {
	return &Poller{
		store:        store,
		ytClient:     ytClient,
		pollInterval: pollInterval,
	}
}

func (p *Poller) Run(ctx context.Context) {
	slog.Info("performance poller starting", "interval", p.pollInterval)
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	// Run once immediately
	p.poll(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("performance poller shutting down")
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func (p *Poller) poll(ctx context.Context) {
	deals, err := p.store.ListDealsForMetricsPolling(ctx)
	if err != nil {
		slog.Error("performance: failed to list deals", "error", err)
		return
	}
	if len(deals) == 0 {
		return
	}

	for _, deal := range deals {
		vc, err := p.store.GetViralCommentByID(ctx, deal.CommentID)
		if err != nil {
			slog.Error("performance: failed to get viral comment", "deal_id", deal.ID, "error", err)
			continue
		}

		newLikeCount, err := p.ytClient.FetchCommentLikeCount(ctx, vc.CommentID)
		if err != nil {
			slog.Warn("performance: failed to fetch comment like count", "deal_id", deal.ID, "comment_id", vc.CommentID, "error", err)
			continue
		}

		minutesSince := time.Since(vc.LastPolled).Minutes()
		if minutesSince < 1 {
			minutesSince = 1
		}
		velocity := float64(int(newLikeCount)-vc.LikeCount) / minutesSince
		if velocity < 0 {
			velocity = 0
		}
		velocity = math.Round(velocity*100) / 100

		if err := p.store.UpdateViralCommentLikesAndVelocity(ctx, vc.ID, int(newLikeCount), velocity); err != nil {
			slog.Error("performance: failed to update viral comment", "deal_id", deal.ID, "error", err)
			continue
		}
		if err := p.store.InsertDealCommentMetric(ctx, deal.ID, int(newLikeCount), velocity); err != nil {
			slog.Error("performance: failed to insert metric", "deal_id", deal.ID, "error", err)
		}
	}

	slog.Info("performance: poll complete", "deals", len(deals))
}
