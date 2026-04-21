package main

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestDoWork_SpanEnded(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	if err := DoWork(context.Background(), tp.Tracer("test")); err != nil {
		t.Fatalf("DoWork вернул ошибку: %v", err)
	}

	ended := sr.Ended()
	if len(ended) != 1 {
		t.Fatalf("ожидался 1 завершенный span, получено %d", len(ended))
	}

	got := ended[0]
	if got.Name() != "demo.operation" {
		t.Errorf("имя span: got %q, want %q", got.Name(), "demo.operation")
	}
	if got.EndTime().IsZero() {
		t.Error("EndTime пустой — span не был завершен")
	}

	var demo string
	for _, a := range got.Attributes() {
		if string(a.Key) == "demo" {
			demo = a.Value.AsString()
		}
	}
	if demo != "exercise" {
		t.Errorf("attribute demo: got %q, want %q", demo, "exercise")
	}
}
