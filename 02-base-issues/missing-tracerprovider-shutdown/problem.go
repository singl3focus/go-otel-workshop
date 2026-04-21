package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	baseissues "github.com/singl3focus/go-otel-workshop/02-base-issues"
)

// =========================
// С WithBatcher(...) экспорт идет через batch processor.
// Если приложение завершится сразу и не вызовет tp.Shutdown(ctx),
// буферизированные spans могут не успеть уйти.
// Shutdown(ctx) штатно завершает processors и освобождает ресурсы.
// =========================

func main() {
	tp := baseissues.NewTracerProvider("example-missing-shutdown")

	// ПЛОХО:
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

	log.Println("program exits immediately without tracer provider shutdown after creating 100 spans")
	// Нет time.Sleep, нет Shutdown: buffered spans могут не успеть выгрузиться.
}
