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
		t.Fatalf("DoWork вернул ошибку: %v", err)
	}

	ended := sr.Ended()
	if len(ended) != 1 {
		t.Fatalf("ожидался 1 завершенный span, получено %d", len(ended))
	}
	got := ended[0]

	var result string
	for _, a := range got.Attributes() {
		if string(a.Key) == "result" {
			result = a.Value.AsString()
		}
	}
	if result != "success" {
		t.Errorf("attribute result: got %q, want %q", result, "success")
	}

	if got.Status().Code != codes.Ok {
		t.Errorf("status code: got %v, want %v", got.Status().Code, codes.Ok)
	}
}
