package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/microads/microads-backend/internal/db"
	"github.com/microads/microads-backend/internal/discovery"
	"github.com/microads/microads-backend/internal/youtube"
)

func main() {
	godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	dbURL := envOrDefault("DATABASE_URL", "postgres://microads:microads_dev@localhost:5432/microads?sslmode=disable")
	redisURL := envOrDefault("REDIS_URL", "redis://localhost:6379/0")
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		slog.Error("YOUTUBE_API_KEY is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, err := db.NewStore(ctx, dbURL, redisURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	ytClient := youtube.NewClient(apiKey)

	threshold, _ := strconv.ParseFloat(envOrDefault("VELOCITY_THRESHOLD", "100"), 64)
	intervalMin, _ := strconv.Atoi(envOrDefault("POLL_INTERVAL_MIN", "5"))

	engine := discovery.NewEngine(store, ytClient, discovery.EngineConfig{
		RegionCode:        envOrDefault("REGION_CODE", "US"),
		MaxVideos:         10,
		MaxComments:       20,
		VelocityThreshold: threshold,
		PollInterval:      time.Duration(intervalMin) * time.Minute,
	})

	go engine.Run(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("discovery worker shutting down")
	cancel()
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
