//go:build !solution

package main

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// EN: CreateOrder starts service.create_order and calls saveOrder.
//
// TODO: preserve the incoming trace and parent span; pass exercise_test.go.
//
// RU: CreateOrder открывает service.create_order и вызывает saveOrder.
//
// TODO: сохраните входящий trace и родительский спан, чтобы прошёл exercise_test.go.
//nolint:staticcheck // SA4009: intentional exercise / намеренное упражнение
func CreateOrder(ctx context.Context, tracer trace.Tracer) error {
	ctx, span := tracer.Start(context.Background(), "service.create_order")
	defer span.End()

	span.SetAttributes(attribute.String("order.id", "123"))
	return saveOrder(ctx, tracer)
}
