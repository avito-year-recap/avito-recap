package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultShutdownTimeout = 10 * time.Second

const (
	StorageMemory     = "memory"
	StorageClickHouse = "clickhouse"

	NarrativeAuto   = "auto"
	NarrativeOllama = "ollama"
	NarrativeOff    = "off"
)

type Config struct {
	// API
	Address  string
	HTTPAddr string

	// Storage
	StorageBackend string
	ClickHouseDSN  string

	// Recap data
	SeedDemoData   bool
	ProfilesPath   string
	ScenariosPath  string
	StaticDir      string
	FrontendDir    string
	AllowedOrigins []string

	// Optional AI storytelling. Business decisions remain deterministic; this
	// provider may only rewrite presentation descriptions.
	NarrativeProvider         string
	OllamaModel               string
	OllamaBaseURL             string
	OllamaKeepAlive           string
	OllamaTimeout             time.Duration
	AINarrativeMaxConcurrency int

	// Graceful shutdown
	ShutdownTimeout time.Duration
}

func FromEnv() (Config, error) {
	timeout := defaultShutdownTimeout

	if raw := strings.TrimSpace(os.Getenv("SHUTDOWN_TIMEOUT")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 {
			return Config{}, fmt.Errorf("invalid SHUTDOWN_TIMEOUT %q", raw)
		}

		timeout = parsed
	}

	address := listenAddress()

	seedDemoData, err := envBool("SEED_DEMO_DATA", true)
	if err != nil {
		return Config{}, err
	}

	storageBackend := strings.ToLower(envOrDefault("STORAGE_BACKEND", StorageClickHouse))
	if storageBackend != StorageMemory && storageBackend != StorageClickHouse {
		return Config{}, fmt.Errorf(
			"invalid STORAGE_BACKEND %q: expected %q or %q",
			storageBackend,
			StorageMemory,
			StorageClickHouse,
		)
	}

	narrativeProvider := strings.ToLower(envOrDefault("AI_NARRATIVE_PROVIDER", NarrativeAuto))
	if narrativeProvider != NarrativeAuto && narrativeProvider != NarrativeOllama && narrativeProvider != NarrativeOff {
		return Config{}, fmt.Errorf(
			"invalid AI_NARRATIVE_PROVIDER %q: expected %q, %q or %q",
			narrativeProvider, NarrativeAuto, NarrativeOllama, NarrativeOff,
		)
	}
	ollamaTimeout := 20 * time.Second
	if raw := strings.TrimSpace(os.Getenv("AI_NARRATIVE_TIMEOUT")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 {
			return Config{}, fmt.Errorf("invalid AI_NARRATIVE_TIMEOUT %q", raw)
		}
		ollamaTimeout = parsed
	}
	aiNarrativeMaxConcurrency, err := envPositiveInt("AI_NARRATIVE_MAX_CONCURRENCY", 2)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Address:  address,
		HTTPAddr: address,

		StorageBackend: storageBackend,
		ClickHouseDSN: envOrDefault(
			"CLICKHOUSE_DSN",
			"clickhouse://recap:recap@clickhouse:9000/recap",
		),

		SeedDemoData: seedDemoData,

		ProfilesPath: envOrDefault(
			"PROFILES_PATH",
			"seeds/profiles.json",
		),

		ScenariosPath: envOrDefault(
			"SCENARIOS_PATH",
			"seeds/scenarios.json",
		),

		StaticDir: envOrDefault(
			"STATIC_DIR",
			"frontend/public",
		),

		FrontendDir: strings.TrimSpace(os.Getenv("FRONTEND_DIR")),

		AllowedOrigins: splitValues(
			envOrDefault(
				"CORS_ALLOWED_ORIGINS",
				"http://localhost:3000,http://localhost:5173",
			),
		),

		NarrativeProvider:         narrativeProvider,
		OllamaModel:               envOrDefault("OLLAMA_MODEL", "qwen3:4b"),
		OllamaBaseURL:             envOrDefault("OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaKeepAlive:           envOrDefault("OLLAMA_KEEP_ALIVE", "5m"),
		OllamaTimeout:             ollamaTimeout,
		AINarrativeMaxConcurrency: aiNarrativeMaxConcurrency,

		ShutdownTimeout: timeout,
	}, nil
}

func listenAddress() string {
	if value := strings.TrimSpace(os.Getenv("API_ADDRESS")); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("HTTP_ADDR")); value != "" {
		return value
	}
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		return ":" + port
	}
	return ":8080"
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}

func splitValues(raw string) []string {
	values := make([]string, 0)
	seen := make(map[string]struct{})

	for _, item := range strings.Split(raw, ",") {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}

		if _, exists := seen[value]; exists {
			continue
		}

		seen[value] = struct{}{}
		values = append(values, value)
	}

	return values
}

func envBool(key string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid %s %q", key, raw)
	}
}

func envPositiveInt(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid %s %q: expected positive integer", key, raw)
	}
	return value, nil
}
