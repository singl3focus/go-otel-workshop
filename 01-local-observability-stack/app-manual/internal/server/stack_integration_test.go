//go:build integration

package server

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/singl3focus/go-otel-workshop/01-local-observability-stack/internal/stackcheck"
	"github.com/singl3focus/go-otel-workshop/01-local-observability-stack/observability"
)

// EN: Prove HTTP -> SDK -> Collector -> Jaeger, including 503 status and parents.
// RU: Проверяем HTTP -> SDK -> Collector -> Jaeger, включая 503 и связи родителей.
func TestLocalStack(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	endpoint := os.Getenv("WORKSHOP_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://127.0.0.1:4317"
	}
	queryURL := os.Getenv("WORKSHOP_JAEGER_URL")
	if queryURL == "" {
		queryURL = "http://127.0.0.1:16686"
	}
	service := "workshop-manual-integration"
	shutdown, err := observability.Setup(ctx, service, endpoint)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		stop, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(stop)
	})
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	server := httptest.NewServer(NewRouter(service, logger))
	defer server.Close()
	expected, err := stackcheck.Probe(ctx, server.URL, service)
	if err != nil {
		t.Fatal(err)
	}
	flush, finish := context.WithTimeout(ctx, 10*time.Second)
	err = shutdown(flush)
	finish()
	if err != nil {
		t.Fatalf("OTLP export / экспорт OTLP: %v", err)
	}
	if err := stackcheck.Wait(ctx, queryURL, expected); err != nil {
		t.Fatal(err)
	}
	for _, trace := range expected {
		t.Logf("verified / проверено: %s %s/trace/%s", trace.ServerName, queryURL, trace.TraceID)
	}
}
