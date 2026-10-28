package main

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestDoWork_FinalStateRecorded(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	if err := DoWork(context.Background(), tp.Tracer("test")); err != nil {
		t.Fatalf("DoWork returned an error / DoWork вернул ошибку: %v", err)
	}

	ended := sr.Ended()
	if len(ended) != 1 {
		t.Fatalf("expected 1 ended span / ожидался 1 завершённый спан, got %d", len(ended))
	}
	got := ended[0]

	var result string
	for _, a := range got.Attributes() {
		if string(a.Key) == "result" {
			result = a.Value.AsString()
		}
	}
	if result != "success" {
		t.Errorf("result attribute / атрибут result: got %q, want %q", result, "success")
	}

	if got.Status().Code != codes.Ok {
		t.Errorf("status / статус: got %v, want %v", got.Status().Code, codes.Ok)
	}
}
