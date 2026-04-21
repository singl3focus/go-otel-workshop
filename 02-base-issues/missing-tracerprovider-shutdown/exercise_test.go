package main

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestRun_SpansExported(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()

	if err := Run(context.Background(), exp); err != nil {
		t.Fatalf("Run вернул ошибку: %v", err)
	}

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("ожидался 1 экспортированный span, получено %d", len(spans))
	}
	if spans[0].Name != "demo.operation" {
		t.Errorf("имя span: got %q, want %q", spans[0].Name, "demo.operation")
	}
}
