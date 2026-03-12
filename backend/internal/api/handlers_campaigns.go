package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/microads/microads-backend/internal/auth"
	"github.com/microads/microads-backend/internal/db"
	"github.com/microads/microads-backend/internal/models"
)

type CampaignHandlers struct {
	Store *db.Store
}

type CreateCampaignRequest struct {
	AdText            string `json:"ad_text"`
	BudgetCents       int    `json:"budget_cents"`
	PricePerPlacement int    `json:"price_per_placement"`
}

func (h *CampaignHandlers) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())

	var req CreateCampaignRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.AdText == "" || req.BudgetCents <= 0 || req.PricePerPlacement <= 0 {
		writeError(w, http.StatusBadRequest, "ad_text, budget_cents, and price_per_placement are required")
		return
	}

	campaign, err := h.Store.CreateCampaign(r.Context(), claims.UserID, req.AdText, req.BudgetCents, req.PricePerPlacement)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create campaign")
		return
	}

	writeJSON(w, http.StatusCreated, campaign)
}

func (h *CampaignHandlers) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())

	campaigns, err := h.Store.ListCampaignsByAdvertiser(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list campaigns")
		return
	}

	writeJSON(w, http.StatusOK, campaigns)
}

func (h *CampaignHandlers) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	campaign, err := h.Store.GetCampaignByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "campaign not found")
		return
	}

	writeJSON(w, http.StatusOK, campaign)
}

type FundCampaignRequest struct {
	TxHash string `json:"tx_hash"`
}

func (h *CampaignHandlers) Fund(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	campaign, err := h.Store.GetCampaignByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "campaign not found")
		return
	}
	if campaign.AdvertiserID != claims.UserID {
		writeError(w, http.StatusForbidden, "not your campaign")
		return
	}
	if campaign.Status != models.CampaignDraft {
		writeError(w, http.StatusBadRequest, "campaign already funded")
		return
	}

	var req FundCampaignRequest
	if err := decodeJSON(r, &req); err != nil || req.TxHash == "" {
		writeError(w, http.StatusBadRequest, "tx_hash is required")
		return
	}

	if err := h.Store.UpdateCampaignStatus(r.Context(), id, models.CampaignFunded, &req.TxHash); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fund campaign")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "funded"})
}

func (h *CampaignHandlers) ListDeals(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	deals, err := h.Store.ListDealsByCampaign(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list deals")
		return
	}

	writeJSON(w, http.StatusOK, deals)
}
