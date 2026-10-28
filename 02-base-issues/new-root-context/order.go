package main

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
)

func saveOrder(ctx context.Context, tracer trace.Tracer) error {
	_, span := tracer.Start(ctx, "repo.save_order")
	defer span.End()

	time.Sleep(10 * time.Millisecond)
	return nil
}
