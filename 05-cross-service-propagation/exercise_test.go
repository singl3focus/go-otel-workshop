package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestReserveInventory_PropagatesTraceContext(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	t.Cleanup(func() {
		otel.SetTracerProvider(noop.NewTracerProvider())
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())
	})

	downstreamTraceID := make(chan trace.TraceID, 1)
	downstream := httptest.NewServer(otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc := trace.SpanContextFromContext(r.Context())
		downstreamTraceID <- sc.TraceID()
		w.WriteHeader(http.StatusNoContent)
	}), "POST "+inventoryReservePath))
	t.Cleanup(downstream.Close)

	tracer := tp.Tracer("test")
	ctx, parent := tracer.Start(context.Background(), "test.parent")
	err := ReserveInventory(ctx, downstream.Client(), downstream.URL)
	parent.End()
	if err != nil {
		t.Fatalf("ReserveInventory вернул ошибку: %v", err)
	}

	select {
	case got := <-downstreamTraceID:
		if got != parent.SpanContext().TraceID() {
			t.Fatalf("downstream TraceID = %s, ожидался %s — trace context не перешел через HTTP",
				got, parent.SpanContext().TraceID())
		}
	default:
		t.Fatal("downstream handler не был вызван")
	}

	ended := sr.Ended()
	var clientSpan sdktrace.ReadOnlySpan
	var serverSpan sdktrace.ReadOnlySpan
	for _, span := range ended {
		switch span.SpanKind() {
		case trace.SpanKindClient:
			clientSpan = span
		case trace.SpanKindServer:
			serverSpan = span
		}
	}

	if clientSpan == nil {
		t.Fatal("ожидался outbound client span — оберни Transport через otelhttp.NewTransport")
	}
	if serverSpan == nil {
		t.Fatal("ожидался downstream server span")
	}
	if serverSpan.SpanContext().TraceID() != parent.SpanContext().TraceID() {
		t.Errorf("downstream server span TraceID = %s, ожидался %s",
			serverSpan.SpanContext().TraceID(), parent.SpanContext().TraceID())
	}
	if serverSpan.Parent().SpanID() != clientSpan.SpanContext().SpanID() {
		t.Errorf("downstream server parent_span_id = %s, ожидался client span %s",
			serverSpan.Parent().SpanID(), clientSpan.SpanContext().SpanID())
	}
}
