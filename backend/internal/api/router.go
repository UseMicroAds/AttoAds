package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/attoads/attoads-backend/internal/auth"
	"github.com/attoads/attoads-backend/internal/db"
	"github.com/attoads/attoads-backend/internal/models"
	ytclient "github.com/attoads/attoads-backend/internal/youtube"
	"golang.org/x/oauth2"
)

func NewRouter(store *db.Store, cfg *Config, oauthCfg *oauth2.Config, yt *ytclient.Client) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	authH := &AuthHandlers{Store: store, OAuthConfig: oauthCfg, JWTSecret: cfg.JWTSecret}
	campaignH := &CampaignHandlers{Store: store}
	bountyH := &BountyHandlers{Store: store}
	marketplaceH := &MarketplaceHandlers{Store: store, YTClient: yt}
	perfH := &PerformanceHandlers{Store: store}

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, map[string]string{"status": "ok"})
	})

	r.Route("/api", func(r chi.Router) {
		// Public auth routes
		r.Get("/auth/google", authH.GoogleLogin)
		r.Post("/auth/google/callback", authH.GoogleCallback)

		// Public marketplace read routes
		r.Get("/marketplace/comments", marketplaceH.ListComments)
		r.Get("/marketplace/comments/by-video/{videoID}/all", marketplaceH.ListAllCommentsByVideo)
		r.Get("/marketplace/comments/by-channel/{channelID}/all", marketplaceH.ListAllCommentsByAuthorChannel)
		r.Get("/marketplace/comments/by-channel/{channelID}", marketplaceH.ListCommentsByAuthorChannel)
		r.Get("/marketplace/comments/{id}", marketplaceH.GetComment)
		r.Get("/marketplace/videos", marketplaceH.ListTrendingVideos)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(cfg.JWTSecret))

			r.Get("/me", authH.GetMe)
			r.Post("/wallet/link", authH.LinkWallet)
			r.Delete("/wallet/link", authH.UnlinkWallet)
			r.Get("/bounties/{id}", bountyH.Get)
			r.Post("/marketplace/comments/register-test", marketplaceH.RegisterCommentForTesting)

			// Advertiser routes
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole(models.RoleAdvertiser))

				r.Post("/campaigns", campaignH.Create)
				r.Get("/campaigns", campaignH.List)
				r.Get("/campaigns/{id}", campaignH.Get)
				r.Post("/campaigns/{id}/fund", campaignH.Fund)
				r.Get("/campaigns/{id}/deals", campaignH.ListDeals)
				r.Post("/bounties", bountyH.Create)
				r.Get("/bounties", bountyH.List)
				r.Post("/bounties/{id}/fund", bountyH.Fund)
				r.Get("/bounties/{id}/deals", bountyH.ListDeals)
				r.Post("/deals", marketplaceH.CreateDeal)
				r.Get("/deals/performance", perfH.ListDeals)
				r.Get("/deals/{id}/performance", perfH.GetDealPerformance)
			})

			// Commenter routes
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole(models.RoleCommenter))

				r.Get("/my/comments", marketplaceH.ListMyComments)
				r.Get("/my/deals", marketplaceH.ListMyDeals)
				r.Get("/my/transactions", marketplaceH.ListMyTransactions)
				r.Get("/bounties/active", bountyH.ListActive)
				r.Get("/bounties/{id}/eligible-comments", bountyH.ListEligibleComments)
				r.Post("/bounties/{id}/claim", bountyH.Claim)
			})
		})
	})

	return r
}
