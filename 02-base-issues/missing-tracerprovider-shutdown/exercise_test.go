package main

import (
	"context"
	"sync/atomic"
	"testing"

	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestRun_SpansExported(t *testing.T) {
	exp := &lifecycleExporter{InMemoryExporter: tracetest.NewInMemoryExporter()}

	if err := Run(context.Background(), exp); err != nil {
		t.Fatalf("Run returned an error / Run вернул ошибку: %v", err)
	}

	if !exp.stopped.Load() {
		t.Fatal("exporter was not shut down / экспортёр не был остановлен")
	}
	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 exported span / ожидался 1 экспортированный спан, got %d", len(spans))
	}
	if spans[0].Name != "demo.operation" {
		t.Errorf("span name / имя спана: got %q, want %q", spans[0].Name, "demo.operation")
	}
}

// EN: Keep exported spans observable after Shutdown; the SDK test exporter resets them.
// RU: Сохраняем спаны после Shutdown: стандартный тестовый экспортёр SDK очищает их.
type lifecycleExporter struct {
	*tracetest.InMemoryExporter
	stopped atomic.Bool
}

func (e *lifecycleExporter) Shutdown(context.Context) error {
	e.stopped.Store(true)
	return nil
}
