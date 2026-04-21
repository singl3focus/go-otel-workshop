package main

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DoWork открывает span "demo.operation" с атрибутом demo=exercise.
//
// TODO: span должен быть корректно принят и успешно отработан —
// добейся того, чтобы exercise_test.go прошел зеленым.
func DoWork(ctx context.Context, tracer trace.Tracer) error {
	_, span := tracer.Start(ctx, "demo.operation")
	span.SetAttributes(attribute.String("demo", "exercise"))
	_ = span

	return nil
}
