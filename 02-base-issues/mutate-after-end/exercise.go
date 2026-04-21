package main

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// DoWork выполняет операцию: фиксирует финальный атрибут result=success
// и статус Ok на span, затем завершает его.
//
// TODO: в итоговом span должны быть атрибут result=success и статус Ok —
// добейся того, чтобы exercise_test.go стал зеленым.
func DoWork(ctx context.Context, tracer trace.Tracer) error {
	_, span := tracer.Start(ctx, "demo.operation")

	span.End()

	span.SetAttributes(attribute.String("result", "success"))
	span.SetStatus(codes.Ok, "done")

	return nil
}
