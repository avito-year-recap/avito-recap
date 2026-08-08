package config

import "os"

type Config struct {
	HTTPAddr      string
	ClickHouseDSN string
}

func Load() Config {
	return Config{
		HTTPAddr:      getEnv("HTTP_ADDR", ":8080"),
		ClickHouseDSN: getEnv("CLICKHOUSE_DSN", "clickhouse://recap:recap@clickhouse:9000/recap"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

