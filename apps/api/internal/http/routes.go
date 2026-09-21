package http

import (
	"database/sql"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"golang.org/x/time/rate"

	"cards-api/internal/config"
	"cards-api/internal/http/handlers"
	appmiddleware "cards-api/internal/http/middleware"
	"cards-api/internal/http/response"
	"cards-api/internal/service"
)

type RouterParams struct {
	Config          *config.Config
	DB              *sql.DB
	AuthService     *service.AuthService
	CampaignService *service.CampaignService
	CardService     *service.CardService
}

func SetupRoutes(params RouterParams) http.Handler {
	r := chi.NewRouter()

	r.Use(appmiddleware.RequestID)
	r.Use(appmiddleware.Logger)
	r.Use(appmiddleware.Recoverer)
	r.Use(appmiddleware.SecurityHeaders)
	r.Use(chimiddleware.CleanPath)

	// The SPA and the API share an origin behind the reverse proxy, so CORS is
	// only enabled when origins are configured explicitly.
	if origins := params.Config.TrustedOrigins; len(origins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins: origins,
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
			ExposedHeaders: []string{"X-Request-ID"},
			// Credentials are incompatible with a wildcard origin.
			AllowCredentials: !slices.Contains(origins, "*"),
			MaxAge:           300,
		}))
	}

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	})

	healthHandler := handlers.NewHealthHandler(params.DB)
	authHandler := handlers.NewAuthHandler(params.AuthService)
	campaignHandler := handlers.NewCampaignHandler(params.CampaignService)
	cardHandler := handlers.NewCardHandler(params.CardService)

	trustProxy := params.Config.TrustProxy
	// Login: 5 attempts up front, then one every 12 seconds per client.
	loginLimiter := appmiddleware.NewRateLimiter(rate.Every(12*time.Second), 5, trustProxy)
	// Card submissions are public: 30 in a burst, then one every 2 seconds per client.
	cardLimiter := appmiddleware.NewRateLimiter(rate.Every(2*time.Second), 30, trustProxy)

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/health", healthHandler.Health)

		// Public campaign & card endpoints
		v1.Get("/campaigns", campaignHandler.ListPublic)
		v1.Get("/campaigns/{slug}", campaignHandler.GetPublic)
		v1.With(cardLimiter.Middleware).Post("/campaigns/{slug}/cards", cardHandler.Create)

		v1.Route("/auth", func(auth chi.Router) {
			auth.Use(appmiddleware.NoStore)
			auth.With(loginLimiter.Middleware).Post("/login", authHandler.Login)
			auth.Post("/logout", authHandler.Logout)
			auth.With(appmiddleware.RequireAuth(params.AuthService)).Get("/me", authHandler.Me)
		})

		v1.Route("/admin", func(admin chi.Router) {
			admin.Use(appmiddleware.NoStore)
			admin.Use(appmiddleware.RequireAuth(params.AuthService))

			admin.Get("/dashboard", cardHandler.DashboardStats)
			admin.Get("/analytics/overview", cardHandler.AnalyticsOverview)

			admin.Get("/campaigns", campaignHandler.ListAdmin)
			admin.Post("/campaigns", campaignHandler.Create)
			admin.Get("/campaigns/{slug}", campaignHandler.GetAdmin)
			admin.Put("/campaigns/{slug}", campaignHandler.Update)
			admin.Delete("/campaigns/{slug}", campaignHandler.Delete)

			admin.Get("/campaigns/{slug}/cards", cardHandler.ListByCampaign)
			admin.Get("/campaigns/{slug}/export", cardHandler.ExportCampaignCSV)
			admin.Get("/cards", cardHandler.ListAll)
		})
	})

	return r
}
