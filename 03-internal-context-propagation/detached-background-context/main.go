package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	icp "github.com/singl3focus/go-otel-workshop/03-internal-context-propagation"
)

func main() {
	tp := icp.NewTracerProvider("example-detached-background-context")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(ctx)
	}()

	otel.SetTracerProvider(tp)

	app := newApp()

	mux := http.NewServeMux()
	mux.Handle("/bad", otelhttp.NewHandler(http.HandlerFunc(app.handleBad), "http.bad"))
	mux.Handle("/good", otelhttp.NewHandler(http.HandlerFunc(app.handleGood), "http.good"))

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

type app struct {
	tracer trace.Tracer
}

func newApp() *app {
	return &app{
		tracer: otel.Tracer("03-internal-context-propagation/detached-background-context"),
	}
}

// Плохой вариант: используем request ctx в фоне после ответа клиенту.
func (a *app) handleBad(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	go func() {
		time.Sleep(250 * time.Millisecond)

		ctx, span := a.tracer.Start(ctx, "background.bad")
		defer span.End()

		if err := simulateDownstreamCall(ctx); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "request context already canceled")
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("bad background task scheduled\n"))
}

// Хороший вариант: сохраняем values/trace linkage, но отвязываемся от cancel родителя.
func (a *app) handleGood(w http.ResponseWriter, r *http.Request) {
	bg := context.WithoutCancel(r.Context())
	bg, cancel := context.WithTimeout(bg, 2*time.Second)
	defer cancel()

	go func() {
		time.Sleep(250 * time.Millisecond)

		ctx, span := a.tracer.Start(bg, "background.good")
		defer span.End()

		span.SetAttributes(attribute.String("context.mode", "detached"))

		if err := simulateDownstreamCall(ctx); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "unexpected error")
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("good background task scheduled\n"))
}

func simulateDownstreamCall(ctx context.Context) error {
	select {
	case <-time.After(100 * time.Millisecond):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
