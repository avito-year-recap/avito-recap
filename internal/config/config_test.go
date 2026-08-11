package config

import (
	"testing"
	"time"
)

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"API_ADDRESS",
		"HTTP_ADDR",
		"PORT",
		"CLICKHOUSE_DSN",
		"PROFILES_PATH",
		"SCENARIOS_PATH",
		"STATIC_DIR",
		"FRONTEND_DIR",
		"CORS_ALLOWED_ORIGINS",
		"SHUTDOWN_TIMEOUT",
		"SEED_DEMO_DATA",
	} {
		t.Setenv(key, "")
	}
}

func TestFromEnvUsesDevelopmentDefaults(t *testing.T) {
	clearConfigEnv(t)

	configured, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if configured.Address != ":8080" ||
		configured.ProfilesPath != "seeds/profiles.json" ||
		configured.ScenariosPath != "seeds/scenarios.json" ||
		configured.StaticDir != "frontend/public" ||
		configured.FrontendDir != "" ||
		!configured.SeedDemoData {
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
	clearConfigEnv(t)
	t.Setenv("API_ADDRESS", "127.0.0.1:9000")
	t.Setenv("PROFILES_PATH", "profiles.json")
	t.Setenv("SCENARIOS_PATH", "scenarios.json")
	t.Setenv("STATIC_DIR", "public")
	t.Setenv("FRONTEND_DIR", "web")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://one.example, https://two.example,https://one.example")
	t.Setenv("SHUTDOWN_TIMEOUT", "3s")
	t.Setenv("SEED_DEMO_DATA", "false")
	configured, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if configured.Address != "127.0.0.1:9000" ||
		configured.ProfilesPath != "profiles.json" ||
		configured.ScenariosPath != "scenarios.json" ||
		configured.StaticDir != "public" ||
		configured.FrontendDir != "web" ||
		configured.ShutdownTimeout != 3*time.Second ||
		configured.SeedDemoData {
		t.Fatalf("unexpected config: %+v", configured)
	}
	if len(configured.AllowedOrigins) != 2 ||
		configured.AllowedOrigins[0] != "https://one.example" ||
		configured.AllowedOrigins[1] != "https://two.example" {
		t.Fatalf("allowed origins = %v", configured.AllowedOrigins)
	}
}

func TestFromEnvUsesRenderPort(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("PORT", "10000")

	configured, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if configured.HTTPAddr != ":10000" {
		t.Fatalf("HTTPAddr = %q, want :10000", configured.HTTPAddr)
	}
}

func TestFromEnvExplicitAddressWinsOverRenderPort(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("PORT", "10000")
	t.Setenv("HTTP_ADDR", "127.0.0.1:9000")

	configured, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if configured.HTTPAddr != "127.0.0.1:9000" {
		t.Fatalf("HTTPAddr = %q", configured.HTTPAddr)
	}
}

func TestFromEnvRejectsInvalidShutdownTimeout(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("SHUTDOWN_TIMEOUT", "forever")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected invalid timeout error")
	}
}

func TestFromEnvRejectsInvalidSeedDemoData(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("SEED_DEMO_DATA", "maybe")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected invalid SEED_DEMO_DATA error")
	}
}
