package main

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Run симулирует жизненный цикл приложения: создает TracerProvider с batch-процессором,
// выполняет работу и возвращает управление.
//
// TODO: все открытые spans должны быть экспортированы до возврата из Run —
// добейся того, чтобы exercise_test.go стал зеленым.
func Run(ctx context.Context, exporter sdktrace.SpanExporter) error {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)

	tracer := tp.Tracer("exercise")
	_, span := tracer.Start(ctx, "demo.operation")
	span.SetAttributes(attribute.String("demo", "exercise"))
	span.End()

	return nil
}
