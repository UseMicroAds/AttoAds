package verifier

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"

	"github.com/microads/microads-backend/internal/db"
	"github.com/microads/microads-backend/internal/models"
	"github.com/microads/microads-backend/internal/settlement"
	ytclient "github.com/microads/microads-backend/internal/youtube"
)

type Worker struct {
	store        *db.Store
	ytClient     *ytclient.Client
	oauthCfg     *oauth2.Config
	settler      *settlement.Settler
	pollInterval time.Duration
}

func NewWorker(store *db.Store, ytClient *ytclient.Client, oauthCfg *oauth2.Config, settler *settlement.Settler, pollInterval time.Duration) *Worker {
	return &Worker{
		store:        store,
		ytClient:     ytClient,
		oauthCfg:     oauthCfg,
		settler:      settler,
		pollInterval: pollInterval,
	}
}

func (w *Worker) Run(ctx context.Context) {
	slog.Info("verification worker starting", "interval", w.pollInterval)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("verification worker shutting down")
			return
		case <-ticker.C:
			w.reconcileTransactions(ctx)
			w.checkDeals(ctx)
		}
	}
}

func (w *Worker) checkDeals(ctx context.Context) {
	deals, err := w.store.ListPendingDeals(ctx)
	if err != nil {
		slog.Error("verifier: failed to list pending deals", "error", err)
		return
	}

	for _, deal := range deals {
		if deal.Status == models.DealPending {
			w.editComment(ctx, deal)
		} else if deal.Status == models.DealEditPending {
			w.verifyComment(ctx, deal)
		} else if deal.Status == models.DealVerified {
			w.tryPayout(ctx, deal)
		}
	}
}

func (w *Worker) editComment(ctx context.Context, deal models.Deal) {
	campaign, err := w.store.GetCampaignByID(ctx, deal.CampaignID)
	if err != nil {
		slog.Error("verifier: failed to get campaign", "deal_id", deal.ID, "error", err)
		return
	}

	comment, err := w.store.GetViralCommentByID(ctx, deal.CommentID)
	if err != nil {
		slog.Error("verifier: failed to get comment", "deal_id", deal.ID, "error", err)
		return
	}

	ytCh, err := w.store.GetYouTubeChannelByChannelID(ctx, comment.AuthorChannelID)
	if err != nil {
		slog.Error("verifier: commenter has no youtube channel", "deal_id", deal.ID, "error", err)
		return
	}

	token := &oauth2.Token{
		AccessToken:  ytCh.AccessToken,
		RefreshToken: ytCh.RefreshToken,
		Expiry:       ytCh.TokenExpiry,
	}

	newText := comment.OriginalText + "\n\n" + campaign.AdText

	if err := w.ytClient.UpdateComment(ctx, token, w.oauthCfg, comment.CommentID, newText); err != nil {
		slog.Error("verifier: failed to edit comment", "deal_id", deal.ID, "error", err)
		_ = w.store.UpdateDealStatus(ctx, deal.ID, models.DealFailed)
		return
	}

	slog.Info("verifier: comment edited", "deal_id", deal.ID, "comment_id", comment.CommentID)
	_ = w.store.UpdateDealStatus(ctx, deal.ID, models.DealEditPending)
}

func (w *Worker) verifyComment(ctx context.Context, deal models.Deal) {
	campaign, err := w.store.GetCampaignByID(ctx, deal.CampaignID)
	if err != nil {
		slog.Error("verifier: failed to get campaign", "deal_id", deal.ID, "error", err)
		return
	}

	comment, err := w.store.GetViralCommentByID(ctx, deal.CommentID)
	if err != nil {
		slog.Error("verifier: failed to get comment", "deal_id", deal.ID, "error", err)
		return
	}

	currentText, err := w.ytClient.FetchCommentText(ctx, comment.CommentID)
	if err != nil {
		slog.Error("verifier: failed to fetch comment text", "deal_id", deal.ID, "error", err)
		return
	}

	if !strings.Contains(currentText, campaign.AdText) {
		slog.Warn("verifier: ad text not found in comment", "deal_id", deal.ID)
		return
	}

	slog.Info("verifier: comment verified", "deal_id", deal.ID)
	_ = w.store.UpdateDealStatus(ctx, deal.ID, models.DealVerified)
	w.tryPayout(ctx, deal)
}

