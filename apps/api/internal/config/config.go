package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultAdminPassword = "admin"
	defaultSessionSecret = "cards-super-secret-key-prod-2026"
	minSecretLength      = 32
)

type Config struct {
	AppEnv         string
	Port           string
	DatabasePath   string
	MigrationsPath string
	AdminPassword  string
	SessionSecret  string
	Timezone       string
	// TrustedOrigins enables CORS for the listed origins. Empty means CORS is
	// disabled, which is the right default: the SPA and API share an origin
	// behind the reverse proxy.
	TrustedOrigins []string
	// TrustProxy makes the API read the client IP from X-Real-IP. Only enable
	// it when the API is reachable exclusively through a trusted reverse proxy,
	// otherwise clients can spoof their address and bypass rate limits.
	TrustProxy bool
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:         strings.ToLower(getEnv("APP_ENV", "production")),
		Port:           getEnv("PORT", "8080"),
		DatabasePath:   getEnv("DATABASE_PATH", "/data/cards.db"),
		MigrationsPath: getEnv("MIGRATIONS_PATH", ""),
		AdminPassword:  getEnv("ADMIN_PASSWORD", defaultAdminPassword),
		SessionSecret:  getEnv("SESSION_SECRET", defaultSessionSecret),
		Timezone:       getEnv("APP_TIMEZONE", "Asia/Riyadh"),
		TrustedOrigins: splitList(os.Getenv("TRUSTED_ORIGINS")),
	}

	trustProxy, err := strconv.ParseBool(getEnv("TRUST_PROXY", "false"))
	if err != nil {
		return nil, fmt.Errorf("TRUST_PROXY must be a boolean: %w", err)
	}
	cfg.TrustProxy = trustProxy

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate refuses to boot production with the well-known development
// credentials: an admin panel protected by "admin" is not protected at all.
func (c *Config) validate() error {
	if c.IsProduction() {
		if c.AdminPassword == defaultAdminPassword {
			return fmt.Errorf("ADMIN_PASSWORD must be changed from the default value in production")
		}
		if c.SessionSecret == defaultSessionSecret || len(c.SessionSecret) < minSecretLength {
			return fmt.Errorf("SESSION_SECRET must be a unique random value of at least %d characters in production", minSecretLength)
		}
	}
	return nil
}

func splitList(raw string) []string {
	var out []string
	for _, item := range strings.Split(raw, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}

func getEnv(key, defaultVal string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultVal
	}
	return val
}
