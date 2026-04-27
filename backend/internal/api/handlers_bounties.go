package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/microads/microads-backend/internal/auth"
	"github.com/microads/microads-backend/internal/db"
	"github.com/microads/microads-backend/internal/models"
)

type BountyHandlers struct {
	Store *db.Store
}

type CreateBountyRequest struct {
	AdText                string  `json:"ad_text"`
	BudgetCents           int     `json:"budget_cents"`
	AmountPerClaimCents   int     `json:"amount_per_claim_cents"`
	VideoCategory         *string `json:"video_category,omitempty"`
	MinLikes              int     `json:"min_likes"`
}

func (h *BountyHandlers) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateBountyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.AdText == "" || req.BudgetCents <= 0 || req.AmountPerClaimCents <= 0 {
		writeError(w, http.StatusBadRequest, "ad_text, budget_cents, and amount_per_claim_cents are required")
		return
	}
	if req.AmountPerClaimCents > req.BudgetCents {
		writeError(w, http.StatusBadRequest, "amount_per_claim_cents cannot exceed budget_cents")
		return
	}

	bounty, err := h.Store.CreateBounty(r.Context(), claims.UserID, req.AdText, req.BudgetCents, req.AmountPerClaimCents, req.VideoCategory, req.MinLikes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create bounty")
		return
	}

	writeJSON(w, http.StatusCreated, bounty)
}

func (h *BountyHandlers) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	bounties, err := h.Store.ListBountiesByAdvertiser(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list bounties")
		return
	}

	writeJSON(w, http.StatusOK, bounties)
}

func (h *BountyHandlers) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bounty id")
		return
	}

	bounty, err := h.Store.GetBountyByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "bounty not found")
		return
	}

	writeJSON(w, http.StatusOK, bounty)
}

func (h *BountyHandlers) ListActive(w http.ResponseWriter, r *http.Request) {
	bounties, err := h.Store.ListActiveBounties(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list bounties")
		return
	}

	writeJSON(w, http.StatusOK, bounties)
}

type FundBountyRequest struct {
	TxHash string `json:"tx_hash"`
}

func (h *BountyHandlers) Fund(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bounty id")
		return
	}

	bounty, err := h.Store.GetBountyByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "bounty not found")
		return
	}
	if bounty.AdvertiserID != claims.UserID {
		writeError(w, http.StatusForbidden, "not your bounty")
		return
	}
	if bounty.Status != models.CampaignDraft {
		writeError(w, http.StatusBadRequest, "bounty already funded")
		return
	}

	var req FundBountyRequest
	if err := decodeJSON(r, &req); err != nil || req.TxHash == "" {
		writeError(w, http.StatusBadRequest, "tx_hash is required")
		return
	}

	if err := h.Store.UpdateBountyStatus(r.Context(), id, models.CampaignFunded, &req.TxHash); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fund bounty")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "funded"})
}

func (h *BountyHandlers) ListDeals(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bounty id")
		return
	}

	deals, err := h.Store.ListDealsByBounty(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list claims")
		return
	}

	writeJSON(w, http.StatusOK, deals)
}

func (h *BountyHandlers) ListEligibleComments(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bounty id")
		return
	}

	// Only show eligible comments that belong to the current user's YouTube channel
	channel, err := h.Store.GetYouTubeChannelByUser(r.Context(), claims.UserID)
	if err != nil {
		// No linked channel: return empty list so they only see "their" comments (none)
		writeJSON(w, http.StatusOK, []models.ViralComment{})
		return
	}

	comments, err := h.Store.ListEligibleCommentsForBounty(r.Context(), id, &channel.ChannelID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list eligible comments")
		return
	}

	writeJSON(w, http.StatusOK, comments)
}

type ClaimBountyRequest struct {
	CommentID uuid.UUID `json:"comment_id"`
}

func (h *BountyHandlers) Claim(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	bountyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bounty id")
		return
	}

	var req ClaimBountyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	bounty, err := h.Store.GetBountyByID(r.Context(), bountyID)
	if err != nil {
		writeError(w, http.StatusNotFound, "bounty not found")
		return
	}
	if bounty.Status != models.CampaignFunded && bounty.Status != models.CampaignActive {
		writeError(w, http.StatusBadRequest, "bounty is not open for claims")
		return
	}

	comment, err := h.Store.GetViralCommentByID(r.Context(), req.CommentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "comment not found")
		return
	}
	if comment.LikeCount < bounty.MinLikes {
		writeError(w, http.StatusBadRequest, "comment does not meet min_likes requirement")
		return
	}

	ch, err := h.Store.GetYouTubeChannelByChannelID(r.Context(), comment.AuthorChannelID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "comment author not registered on platform")
		return
	}
	if ch.UserID != claims.UserID {
		writeError(w, http.StatusForbidden, "you can only claim bounties for your own comments")
		return
	}

	deal, err := h.Store.CreateBountyClaim(r.Context(), bountyID, req.CommentID, ch.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to claim bounty")
		return
	}

	writeJSON(w, http.StatusCreated, deal)
}
