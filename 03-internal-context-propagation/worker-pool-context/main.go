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
	tp := icp.NewTracerProvider("example-worker-pool-context")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(ctx)
	}()

	otel.SetTracerProvider(tp)

	app := newApp()

	mux := http.NewServeMux()
	mux.Handle("/enqueue", otelhttp.NewHandler(http.HandlerFunc(app.handleEnqueue), "http.enqueue"))

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

type job struct {
	parent  trace.SpanContext
	orderID string
}

type app struct {
	tracer trace.Tracer
	queue  chan job
}

func newApp() *app {
	a := &app{
		tracer: otel.Tracer("03-internal-context-propagation/worker-pool-context"),
		queue:  make(chan job, 16),
	}

	for i := 0; i < 2; i++ {
		go a.worker(i)
	}

	return a
}

func (a *app) handleEnqueue(w http.ResponseWriter, r *http.Request) {
	ctx, span := a.tracer.Start(r.Context(), "handler.enqueue")
	defer span.End()

	a.queue <- job{
		parent:  trace.SpanContextFromContext(ctx),
		orderID: "order-123",
	}

	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("job accepted\n"))
}

func (a *app) worker(id int) {
	for j := range a.queue {
		a.process(id, j)
	}
}

func (a *app) process(id int, j job) {
	ctx := trace.ContextWithSpanContext(context.Background(), j.parent)
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, span := a.tracer.Start(ctx, "worker.process")
	defer span.End()

	span.SetAttributes(
		attribute.Int("worker.id", id),
		attribute.String("order.id", j.orderID),
	)

	time.Sleep(150 * time.Millisecond)
}
