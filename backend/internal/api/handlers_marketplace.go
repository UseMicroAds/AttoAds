package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/microads/microads-backend/internal/auth"
	"github.com/microads/microads-backend/internal/db"
	"github.com/microads/microads-backend/internal/models"
	ytclient "github.com/microads/microads-backend/internal/youtube"
)

type MarketplaceHandlers struct {
	Store    *db.Store
	YTClient *ytclient.Client
}

type MarketplaceCommentResponse struct {
	ID                uuid.UUID            `json:"id"`
	VideoID           string               `json:"video_id"`
	VideoTitle        string               `json:"video_title,omitempty"`
	VideoCategory     *string              `json:"video_category,omitempty"`
	CommentID         string               `json:"comment_id"`
	AuthorChannelID   string               `json:"author_channel_id"`
	AuthorDisplayName string               `json:"author_display_name"`
	OriginalText      string               `json:"original_text"`
	LikeCount         int                  `json:"like_count"`
	Velocity          float64              `json:"velocity"`
	Status            models.CommentStatus `json:"status"`
	PublishedAt       *time.Time           `json:"published_at"`
	UpdatedAt         *time.Time           `json:"updated_at"`
	FirstSeen         time.Time            `json:"first_seen"`
}

func (h *MarketplaceHandlers) ListComments(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := h.Store.Pool.Query(
		r.Context(),
		`SELECT vc.id, tv.video_id, vc.comment_id, vc.author_channel_id, vc.author_display_name,
		        vc.original_text, vc.like_count, vc.velocity, vc.status, vc.published_at, vc.updated_at, vc.first_seen
		 FROM viral_comments vc
		 JOIN trending_videos tv ON tv.id = vc.video_id
		 WHERE vc.status = 'available'
		   AND vc.published_at >= now() - interval '5 days'
		 ORDER BY vc.velocity DESC
		 LIMIT $1 OFFSET $2`,
		limit,
		offset,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list comments")
		return
	}
	defer rows.Close()

	comments := make([]MarketplaceCommentResponse, 0, limit)
	for rows.Next() {
		var c MarketplaceCommentResponse
		if err := rows.Scan(
			&c.ID,
			&c.VideoID,
			&c.VideoTitle,
			&c.VideoCategory,
			&c.CommentID,
			&c.AuthorChannelID,
			&c.AuthorDisplayName,
			&c.OriginalText,
			&c.LikeCount,
			&c.Velocity,
			&c.Status,
			&c.PublishedAt,
			&c.UpdatedAt,
			&c.FirstSeen,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list comments")
			return
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list comments")
		return
	}

	writeJSON(w, http.StatusOK, comments)
}

func (h *MarketplaceHandlers) GetComment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid comment id")
		return
	}

	comment, err := h.Store.GetViralCommentByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "comment not found")
		return
	}

	writeJSON(w, http.StatusOK, comment)
}

type CreateDealRequest struct {
	CampaignID uuid.UUID `json:"campaign_id"`
	CommentID  uuid.UUID `json:"comment_id"`
}

func (h *MarketplaceHandlers) CreateDeal(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())

	var req CreateDealRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	comment, err := h.Store.GetViralCommentByID(r.Context(), req.CommentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "comment not found")
		return
	}

	ch, err := h.Store.GetYouTubeChannelByChannelID(r.Context(), comment.AuthorChannelID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "comment author not registered on platform")
		return
	}

	deal, err := h.Store.CreateDeal(r.Context(), req.CampaignID, req.CommentID, ch.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create deal")
		return
	}

	_ = claims
	writeJSON(w, http.StatusCreated, deal)
}

func (h *MarketplaceHandlers) ListMyDeals(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())

	deals, err := h.Store.ListDealsByCommenter(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list deals")
		return
	}

	writeJSON(w, http.StatusOK, deals)
}

func (h *MarketplaceHandlers) ListMyComments(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())

	ytCh, err := h.Store.GetYouTubeChannelByUser(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "no youtube channel linked")
		return
	}

	comments, err := h.Store.ListCommentsByAuthorChannel(r.Context(), ytCh.ChannelID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list comments")
		return
	}

	writeJSON(w, http.StatusOK, comments)
}

func (h *MarketplaceHandlers) ListCommentsByAuthorChannel(w http.ResponseWriter, r *http.Request) {
	channelID := chi.URLParam(r, "channelID")
	if channelID == "" {
		writeError(w, http.StatusBadRequest, "channel id is required")
		return
	}

	comments, err := h.Store.ListCommentsByAuthorChannel(r.Context(), channelID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list comments")
		return
	}

	writeJSON(w, http.StatusOK, comments)
}

type ChannelAuthoredComment struct {
	CommentID            string                `json:"comment_id"`
	VideoID              string                `json:"video_id"`
	AuthorChannelID      string                `json:"author_channel_id"`
	AuthorDisplayName    string                `json:"author_display_name"`
	Text                 string                `json:"text"`
	LikeCount            int64                 `json:"like_count"`
	PublishedAt          *time.Time            `json:"published_at,omitempty"`
	UpdatedAt            *time.Time            `json:"updated_at,omitempty"`
	MarketplaceCommentID *uuid.UUID            `json:"marketplace_comment_id,omitempty"`
	MarketplaceStatus    *models.CommentStatus `json:"marketplace_status,omitempty"`
}

