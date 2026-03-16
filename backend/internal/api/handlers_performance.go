package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/microads/microads-backend/internal/auth"
	"github.com/microads/microads-backend/internal/db"
)

type PerformanceHandlers struct {
	Store *db.Store
}

// DealPerformanceRow is one row in the advertiser's deal performance list.
type DealPerformanceRow struct {
	ID                string   `json:"id"`
	CampaignID        *string  `json:"campaign_id,omitempty"`
	BountyID          *string  `json:"bounty_id,omitempty"`
	Status            string   `json:"status"`
	EditedAt          *string  `json:"edited_at,omitempty"`
	CreatedAt         string   `json:"created_at"`
	CommentLink       string   `json:"comment_link"`
	OriginalText      string   `json:"original_text"`
	LikeCount         int      `json:"like_count"`
	Velocity          float64  `json:"velocity"`
}

func (h *PerformanceHandlers) ListDeals(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	rows, err := h.Store.ListDealsForAdvertiser(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list deals")
		return
	}

	out := make([]DealPerformanceRow, 0, len(rows))
	for _, row := range rows {
		link := fmt.Sprintf("https://www.youtube.com/watch?v=%s&lc=%s", row.YouTubeVideoID, row.YouTubeCommentID)
		d := row.Deal
		o := DealPerformanceRow{
			ID:           d.ID.String(),
			Status:       string(d.Status),
			CreatedAt:    d.CreatedAt.Format(time.RFC3339),
			CommentLink:  link,
			OriginalText: row.OriginalText,
			LikeCount:    row.LikeCount,
			Velocity:     row.Velocity,
		}
		if d.CampaignID != nil {
			s := d.CampaignID.String()
			o.CampaignID = &s
		}
		if d.BountyID != nil {
			s := d.BountyID.String()
			o.BountyID = &s
		}
		if d.EditedAt != nil {
			s := d.EditedAt.Format(time.RFC3339)
			o.EditedAt = &s
		}
		out = append(out, o)
	}

	writeJSON(w, http.StatusOK, out)
}

// DealPerformanceResponse is the response for a single deal's performance (metrics for graph).
type DealPerformanceResponse struct {
	Deal        DealPerformanceRow        `json:"deal"`
	Metrics     []DealMetricPoint         `json:"metrics"`
}

type DealMetricPoint struct {
	CapturedAt string  `json:"captured_at"`
	LikeCount  int     `json:"like_count"`
	Velocity   float64 `json:"velocity"`
}

func (h *PerformanceHandlers) GetDealPerformance(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	dealID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid deal id")
		return
	}

	found, err := h.Store.GetDealForAdvertiser(r.Context(), claims.UserID, dealID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get deal")
		return
	}
	if found == nil {
		writeError(w, http.StatusNotFound, "deal not found")
		return
	}

	metrics, err := h.Store.GetDealCommentMetrics(r.Context(), dealID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get metrics")
		return
	}

	link := fmt.Sprintf("https://www.youtube.com/watch?v=%s&lc=%s", found.YouTubeVideoID, found.YouTubeCommentID)
	d := found.Deal
	row := DealPerformanceRow{
		ID:           d.ID.String(),
		Status:       string(d.Status),
		CreatedAt:    d.CreatedAt.Format(time.RFC3339),
		CommentLink:  link,
		OriginalText: found.OriginalText,
		LikeCount:    found.LikeCount,
		Velocity:     found.Velocity,
	}
	if d.CampaignID != nil {
		s := d.CampaignID.String()
		row.CampaignID = &s
	}
	if d.BountyID != nil {
		s := d.BountyID.String()
		row.BountyID = &s
	}
	if d.EditedAt != nil {
		s := d.EditedAt.Format(time.RFC3339)
		row.EditedAt = &s
	}

	points := make([]DealMetricPoint, 0, len(metrics))
	for _, m := range metrics {
		points = append(points, DealMetricPoint{
			CapturedAt: m.CapturedAt.Format(time.RFC3339),
			LikeCount:  m.LikeCount,
			Velocity:   m.Velocity,
		})
	}

	writeJSON(w, http.StatusOK, DealPerformanceResponse{Deal: row, Metrics: points})
}
