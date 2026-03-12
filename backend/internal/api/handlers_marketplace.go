package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/microads/microads-backend/internal/auth"
	"github.com/microads/microads-backend/internal/db"
)

type MarketplaceHandlers struct {
	Store *db.Store
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

	comments, err := h.Store.ListAvailableComments(r.Context(), limit, offset)
	if err != nil {
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
