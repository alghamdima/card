package config

import "testing"

func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for _, key := range []string{"APP_ENV", "ADMIN_PASSWORD", "SESSION_SECRET", "TRUSTED_ORIGINS", "TRUST_PROXY", "PORT"} {
		t.Setenv(key, "")
	}
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

func TestProductionRejectsDefaultCredentials(t *testing.T) {
	strongSecret := "0123456789abcdef0123456789abcdef"

	cases := map[string]map[string]string{
		"default password":  {"APP_ENV": "production", "SESSION_SECRET": strongSecret},
		"default secret":    {"APP_ENV": "production", "ADMIN_PASSWORD": "s3cure-pass"},
		"short secret":      {"APP_ENV": "production", "ADMIN_PASSWORD": "s3cure-pass", "SESSION_SECRET": "short"},
		"implicit env":      {},
		"explicit defaults": {"APP_ENV": "production", "ADMIN_PASSWORD": "admin", "SESSION_SECRET": defaultSessionSecret},
	}
	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			setEnv(t, env)
			if _, err := Load(); err == nil {
				t.Fatal("expected Load to refuse insecure production configuration")
			}
		})
	}

	t.Run("valid production config", func(t *testing.T) {
		setEnv(t, map[string]string{"APP_ENV": "production", "ADMIN_PASSWORD": "s3cure-pass", "SESSION_SECRET": strongSecret})
		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg.TrustedOrigins) != 0 || cfg.TrustProxy {
			t.Fatalf("CORS and proxy trust must be opt-in, got %+v", cfg)
		}
	})
}

func TestDevelopmentAllowsDefaultsAndParsesLists(t *testing.T) {
	setEnv(t, map[string]string{"APP_ENV": "development", "TRUSTED_ORIGINS": " http://a.test, ,http://b.test ", "TRUST_PROXY": "true"})
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.TrustedOrigins) != 2 || cfg.TrustedOrigins[1] != "http://b.test" || !cfg.TrustProxy {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestInvalidTrustProxyIsRejected(t *testing.T) {
	setEnv(t, map[string]string{"APP_ENV": "development", "TRUST_PROXY": "maybe"})
	if _, err := Load(); err == nil {
		t.Fatal("expected an error for a non-boolean TRUST_PROXY")
	}
}
