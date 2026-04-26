package main

import (
	"context"

	"go.opentelemetry.io/otel"
	stdouttrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// WithSyncer  — экспорт прямо в span.End(). Подходит для тестов и CLI.
// WithBatcher — спан кладется в очередь, экспорт по таймеру или при заполнении.
//
//	Без tp.Shutdown короткая программа выйдет до того, как буфер
//	выгрузится — спан потеряется.
//
// Запусти как есть — увидишь, что в выводе НЕТ спана "buffered.span":
// батчер не успел вытолкнуть очередь до выхода main. Раскомментируй строку
// с defer Shutdown — спан появится.
//
// Это и есть точка входа в антипаттерн "missing TracerProvider Shutdown" —
// см. 02-base-issues/missing-tracerprovider-shutdown/.
func main() {
	exp, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())

	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp))
	// defer func() { _ = tp.Shutdown(context.Background()) }()
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("00-otel-basics/05-batcher-vs-syncer")
	_, span := tracer.Start(context.Background(), "buffered.span")
	span.End()
}