func (w *Worker) tryPayout(ctx context.Context, deal models.Deal) {
	campaign, err := w.store.GetCampaignByID(ctx, deal.CampaignID)
	if err != nil {
		slog.Error("verifier: failed to get campaign for payout", "deal_id", deal.ID, "error", err)
		return
	}

	existingTx, err := w.store.GetTransactionByDeal(ctx, deal.ID)
	if err == nil {
		if existingTx.Status == models.TxConfirmed {
			_ = w.store.UpdateDealStatus(ctx, deal.ID, models.DealPaid)
			slog.Info("verifier: deal already paid", "deal_id", deal.ID, "tx_hash", existingTx.TxHash)
			return
		}
		if existingTx.Status == models.TxPending {
			slog.Info("verifier: payout already submitted, waiting for confirmation", "deal_id", deal.ID, "tx_hash", existingTx.TxHash)
			return
		}
		slog.Warn("verifier: last payout transaction failed, retrying", "deal_id", deal.ID, "tx_hash", existingTx.TxHash)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		slog.Error("verifier: failed to check existing transaction", "deal_id", deal.ID, "error", err)
		return
	}

	wallet, err := w.store.GetWalletByUser(ctx, deal.CommenterID)
	if err != nil {
		slog.Error("verifier: commenter has no wallet", "deal_id", deal.ID, "error", err)
		return
	}

	if w.settler == nil {
		slog.Warn("verifier: settlement disabled, cannot payout", "deal_id", deal.ID)
		return
	}

	txHash, err := w.settler.Release(ctx, deal.ID.String(), deal.CampaignID.String(), wallet.Address, campaign.PricePerPlacement)
	if err != nil {
		slog.Error("verifier: failed to release funds", "deal_id", deal.ID, "error", err)
		return
	}

	_, err = w.store.UpsertTransaction(ctx, deal.ID, txHash, campaign.PricePerPlacement, models.TxPending)
	if err != nil {
		slog.Error("verifier: failed to record transaction", "deal_id", deal.ID, "error", err)
		return
	}

	slog.Info("verifier: payout submitted", "deal_id", deal.ID, "tx_hash", txHash)
}

func (w *Worker) reconcileTransactions(ctx context.Context) {
	if w.settler == nil {
		return
	}

	txs, err := w.store.ListPendingTransactions(ctx, 200)
	if err != nil {
		slog.Error("verifier: failed to list pending transactions", "error", err)
		return
	}

	for _, tx := range txs {
		mined, success, err := w.settler.TransactionReceiptStatus(ctx, tx.TxHash)
		if err != nil {
			slog.Error("verifier: failed to check tx receipt", "tx_id", tx.ID, "tx_hash", tx.TxHash, "error", err)
			continue
		}
		if !mined {
			continue
		}

		if success {
			_ = w.store.UpdateTransactionStatus(ctx, tx.ID, models.TxConfirmed)
			_ = w.store.UpdateDealStatus(ctx, tx.DealID, models.DealPaid)
			slog.Info("verifier: payout confirmed", "tx_id", tx.ID, "tx_hash", tx.TxHash, "deal_id", tx.DealID)
			continue
		}

		_ = w.store.UpdateTransactionStatus(ctx, tx.ID, models.TxFailed)
		_ = w.store.UpdateDealStatus(ctx, tx.DealID, models.DealVerified)
		slog.Warn("verifier: payout transaction failed on-chain, will retry", "tx_id", tx.ID, "tx_hash", tx.TxHash, "deal_id", tx.DealID)
	}
}

