package baseissues

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// =========================
// good.span завершится и уйдет в export path.
// broken.span — нет, потому что завершение span связано с End(),
// а Shutdown завершает processors, но не “додумывает” за тебя незавершенные spans. TracerProvider.Shutdown(ctx) закрывает провайдер и processors;
// после него методы становятся no-op.
// =========================

func main2() {
	tp := newTracerProvider("example-missing-span-end")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = tp.Shutdown(ctx)
	}()

	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("examples/03-missing-span-end")

	ctx := context.Background()

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
