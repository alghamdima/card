package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	_ "time/tzdata" // bundle the timezone database so APP_TIMEZONE works on minimal images

	"cards-api/internal/config"
	"cards-api/internal/database"
	internalhttp "cards-api/internal/http"
	"cards-api/internal/repository"
	"cards-api/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Fatal: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if !cfg.IsProduction() {
		log.Printf("Running in %q mode: insecure development defaults are allowed", cfg.AppEnv)
	}

	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return err
	}

	db, err := database.Connect(cfg.DatabasePath)
	if err != nil {
		return err
	}
	defer db.Close()
	log.Printf("Connected to SQLite database at %s (WAL mode)", cfg.DatabasePath)

	// A schema that failed to migrate must stop the boot: serving traffic on a
	// half-migrated database only turns into confusing runtime errors.
	if err := database.RunMigrations(db, migrationsDir(cfg.MigrationsPath)); err != nil {
		return err
	}

	campaignRepo := repository.NewCampaignRepository(db)
	cardRepo := repository.NewCardRepository(db)

	router := internalhttp.SetupRoutes(internalhttp.RouterParams{
		Config:          cfg,
		DB:              db,
		AuthService:     service.NewAuthService(cfg.AdminPassword, cfg.SessionSecret),
		CampaignService: service.NewCampaignService(campaignRepo),
		CardService:     service.NewCardService(cardRepo, campaignRepo, loc),
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second, // campaign uploads carry base64 artwork
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 16,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Cards API server listening on http://0.0.0.0:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return err
	case <-stop:
	}

	log.Println("Shutting down Cards API server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced shutdown: %v", err)
	}
	log.Println("Cards API server stopped.")
	return nil
}

// migrationsDir resolves the migration scripts: explicit config first, then
// ./migrations, then next to the executable.
func migrationsDir(configured string) string {
	if configured != "" {
		return configured
	}
	if _, err := os.Stat("./migrations"); err == nil {
		return "./migrations"
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), "migrations")
	}
	return "./migrations"
}
