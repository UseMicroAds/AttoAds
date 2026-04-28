package api

import (
	"fmt"
	"os"
)

type Config struct {
	Port            string
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	GoogleClientID  string
	GoogleSecret    string
	GoogleRedirect  string
	YouTubeAPIKey   string
	FrontendURL     string
	OperatorPrivKey string
	EscrowContract  string
	BaseRPCURL      string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://attoads:attoads_dev@localhost:5432/attoads?sslmode=disable"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		GoogleClientID:  getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleSecret:    getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirect:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:3000/api/auth/callback/google"),
		YouTubeAPIKey:   getEnv("YOUTUBE_API_KEY", ""),
		FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:3000"),
		OperatorPrivKey: getEnv("OPERATOR_PRIVATE_KEY", ""),
		EscrowContract:  getEnv("ESCROW_CONTRACT_ADDRESS", ""),
		BaseRPCURL:      getEnv("BASE_RPC_URL", "https://sepolia.base.org"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
