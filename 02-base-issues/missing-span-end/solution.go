//go:build solution

package main

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func DoWorkSolution(ctx context.Context, tracer trace.Tracer) error {
	_, span := tracer.Start(ctx, "demo.operation")
	defer span.End()

	span.SetAttributes(attribute.String("demo", "exercise"))
	return nil
}
