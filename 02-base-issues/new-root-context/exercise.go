package main

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// CreateOrder открывает span "service.create_order" и вызывает saveOrder.
//
// TODO: все spans в цепочке должны принадлежать одному трейсу —
// добейся того, чтобы exercise_test.go стал зеленым.
//
//nolint:staticcheck // SA4009: intentional anti-pattern — ctx намеренно заменен на context.Background(), задание — найти и починить.
func CreateOrder(ctx context.Context, tracer trace.Tracer) error {
	ctx, span := tracer.Start(context.Background(), "service.create_order")
	defer span.End()

	span.SetAttributes(attribute.String("order.id", "123"))
	return saveOrder(ctx, tracer)
}

func saveOrder(ctx context.Context, tracer trace.Tracer) error {
	_, span := tracer.Start(ctx, "repo.save_order")
	defer span.End()

	time.Sleep(10 * time.Millisecond)
	return nil
}
