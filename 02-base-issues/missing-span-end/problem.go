package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	baseissues "github.com/singl3focus/go-otel-workshop/02-base-issues"
)

// =========================
// good.span завершится и уйдет в export path.
// broken.span — нет, потому что завершение span связано с End(),
// а Shutdown завершает processors, но не "додумывает" за тебя незавершенные spans.
// TracerProvider.Shutdown(ctx) закрывает провайдер и processors;
// после него методы становятся no-op.
// =========================

func main() {
	ctx := context.Background()

	tp := baseissues.NewTracerProvider("example-missing-span-end")
	defer func() {
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		_ = tp.Shutdown(ctx)
	}()

	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("02-base-issues/missing-span-end")

	{
		_, span := tracer.Start(ctx, "good.span")
		span.SetAttributes(attribute.String("demo", "good"))
		span.End()
	}

	{
		_, span := tracer.Start(ctx, "broken.span")
		span.SetAttributes(attribute.String("demo", "missing-end"))

		// ПЛОХО:
		// span.End() забыли.
		// Такой span не пройдет нормальный путь завершения через processor.OnEnd(...).
		_ = span
	}

	log.Println("done, wait a bit...")
	time.Sleep(500 * time.Millisecond)
}
