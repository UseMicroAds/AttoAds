package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/attoads/attoads-backend/internal/auth"
	"github.com/attoads/attoads-backend/internal/db"
	"github.com/attoads/attoads-backend/internal/settlement"
	"github.com/attoads/attoads-backend/internal/verifier"
	"github.com/attoads/attoads-backend/internal/youtube"
)

func main() {
	godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	dbURL := envOrDefault("DATABASE_URL", "postgres://attoads:attoads_dev@localhost:5432/attoads?sslmode=disable")
	redisURL := envOrDefault("REDIS_URL", "redis://localhost:6379/0")
	apiKey := os.Getenv("YOUTUBE_API_KEY")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, err := db.NewStore(ctx, dbURL, redisURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	ytClient := youtube.NewClient(apiKey)
	oauthCfg := auth.NewGoogleOAuthConfig(
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_CLIENT_SECRET"),
		envOrDefault("GOOGLE_REDIRECT_URL", "http://localhost:3000/api/auth/callback/google"),
	)

	var settler *settlement.Settler
	privKey := os.Getenv("OPERATOR_PRIVATE_KEY")
	escrowAddr := os.Getenv("ESCROW_CONTRACT_ADDRESS")
	rpcURL := envOrDefault("BASE_RPC_URL", "https://sepolia.base.org")

	if privKey != "" && escrowAddr != "" {
		settler, err = settlement.NewSettler(rpcURL, privKey, escrowAddr, 84532) // Base Sepolia chain ID
		if err != nil {
			slog.Error("failed to create settler", "error", err)
			os.Exit(1)
		}
	} else {
		slog.Warn("settlement disabled: OPERATOR_PRIVATE_KEY or ESCROW_CONTRACT_ADDRESS not set")
	}

	w := verifier.NewWorker(store, ytClient, oauthCfg, settler, 2*time.Minute)
	go w.Run(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("verifier worker shutting down")
	cancel()
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
