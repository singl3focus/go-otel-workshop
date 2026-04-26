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

// Что можно навесить на span между Start и End:
//
//	SetAttributes — структурированные key/value (фильтр в бэкенде).
//	AddEvent      — момент времени внутри span'а (лог-как-событие).
//	SetStatus     — финальный статус: codes.Ok / codes.Error.
//	RecordError   — кладет error как event со стектрейсом и atom-атрибутами.
//
// RecordError не выставляет статус Error сам — нужно дополнить SetStatus.
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
