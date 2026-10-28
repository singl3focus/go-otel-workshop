package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	baseissues "github.com/singl3focus/go-otel-workshop/02-base-issues"
)

// EN: WithBatcher exports through a background batch processor.
// If the application exits immediately without tp.Shutdown(ctx),
// buffered spans may never be exported. Shutdown flushes the processor
// and releases its resources.
//
// RU: =========================
// С WithBatcher(...) экспорт идет через batch processor.
// Если приложение завершится сразу и не вызовет tp.Shutdown(ctx),
// буферизированные spans могут не успеть уйти.
// Shutdown(ctx) штатно завершает processors и освобождает ресурсы.
// =========================

func main() {
	tp := baseissues.NewTracerProvider("example-missing-shutdown")

	// EN: BAD: deliberately omit tp.Shutdown.
	//
	// RU: ПЛОХО:
	// Мы намеренно НЕ вызываем tp.Shutdown(...).
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("02-base-issues/missing-tracerprovider-shutdown")

	ctx := context.Background()
	for i := range 100 {
		_, span := tracer.Start(ctx, "process.exit.without.shutdown")
		span.SetAttributes(
			attribute.String("demo", "missing-shutdown"),
			attribute.Int("items", 42),
			attribute.Int("span.index", i),
		)
		span.End()
	}

	log.Println("program exits after 100 spans without provider shutdown / выход после 100 спанов без shutdown провайдера")
	// EN: No sleep and no Shutdown: buffered spans can be lost on process exit.
	//
	// RU: Нет time.Sleep, нет Shutdown: buffered spans могут не успеть выгрузиться.
}
