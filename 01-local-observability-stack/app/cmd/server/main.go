package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/singl3focus/go-otel-workshop/01-local-observability-stack/app/internal/config"
	"github.com/singl3focus/go-otel-workshop/01-local-observability-stack/app/internal/observability"
	serverpkg "github.com/singl3focus/go-otel-workshop/01-local-observability-stack/app/internal/server"
)

func main() {
	cfg := config.LoadFromEnv()
	logger := newLogger(cfg.LogLevel)

	otelShutdown, err := observability.Setup(context.Background(), cfg.AppName, cfg.OTLPExporterEndpoint)
	if err != nil {
		logger.Error("failed to initialize OpenTelemetry", "error", err)
		return
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := otelShutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown OpenTelemetry", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := serverpkg.New(cfg.AppName, cfg.HTTPPort, logger)

	if err := server.Run(ctx); err != nil {
		logger.Error("server failed", "error", err)
		return
	}
}

func newLogger(level string) *slog.Logger {
	lvl := new(slog.LevelVar)
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		lvl.Set(slog.LevelDebug)
	case "warn":
		lvl.Set(slog.LevelWarn)
	case "error":
		lvl.Set(slog.LevelError)
	default:
		lvl.Set(slog.LevelInfo)
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}
