// takeout runs a once-daily job to sync commenter comment export.
// 1) Data Portability API for channels that granted it (EEA/UK/CH).
// 2) Process user-uploaded Takeout ZIPs when Data Portability is not available.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/microads/microads-backend/internal/auth"
	"github.com/microads/microads-backend/internal/dataportability"
	"github.com/microads/microads-backend/internal/db"
	"github.com/microads/microads-backend/internal/models"
	"github.com/microads/microads-backend/internal/takeout"
	"golang.org/x/oauth2"
)

const (
	runInterval   = 24 * time.Hour
	exportDays    = 30
	pollInterval  = 20 * time.Second
	exportTimeout = 15 * time.Minute
)

func main() {
	godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	dbURL := envOrDefault("DATABASE_URL", "postgres://microads:microads_dev@localhost:5432/microads?sslmode=disable")
	redisURL := envOrDefault("REDIS_URL", "redis://localhost:6379/0")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, err := db.NewStore(ctx, dbURL, redisURL)
	if err != nil {
		slog.Error("takeout cron: failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	run := func() {
		runExportSync(ctx, store)
	}

	run()

	ticker := time.NewTicker(runInterval)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("takeout cron: shutting down")
	cancel()
}

func runExportSync(ctx context.Context, store *db.Store) {
	slog.Info("takeout cron: running daily comment export sync")

	// 1) Data Portability API for channels that granted it (EEA/UK/CH).
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	dpRedirect := envOrDefault("GOOGLE_DP_REDIRECT_URL", "http://localhost:8080/api/auth/dataportability/callback")
	if clientID != "" && clientSecret != "" {
		dpOAuth := auth.NewDataPortabilityOAuthConfig(clientID, clientSecret, dpRedirect)
		channels, err := store.ListYouTubeChannelsWithDPToken(ctx)
		if err != nil {
			slog.Error("takeout cron: list channels failed", "error", err)
		} else if len(channels) > 0 {
			slog.Info("takeout cron: syncing Data Portability export for channels", "count", len(channels))
			since := time.Now().AddDate(0, 0, -exportDays)
			end := time.Now()
			for _, ch := range channels {
				syncOneChannelDP(ctx, store, dpOAuth, ch, since, end)
			}
		} else {
			slog.Info("takeout cron: no channels with Data Portability token")
		}
	} else {
		slog.Info("takeout cron: Google OAuth not configured, skipping Data Portability export")
	}

	// 2) Process user-uploaded Takeout ZIPs (when Data Portability is not available).
	processPendingTakeoutUploads(ctx, store)
}

func syncOneChannelDP(ctx context.Context, store *db.Store, dpOAuth *oauth2.Config, ch *models.YouTubeChannel, since, end time.Time) {
	channelID := ch.ChannelID
	exp := time.Time{}
	if ch.DPTokenExpiry != nil {
		exp = *ch.DPTokenExpiry
	}
	accessTok := ""
	if ch.DPAccessToken != nil {
		accessTok = *ch.DPAccessToken
	}
	refreshTok := ""
	if ch.DPRefreshToken != nil {
		refreshTok = *ch.DPRefreshToken
	}

	token := &oauth2.Token{
		AccessToken:  accessTok,
		RefreshToken: refreshTok,
		Expiry:       exp,
	}
	tokenSource := dpOAuth.TokenSource(ctx, token)
	client := dataportability.NewClient(tokenSource)

	jobID, err := client.InitiateExport(ctx, since, end)
	if err != nil {
		slog.Warn("takeout cron: initiate export failed", "channel_id", channelID, "error", err)
		return
	}

	slog.Info("takeout cron: export started", "channel_id", channelID, "job_id", jobID)

	deadline := time.Now().Add(exportTimeout)
	var state *dataportability.ArchiveState
	for time.Now().Before(deadline) {
		state, err = client.GetArchiveState(ctx, jobID)
		if err != nil {
			slog.Warn("takeout cron: get state failed", "channel_id", channelID, "error", err)
			time.Sleep(pollInterval)
			continue
		}
		switch state.State {
		case dataportability.StateComplete:
			goto download
		case dataportability.StateFailed, dataportability.StateCancelled:
			slog.Warn("takeout cron: export job ended", "channel_id", channelID, "state", state.State)
			return
		default:
			time.Sleep(pollInterval)
		}
	}
	slog.Warn("takeout cron: export timeout", "channel_id", channelID)
	return

download:
	if len(state.URLs) == 0 {
		slog.Info("takeout cron: export complete but no URLs", "channel_id", channelID)
		return
	}
	comments, err := client.DownloadAndParseComments(ctx, state.URLs[0])
	if err != nil {
		slog.Warn("takeout cron: download/parse failed", "channel_id", channelID, "error", err)
		return
	}

	rows := make([]models.CommentExportRow, len(comments))
	for i := range comments {
		rows[i] = models.CommentExportRow{
			CommentID:   comments[i].CommentID,
			VideoID:     comments[i].VideoID,
			TextDisplay: comments[i].TextDisplay,
			LikeCount:   comments[i].LikeCount,
			PublishedAt: comments[i].PublishedAt,
		}
	}
	if err := store.ReplaceCommentExportForChannel(ctx, channelID, rows); err != nil {
		slog.Warn("takeout cron: save export failed", "channel_id", channelID, "error", err)
		return
	}
	slog.Info("takeout cron: export cached", "channel_id", channelID, "comment_count", len(rows))
}

func processPendingTakeoutUploads(ctx context.Context, store *db.Store) {
	uploads, err := store.ListPendingTakeoutUploads(ctx)
	if err != nil {
		slog.Error("takeout cron: list pending uploads failed", "error", err)
		return
	}
	if len(uploads) == 0 {
		return
	}
	slog.Info("takeout cron: processing Takeout uploads", "count", len(uploads))
	for _, u := range uploads {
		comments, err := takeout.ParseTakeoutZip(u.FilePath)
		if err != nil {
			slog.Warn("takeout cron: parse upload failed", "upload_id", u.ID, "channel_id", u.ChannelID, "error", err)
			_ = store.MarkTakeoutUploadFailed(ctx, u.ID, err.Error())
			continue
		}
		if err := store.ReplaceCommentExportForChannel(ctx, u.ChannelID, comments); err != nil {
			slog.Warn("takeout cron: save Takeout export failed", "upload_id", u.ID, "channel_id", u.ChannelID, "error", err)
			_ = store.MarkTakeoutUploadFailed(ctx, u.ID, err.Error())
			continue
		}
		if err := store.MarkTakeoutUploadProcessed(ctx, u.ID); err != nil {
			slog.Warn("takeout cron: mark processed failed", "upload_id", u.ID, "error", err)
			continue
		}
		slog.Info("takeout cron: Takeout upload processed", "upload_id", u.ID, "channel_id", u.ChannelID, "comment_count", len(comments))
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
