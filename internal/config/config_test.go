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
		"STORAGE_BACKEND",
		"CLICKHOUSE_DSN",
		"PROFILES_PATH",
		"SCENARIOS_PATH",
		"STATIC_DIR",
		"FRONTEND_DIR",
		"CORS_ALLOWED_ORIGINS",
		"SHUTDOWN_TIMEOUT",
		"SEED_DEMO_DATA",
		"AI_NARRATIVE_PROVIDER",
		"AI_NARRATIVE_TIMEOUT",
		"AI_NARRATIVE_MAX_CONCURRENCY",
		"OLLAMA_MODEL",
		"OLLAMA_BASE_URL",
		"OLLAMA_KEEP_ALIVE",
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
		configured.StorageBackend != StorageClickHouse ||
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
	if configured.NarrativeProvider != NarrativeAuto || configured.OllamaModel != "qwen3:4b" || configured.OllamaBaseURL != "http://localhost:11434" || configured.OllamaKeepAlive != "5m" || configured.OllamaTimeout != 20*time.Second || configured.AINarrativeMaxConcurrency != 2 {
		t.Fatalf("unexpected AI defaults: %+v", configured)
	}
	if len(configured.AllowedOrigins) != 2 {
		t.Fatalf("allowed origins = %v", configured.AllowedOrigins)
	}
}

func TestFromEnvParsesOverrides(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("API_ADDRESS", "127.0.0.1:9000")
	t.Setenv("STORAGE_BACKEND", "memory")
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
		configured.StorageBackend != StorageMemory ||
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

func TestFromEnvRejectsInvalidStorageBackend(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("STORAGE_BACKEND", "sqlite")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected invalid STORAGE_BACKEND error")
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

func TestFromEnvParsesOllamaNarrativeConfig(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("AI_NARRATIVE_PROVIDER", "ollama")
	t.Setenv("OLLAMA_MODEL", "gemma3:4b")
	t.Setenv("OLLAMA_BASE_URL", "http://ollama.test:11434")
	t.Setenv("OLLAMA_KEEP_ALIVE", "10m")
	t.Setenv("AI_NARRATIVE_TIMEOUT", "12s")
	t.Setenv("AI_NARRATIVE_MAX_CONCURRENCY", "3")

	configured, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if configured.NarrativeProvider != NarrativeOllama || configured.OllamaModel != "gemma3:4b" || configured.OllamaBaseURL != "http://ollama.test:11434" || configured.OllamaKeepAlive != "10m" || configured.OllamaTimeout != 12*time.Second || configured.AINarrativeMaxConcurrency != 3 {
		t.Fatalf("unexpected AI config: %+v", configured)
	}
}

func TestFromEnvRejectsInvalidNarrativeProvider(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("AI_NARRATIVE_PROVIDER", "magic")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected invalid AI_NARRATIVE_PROVIDER error")
	}
}

func TestFromEnvRejectsInvalidNarrativeConcurrency(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("AI_NARRATIVE_MAX_CONCURRENCY", "0")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected invalid AI_NARRATIVE_MAX_CONCURRENCY error")
	}
}
