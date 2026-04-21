//go:build solution

package main

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func RunSolution(ctx context.Context, exporter sdktrace.SpanExporter) error {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)
	defer func() { _ = tp.Shutdown(ctx) }()

	tracer := tp.Tracer("exercise")
	_, span := tracer.Start(ctx, "demo.operation")
	span.SetAttributes(attribute.String("demo", "exercise"))
	span.End()

	return nil
}
