package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const defaultShutdownTimeout = 10 * time.Second

type Config struct {
	// API
	Address  string
	HTTPAddr string

	// ClickHouse
	ClickHouseDSN string

	// Recap data
	SeedDemoData   bool
	ProfilesPath   string
	ScenariosPath  string
	StaticDir      string
	AllowedOrigins []string

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

	address := envOrDefault(
		"API_ADDRESS",
		envOrDefault("HTTP_ADDR", ":8080"),
	)

	seedDemoData, err := envBool("SEED_DEMO_DATA", true)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Address:  address,
		HTTPAddr: address,

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
			"static",
		),

		AllowedOrigins: splitValues(
			envOrDefault(
				"CORS_ALLOWED_ORIGINS",
				"http://localhost:3000,http://localhost:5173",
			),
		),

		ShutdownTimeout: timeout,
	}, nil
}

// Оставляем для старого кода, который вызывает config.Load().
func Load() Config {
	cfg, err := FromEnv()
	if err != nil {
		panic(err)
	}

	return cfg
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
