package main

import (
	"context"
	"errors"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	stdouttrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// EN: Enrich a recording span between Start and End:
//
// SetAttributes adds structured key/value data.
// AddEvent records a timestamped event inside the span.
// SetStatus marks the operation as Ok or Error (Unset is the default).
// RecordError adds an exception event; stack capture is opt-in.
//
// RecordError does not set Error status. Call SetStatus explicitly.
//
// RU: Между Start и End записывающий span можно дополнить:
//
// SetAttributes добавляет структурированные пары ключ/значение.
// AddEvent записывает событие с отметкой времени.
// SetStatus задаёт Ok или Error; исходный статус — Unset.
// RecordError добавляет событие исключения; стек включается отдельно.
//
// RecordError сам не задаёт статус Error: нужен явный SetStatus.
func main() {
	exp, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	defer func() { _ = tp.Shutdown(context.Background()) }()
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("00-otel-basics/02-attributes-events-status")

	_, okSpan := tracer.Start(context.Background(), "ok.span")
	okSpan.SetAttributes(
		attribute.String("user.id", "42"),
		attribute.Int("items.count", 3),
	)
	okSpan.AddEvent("cache.hit")
	okSpan.SetStatus(codes.Ok, "")
	okSpan.End()

	_, badSpan := tracer.Start(context.Background(), "bad.span")
	err := errors.New("db connection refused")
	badSpan.RecordError(err)
	badSpan.SetStatus(codes.Error, err.Error())
	badSpan.End()

	log.Println("done")
}
