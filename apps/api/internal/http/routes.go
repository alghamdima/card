package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"golang.org/x/time/rate"

	"cards-api/internal/config"
	"cards-api/internal/http/handlers"
	appmiddleware "cards-api/internal/http/middleware"
	"cards-api/internal/service"
)

type RouterParams struct {
	Config          *config.Config
	AuthService     *service.AuthService
	CampaignService *service.CampaignService
	CardService     *service.CardService
}

func SetupRoutes(params RouterParams) http.Handler {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(appmiddleware.RequestID)
	r.Use(appmiddleware.Logger)
	r.Use(appmiddleware.Recoverer)
	r.Use(chimiddleware.CleanPath)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   params.Config.TrustedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(params.AuthService)
	campaignHandler := handlers.NewCampaignHandler(params.CampaignService)
	cardHandler := handlers.NewCardHandler(params.CardService)

	// API v1 Sub-router
	r.Route("/api/v1", func(v1 chi.Router) {
		// Health check
		v1.Get("/health", healthHandler.Health)

		// Public Campaign & Card Endpoints
		v1.Get("/campaigns", campaignHandler.ListPublic)
		v1.Get("/campaigns/{slug}", campaignHandler.GetPublic)
		v1.Post("/campaigns/{slug}/cards", cardHandler.Create)

		// Auth endpoints (with rate limiting for login)
		v1.Route("/auth", func(auth chi.Router) {
			// Limit to 5 requests per second, burst 10
			auth.With(appmiddleware.RateLimit(rate.Every(12*1000*1000*1000), 5)).Post("/login", authHandler.Login)
			auth.Post("/logout", authHandler.Logout)
			auth.With(appmiddleware.RequireAuth(params.AuthService)).Get("/me", authHandler.Me)
		})

		// Protected Admin Routes
		v1.Route("/admin", func(admin chi.Router) {
			admin.Use(appmiddleware.RequireAuth(params.AuthService))

			// Dashboard stats
			admin.Get("/dashboard", cardHandler.DashboardStats)

			// Campaigns management
			admin.Get("/campaigns", campaignHandler.ListAdmin)
			admin.Post("/campaigns", campaignHandler.Create)
			admin.Get("/campaigns/{slug}", campaignHandler.GetAdmin)
			admin.Put("/campaigns/{slug}", campaignHandler.Update)
			admin.Delete("/campaigns/{slug}", campaignHandler.Delete)

			// Cards management
			admin.Get("/campaigns/{slug}/cards", cardHandler.ListByCampaign)
			admin.Get("/cards", cardHandler.ListAll)
		})
	})

	return r
}
