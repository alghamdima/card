package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"cards-api/internal/config"
	"cards-api/internal/database"
	internalhttp "cards-api/internal/http"
	"cards-api/internal/repository"
	"cards-api/internal/service"
)

func main() {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Fatal: failed to load configuration: %v", err)
	}

	// 2. Connect to SQLite
	db, err := database.Connect(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Fatal: failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Printf("Connected to SQLite database at: %s (WAL Mode)", cfg.DatabasePath)

	// 3. Run migrations
	migrationsDir := os.Getenv("MIGRATIONS_PATH")
	if migrationsDir == "" {
		migrationsDir = "./migrations"
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			// check relative to binary
			ex, _ := os.Executable()
			migrationsDir = filepath.Join(filepath.Dir(ex), "migrations")
		}
	}
	if err := database.RunMigrations(db, migrationsDir); err != nil {
		log.Printf("Warning: migrations returned: %v", err)
	}

	// 4. Repositories
	campaignRepo := repository.NewCampaignRepository(db)
	cardRepo := repository.NewCardRepository(db)

	// 5. Services
	authService := service.NewAuthService(cfg.AdminPassword, cfg.SessionSecret)
	campaignService := service.NewCampaignService(campaignRepo)
	cardService := service.NewCardService(cardRepo, campaignRepo)

	// 6. Router
	router := internalhttp.SetupRoutes(internalhttp.RouterParams{
		Config:          cfg,
		AuthService:     authService,
		CampaignService: campaignService,
		CardService:     cardService,
	})

	// 7. Server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 8. Graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Cards API server listening on http://0.0.0.0:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down Cards API server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced shutdown error: %v", err)
	}

	log.Println("Cards API server stopped.")
}
