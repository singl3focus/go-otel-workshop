package main

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestCreateOrder_InheritsParentTrace(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	tracer := tp.Tracer("test")

	ctx, parent := tracer.Start(context.Background(), "test.parent")
	if err := CreateOrder(ctx, tracer); err != nil {
		t.Fatalf("CreateOrder returned an error / CreateOrder вернул ошибку: %v", err)
	}
	parent.End()

	ended := sr.Ended()
	if len(ended) != 3 {
		t.Fatalf("expected 3 ended spans / ожидались 3 завершённых спана (test.parent, service.create_order, repo.save_order), got %d", len(ended))
	}

	parentTraceID := parent.SpanContext().TraceID()
	parentSpanID := parent.SpanContext().SpanID()

	var createOrderSpan sdktrace.ReadOnlySpan
	for _, s := range ended {
		if s.SpanContext().TraceID() != parentTraceID {
			t.Errorf("span %q: TraceID %s differs from parent / отличается от родительского %s",
				s.Name(), s.SpanContext().TraceID(), parentTraceID)
		}
		if s.Name() == "service.create_order" {
			createOrderSpan = s
		}
	}

	if createOrderSpan == nil {
		t.Fatal("service.create_order span not found / спан service.create_order не найден")
	}
	if createOrderSpan.Parent().SpanID() != parentSpanID {
		t.Errorf("service.create_order.parent_span_id = %s, expected / ожидался %s",
			createOrderSpan.Parent().SpanID(), parentSpanID)
	}
}
