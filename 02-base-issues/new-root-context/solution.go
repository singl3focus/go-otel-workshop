//go:build solution

package main

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func CreateOrderSolution(ctx context.Context, tracer trace.Tracer) error {
	ctx, span := tracer.Start(ctx, "service.create_order")
	defer span.End()

	span.SetAttributes(attribute.String("order.id", "123"))
	return saveOrder(ctx, tracer)
}
