package server

import (
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"

	"github.com/singl3focus/go-otel-workshop/01-local-observability-stack/app/internal/server/handlers"
	"github.com/singl3focus/go-otel-workshop/01-local-observability-stack/app/internal/server/middleware"
	"github.com/singl3focus/go-otel-workshop/01-local-observability-stack/app/internal/service"
)

func NewRouter(appName string, logger *slog.Logger) http.Handler {
	svc := service.New(appName)
	h := handlers.New(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /work", h.Work)
	mux.HandleFunc("GET /error", h.Error)

	handler := middleware.RequestLogging(logger)(mux)
	tracer := otel.Tracer("local-observability-stack/http")
	return middleware.RequestTracing(tracer)(handler)
}
