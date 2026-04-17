package config

import "os"

type Config struct {
	AppName              string
	HTTPPort             string
	LogLevel             string
	OTLPExporterEndpoint string
}

func LoadFromEnv() Config {
	return Config{
		AppName:              envOrDefault("APP_NAME", "local-observability-app"),
		HTTPPort:             envOrDefault("HTTP_PORT", "8080"),
		LogLevel:             envOrDefault("LOG_LEVEL", "info"),
		OTLPExporterEndpoint: envOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4317"),
	}
}

func envOrDefault(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return defaultValue
}
