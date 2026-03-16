package performance

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/microads/microads-backend/internal/db"
	ytclient "github.com/microads/microads-backend/internal/youtube"
)

type Poller struct {
	store        *db.Store
	ytClient     *ytclient.Client
	pollInterval time.Duration

	// per-comment state used for velocity calculation within the performance pipeline
	lastLikeCount map[uuid.UUID]int
	lastPolled    map[uuid.UUID]time.Time
}

func NewPoller(store *db.Store, ytClient *ytclient.Client, pollInterval time.Duration) *Poller {
	return &Poller{
		store:        store,
		ytClient:     ytClient,
		pollInterval: pollInterval,
		lastLikeCount: make(map[uuid.UUID]int),
		lastPolled:    make(map[uuid.UUID]time.Time),
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

		// Compute velocity based on per-comment state tracked within this poller,
		// instead of relying on vc.LastPolled/vc.LikeCount which may be updated
		// by other workers (e.g. discovery).
		prevLikeCount, hasPrevLike := p.lastLikeCount[vc.ID]
		prevPolledAt, hasPrevPolled := p.lastPolled[vc.ID]

		var velocity float64
		if hasPrevLike && hasPrevPolled {
			minutesSince := time.Since(prevPolledAt).Minutes()
			if minutesSince < 1 {
				minutesSince = 1
			}
			velocity = float64(int(newLikeCount)-prevLikeCount) / minutesSince
			if velocity < 0 {
				velocity = 0
			}
			velocity = math.Round(velocity*100) / 100
		} else {
			// No previous state for this comment in the performance pipeline;
			// establish a baseline with zero velocity.
			velocity = 0
		}

		if err := p.store.UpdateViralCommentLikesAndVelocity(ctx, vc.ID, int(newLikeCount), velocity); err != nil {
			slog.Error("performance: failed to update viral comment", "deal_id", deal.ID, "error", err)
			continue
		}
		if err := p.store.InsertDealCommentMetric(ctx, deal.ID, int(newLikeCount), velocity); err != nil {
			slog.Error("performance: failed to insert metric", "deal_id", deal.ID, "error", err)
		}

		// Update per-comment state for the next poll cycle.
		p.lastLikeCount[vc.ID] = int(newLikeCount)
		p.lastPolled[vc.ID] = time.Now()
	}

	slog.Info("performance: poll complete", "deals", len(deals))
}
