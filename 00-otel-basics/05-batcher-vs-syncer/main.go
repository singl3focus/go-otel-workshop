package main

import (
	"context"

	"go.opentelemetry.io/otel"
	stdouttrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// EN: WithSyncer exports during span.End, useful for small demos and tests.
// WithBatcher queues spans and exports periodically or when a batch fills.
//
// A short process can exit before its queue is exported without Shutdown.
// Run as-is: buffered.span may not appear. Uncomment the deferred Shutdown
// to flush it reliably. A fast background export can still win the race.
//
// Continue with 02-base-issues/missing-tracerprovider-shutdown/.
//
// RU: WithSyncer экспортирует спан при End: удобно для небольших примеров и тестов.
// WithBatcher ставит спаны в очередь и экспортирует по таймеру или размеру батча.
//
// Без Shutdown короткий процесс может выйти раньше экспорта.
// Запустите как есть: buffered.span может отсутствовать. Раскомментируйте
// отложенный Shutdown, чтобы надёжно выгрузить очередь. Быстрый фоновый
// экспорт иногда успевает сработать и без него.
//
// Следующее упражнение: 02-base-issues/missing-tracerprovider-shutdown/.
func main() {
	exp, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())

	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp))
	// defer func() { _ = tp.Shutdown(context.Background()) }()
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("00-otel-basics/05-batcher-vs-syncer")
	_, span := tracer.Start(context.Background(), "buffered.span")
	span.End()
}
