package main

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	baseissues "github.com/singl3focus/go-otel-workshop/02-base-issues"
)

// EN: good.span reaches the export pipeline because it is ended.
// broken.span does not: Shutdown stops processors and flushes ended spans,
// but does not end spans on the application's behalf.
// After provider shutdown, newly started spans are non-recording.
//
// RU: good.span попадёт в экспорт, потому что вызван End.
// broken.span не попадёт: Shutdown выгружает завершённые спаны и останавливает
// процессоры, но не завершает спаны за приложение.
// После shutdown новые спаны провайдера не записываются.

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

		// EN: BAD: span.End is missing, so processor.OnEnd is never called.
		//
		// RU: ПЛОХО:
		// span.End() забыли.
		// Такой span не пройдет нормальный путь завершения через processor.OnEnd(...).
		_ = span
	}

	log.Println("done, wait a bit... / готово, дождитесь экспорта...")
	time.Sleep(500 * time.Millisecond)
}
