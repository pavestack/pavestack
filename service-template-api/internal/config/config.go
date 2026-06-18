package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServiceName  string
	ListenAddr   string
	LogLevel     string
	OTLPEndpoint string
	Ready        bool
}

func Load() Config {
	return Config{
		ServiceName:  env("SERVICE_NAME", "service-template-api"),
		ListenAddr:   env("LISTEN_ADDR", ":8080"),
		LogLevel:     env("LOG_LEVEL", "info"),
		OTLPEndpoint: env("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		Ready:        envBool("READY", true),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
