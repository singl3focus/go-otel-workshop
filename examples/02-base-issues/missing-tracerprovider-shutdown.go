package baseissues

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// =========================
// С WithBatcher(...) экспорт идет через batch processor.
// Если приложение завершится сразу и не вызовет tp.Shutdown(ctx),
// буферизованные spans могут не успеть уйти.
// Shutdown(ctx) штатно завершает processors и освобождает ресурсы.
// =========================

func main4() {
	tp := newTracerProvider("example-missing-shutdown")

	// ПЛОХО:
	// Мы намеренно НЕ вызываем tp.Shutdown(...).
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("examples/05-missing-tracerprovider-shutdown")

	ctx := context.Background()
	_, span := tracer.Start(ctx, "process.exit.without.shutdown")
	span.SetAttributes(
		attribute.String("demo", "missing-shutdown"),
		attribute.Int("items", 42),
	)
	span.End()

	log.Println("program exits immediately without tracer provider shutdown")
	// Нет time.Sleep, нет Shutdown: buffered spans могут не успеть выгрузиться.
}