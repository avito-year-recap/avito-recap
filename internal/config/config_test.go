package config

import (
	"testing"
	"time"
)

func TestFromEnvUsesDevelopmentDefaults(t *testing.T) {
	for _, key := range []string{
		"API_ADDRESS",
		"PROFILES_PATH",
		"SCENARIOS_PATH",
		"STATIC_DIR",
		"CORS_ALLOWED_ORIGINS",
		"SHUTDOWN_TIMEOUT",
	} {
		t.Setenv(key, "")
	}
	configured, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if configured.Address != ":8080" ||
		configured.ProfilesPath != "seeds/profiles.json" ||
		configured.ScenariosPath != "seeds/scenarios.json" ||
		configured.StaticDir != "static" {
		t.Fatalf("unexpected defaults: %+v", configured)
	}
	if configured.ShutdownTimeout != 10*time.Second {
		t.Fatalf("shutdown timeout = %s", configured.ShutdownTimeout)
	}
	if len(configured.AllowedOrigins) != 2 {
		t.Fatalf("allowed origins = %v", configured.AllowedOrigins)
	}
}

func TestFromEnvParsesOverrides(t *testing.T) {
	t.Setenv("API_ADDRESS", "127.0.0.1:9000")
	t.Setenv("PROFILES_PATH", "profiles.json")
	t.Setenv("SCENARIOS_PATH", "scenarios.json")
	t.Setenv("STATIC_DIR", "public")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://one.example, https://two.example,https://one.example")
	t.Setenv("SHUTDOWN_TIMEOUT", "3s")
	configured, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if configured.Address != "127.0.0.1:9000" ||
		configured.ProfilesPath != "profiles.json" ||
		configured.ScenariosPath != "scenarios.json" ||
		configured.StaticDir != "public" ||
		configured.ShutdownTimeout != 3*time.Second {
		t.Fatalf("unexpected config: %+v", configured)
	}
	if len(configured.AllowedOrigins) != 2 ||
		configured.AllowedOrigins[0] != "https://one.example" ||
		configured.AllowedOrigins[1] != "https://two.example" {
		t.Fatalf("allowed origins = %v", configured.AllowedOrigins)
	}
}

func TestFromEnvRejectsInvalidShutdownTimeout(t *testing.T) {
	t.Setenv("SHUTDOWN_TIMEOUT", "forever")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected invalid timeout error")
	}
}
