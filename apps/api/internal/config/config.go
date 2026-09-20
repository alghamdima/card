package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AppEnv         string
	Port           string
	DatabasePath   string
	AdminPassword  string
	SessionSecret  string
	TrustedOrigins []string
}

func Load() (*Config, error) {
	appEnv := getEnv("APP_ENV", "production")
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DATABASE_PATH", "/data/cards.db")
	adminPassword := getEnv("ADMIN_PASSWORD", "admin")
	sessionSecret := getEnv("SESSION_SECRET", "cards-super-secret-key-prod-2026")

	originsRaw := getEnv("TRUSTED_ORIGINS", "*")
	var origins []string
	for _, o := range strings.Split(originsRaw, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			origins = append(origins, o)
		}
	}
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	if sessionSecret == "" {
		return nil, fmt.Errorf("SESSION_SECRET cannot be empty")
	}

	return &Config{
		AppEnv:         appEnv,
		Port:           port,
		DatabasePath:   dbPath,
		AdminPassword:  adminPassword,
		SessionSecret:  sessionSecret,
		TrustedOrigins: origins,
	}, nil
}

func getEnv(key, defaultVal string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultVal
	}
	return val
}
