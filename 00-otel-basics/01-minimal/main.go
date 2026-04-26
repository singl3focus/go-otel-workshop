package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	stdouttrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Минимальный путь от пустого main до записанного span:
//
//  1. exporter           — куда отправлять (stdout, OTLP, ...).
//  2. TracerProvider     — фабрика трейсеров; держит pipeline экспортеров.
//  3. otel.SetTracerProvider — глобальный default, который читают otel.Tracer(...).
//  4. tracer.Start(...)  — открывает span, кладет его в новый ctx.
//  5. span.End()         — фиксирует endTime и пушит span в pipeline.
//  6. tp.Shutdown(...)   — обязательно: иначе батчер не вытолкнет буфер
//     (здесь WithSyncer, поэтому даже без Shutdown спан
//     ушел бы — но привычку лучше выработать сразу).
func main() {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	defer func() { _ = tp.Shutdown(context.Background()) }()

	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("00-otel-basics/01-minimal")
	_, span := tracer.Start(context.Background(), "hello")
	span.End()
}
