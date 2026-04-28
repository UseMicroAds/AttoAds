package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/attoads/attoads-backend/internal/auth"
	"github.com/attoads/attoads-backend/internal/db"
	"github.com/attoads/attoads-backend/internal/models"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

type AuthHandlers struct {
	Store       *db.Store
	OAuthConfig *oauth2.Config
	JWTSecret   string
}

func (h *AuthHandlers) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	url := h.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

type OAuthCallbackRequest struct {
	Code string          `json:"code"`
	Role models.UserRole `json:"role"`
}

func (h *AuthHandlers) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	var req OAuthCallbackRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "code is required")
		return
	}
	if req.Role != models.RoleCommenter && req.Role != models.RoleAdvertiser {
		writeError(w, http.StatusBadRequest, "role must be commenter or advertiser")
		return
	}

	token, err := h.OAuthConfig.Exchange(r.Context(), req.Code)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to exchange code")
		return
	}

	ytService, err := yt.NewService(r.Context(), option.WithTokenSource(h.OAuthConfig.TokenSource(r.Context(), token)))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create youtube service")
		return
	}

	channelResp, err := ytService.Channels.List([]string{"snippet"}).Mine(true).Do()
	if err != nil || len(channelResp.Items) == 0 {
		writeError(w, http.StatusBadRequest, "failed to fetch youtube channel")
		return
	}

	channel := channelResp.Items[0]

	idTokenRaw, ok := token.Extra("id_token").(string)
	if !ok {
		writeError(w, http.StatusInternalServerError, "missing id_token")
		return
	}

	claims, err := parseGoogleIDToken(idTokenRaw)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid id_token")
		return
	}
	email, _ := claims["email"].(string)
	if email == "" {
		writeError(w, http.StatusBadRequest, "email not found in token")
		return
	}

	user, err := h.Store.GetUserByEmail(r.Context(), email)
	if errors.Is(err, pgx.ErrNoRows) {
		user, err = h.Store.CreateUser(r.Context(), email, req.Role)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create user")
			return
		}
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	_, err = h.Store.UpsertYouTubeChannel(r.Context(),
		user.ID, channel.Id, channel.Snippet.Title,
		token.AccessToken, token.RefreshToken, token.Expiry,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save channel")
		return
	}

	jwtToken, err := auth.GenerateToken(h.JWTSecret, user.ID, user.Email, user.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": jwtToken,
		"user":  user,
	})
}

type LinkWalletRequest struct {
	Address string `json:"address"`
}

func (h *AuthHandlers) LinkWallet(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req LinkWalletRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Address == "" {
		writeError(w, http.StatusBadRequest, "address is required")
		return
	}

	wallet, err := h.Store.UpsertWallet(r.Context(), claims.UserID, req.Address, "base")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to link wallet")
		return
	}

	writeJSON(w, http.StatusOK, wallet)
}

func (h *AuthHandlers) UnlinkWallet(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.Store.DeleteWalletByUser(r.Context(), claims.UserID); err != nil {
		slog.Error("wallet unlink failed", "user_id", claims.UserID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to unlink wallet")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "wallet unlinked"})
}

func (h *AuthHandlers) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.Store.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	wallet, _ := h.Store.GetWalletByUser(r.Context(), claims.UserID)
	ytChannel, _ := h.Store.GetYouTubeChannelByUser(r.Context(), claims.UserID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":    user,
		"wallet":  wallet,
		"channel": ytChannel,
	})
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// MVP: trust the id_token from a direct token exchange without JWKS verification
func parseGoogleIDToken(idToken string) (map[string]interface{}, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid jwt format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}