func (h *MarketplaceHandlers) ListAllCommentsByAuthorChannel(w http.ResponseWriter, r *http.Request) {
	channelID := chi.URLParam(r, "channelID")
	if channelID == "" {
		writeError(w, http.StatusBadRequest, "channel id is required")
		return
	}

	if h.YTClient == nil {
		writeError(w, http.StatusInternalServerError, "youtube client is not configured")
		return
	}

	comments, err := h.YTClient.FetchCommentsAuthoredOnOwnChannel(r.Context(), channelID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch channel comments")
		return
	}

	writeJSON(w, http.StatusOK, h.mapMarketplaceStatus(r, comments))
}

func (h *MarketplaceHandlers) ListAllCommentsByVideo(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "videoID")
	if videoID == "" {
		writeError(w, http.StatusBadRequest, "video id is required")
		return
	}

	if h.YTClient == nil {
		writeError(w, http.StatusInternalServerError, "youtube client is not configured")
		return
	}

	comments, err := h.YTClient.FetchAllCommentsByVideo(r.Context(), videoID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch video comments")
		return
	}

	writeJSON(w, http.StatusOK, h.mapMarketplaceStatus(r, comments))
}

func (h *MarketplaceHandlers) mapMarketplaceStatus(r *http.Request, comments []ytclient.Comment) []ChannelAuthoredComment {
	out := make([]ChannelAuthoredComment, 0, len(comments))
	for _, c := range comments {
		row := ChannelAuthoredComment{
			CommentID:         c.CommentID,
			VideoID:           c.VideoID,
			AuthorChannelID:   c.AuthorChannelID,
			AuthorDisplayName: c.AuthorDisplayName,
			Text:              c.TextDisplay,
			LikeCount:         c.LikeCount,
			PublishedAt:       c.PublishedAt,
			UpdatedAt:         c.UpdatedAt,
		}

		vc, err := h.Store.GetViralCommentByCommentID(r.Context(), c.CommentID)
		if err == nil {
			row.MarketplaceCommentID = &vc.ID
			status := vc.Status
			row.MarketplaceStatus = &status
		}

		out = append(out, row)
	}
	return out
}

type RegisterCommentForTestingRequest struct {
	CommentID         string `json:"comment_id"`
	VideoID           string `json:"video_id"`
	AuthorChannelID   string `json:"author_channel_id"`
	AuthorDisplayName string `json:"author_display_name"`
	Text              string `json:"text"`
	LikeCount         int64  `json:"like_count"`
}

func (h *MarketplaceHandlers) RegisterCommentForTesting(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req RegisterCommentForTestingRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CommentID == "" || req.AuthorChannelID == "" {
		writeError(w, http.StatusBadRequest, "comment_id and author_channel_id are required")
		return
	}

	if req.Text == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}

	// Commenters may only register comments from their own linked YouTube channel
	if claims.Role == models.RoleCommenter {
		ch, err := h.Store.GetYouTubeChannelByUser(r.Context(), claims.UserID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "link your YouTube channel first")
			return
		}
		// Normalize for comparison: YouTube may return channel IDs with different casing/whitespace
		linkedID := strings.TrimSpace(ch.ChannelID)
		commentAuthorID := strings.TrimSpace(req.AuthorChannelID)
		if !strings.EqualFold(linkedID, commentAuthorID) {
			writeError(w, http.StatusForbidden, "you can only register comments from your own channel")
			return
		}
	}

	_, err := h.Store.GetYouTubeChannelByChannelID(r.Context(), req.AuthorChannelID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "comment author not registered on platform")
		return
	}

	existing, err := h.Store.GetViralCommentByCommentID(r.Context(), req.CommentID)
	if err == nil {
		writeJSON(w, http.StatusOK, existing)
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, "failed to check existing comment")
		return
	}

	videoID := req.VideoID
	videoTitle := "Testing Comment Source"
	videoChannelTitle := req.AuthorDisplayName
	thumbnailURL := ""
	viewCount := int64(0)

	if videoID == "" {
		videoID = "testing-channel-" + req.AuthorChannelID
	}

	var videoCategory *string
	if h.YTClient != nil && req.VideoID != "" {
		video, err := h.YTClient.FetchVideo(r.Context(), req.VideoID)
		if err == nil {
			videoID = video.VideoID
			videoTitle = video.Title
			videoChannelTitle = video.ChannelTitle
			thumbnailURL = video.ThumbnailURL
			viewCount = int64(video.ViewCount)
			if video.CategoryID != "" {
				titles, err := h.YTClient.FetchVideoCategoryTitles(r.Context(), "US")
				if err == nil {
					if t := titles[video.CategoryID]; t != "" {
						videoCategory = &t
					}
				}
			}
		}
	}
	tv, err := h.Store.UpsertTrendingVideo(
		r.Context(),
		videoID,
		videoTitle,
		videoChannelTitle,
		thumbnailURL,
		viewCount,
		videoCategory,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to upsert test video")
		return
	}

	vc, err := h.Store.UpsertViralComment(
		r.Context(),
		tv.ID,
		req.CommentID,
		req.AuthorChannelID,
		req.AuthorDisplayName,
		req.Text,
		int(req.LikeCount),
		nil,
		nil,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to register comment for testing")
		return
	}

	writeJSON(w, http.StatusCreated, vc)
}

func (h *MarketplaceHandlers) ListMyTransactions(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())

	txs, err := h.Store.ListTransactionsByCommenter(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list transactions")
		return
	}

	writeJSON(w, http.StatusOK, txs)
}

func (h *MarketplaceHandlers) ListTrendingVideos(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	videos, err := h.Store.ListTrendingVideos(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list videos")
		return
	}

	writeJSON(w, http.StatusOK, videos)
}
