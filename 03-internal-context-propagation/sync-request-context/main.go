package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	icp "github.com/singl3focus/go-otel-workshop/03-internal-context-propagation"
)

func main() {
	tp := icp.NewTracerProvider("example-sync-request-context")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(ctx)
	}()

	otel.SetTracerProvider(tp)

	app := newApp()

	mux := http.NewServeMux()
	mux.Handle("/sync", otelhttp.NewHandler(http.HandlerFunc(app.handleSync), "http.sync"))

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

type app struct {
	tracer  trace.Tracer
	service *service
}

func newApp() *app {
	return &app{
		tracer:  otel.Tracer("03-internal-context-propagation/sync-request-context"),
		service: &service{tracer: otel.Tracer("03-internal-context-propagation/sync-request-context/service")},
	}
}

func (a *app) handleSync(w http.ResponseWriter, r *http.Request) {
	ctx, span := a.tracer.Start(r.Context(), "handler.sync")
	defer span.End()

	if err := a.service.process(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("sync ok\n"))
}

type service struct {
	tracer trace.Tracer
}

func (s *service) process(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "service.process")
	defer span.End()

	span.SetAttributes(attribute.String("flow.type", "sync"))

	return s.save(ctx)
}

func (s *service) save(ctx context.Context) error {
	_, span := s.tracer.Start(ctx, "repo.save")
	defer span.End()

	time.Sleep(120 * time.Millisecond)
	return nil
}
