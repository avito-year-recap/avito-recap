package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const defaultShutdownTimeout = 10 * time.Second

type Config struct {
	Address         string
	ProfilesPath    string
	ScenariosPath   string
	StaticDir       string
	AllowedOrigins  []string
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
	return Config{
		Address:         envOrDefault("API_ADDRESS", ":8080"),
		ProfilesPath:    envOrDefault("PROFILES_PATH", "seeds/profiles.json"),
		ScenariosPath:   envOrDefault("SCENARIOS_PATH", "seeds/scenarios.json"),
		StaticDir:       envOrDefault("STATIC_DIR", "static"),
		AllowedOrigins:  splitValues(envOrDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")),
		ShutdownTimeout: timeout,
	}, nil
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
