package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	nsc "github.com/singl3focus/go-otel-workshop/04-non-standard-cases"
)

const (
	totalDocs = 500
	batchSize = 100
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Printf("reindex failed: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) (err error) {
	tp := nsc.NewTracerProvider("example-reindex")
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(shutdownCtx)
	}()
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("04-non-standard-cases/reindex")

	ctx, root := tracer.Start(ctx, "reindex.run",
		trace.WithAttributes(
			attribute.String("index.name", "orders_v3"),
			attribute.Int("source.total", totalDocs),
			attribute.Int("batch.size", batchSize),
		),
	)
	defer func() {
		if err != nil {
			root.RecordError(err)
			root.SetStatus(codes.Error, "reindex failed")
		}
		root.End()
	}()

	var indexed, failed int
	for offset := 0; offset < totalDocs; offset += batchSize {
		n, f, berr := indexBatch(ctx, tracer, offset, batchSize)
		indexed += n
		failed += f
		if berr != nil {
			root.SetAttributes(
				attribute.Int("reindex.indexed", indexed),
				attribute.Int("reindex.failed", failed),
			)
			return berr
		}
	}

	root.SetAttributes(
		attribute.Int("reindex.indexed", indexed),
		attribute.Int("reindex.failed", failed),
	)
	log.Printf("reindex done: indexed=%d failed=%d", indexed, failed)
	return nil
}

// indexBatch - span на БАТЧ, не на документ.
//
// Почему: при 10M документов span-per-doc ложит экспортер и стоимость хранения.
// Уровень батча - естественная граница observability: latency batch-запроса
// в поисковый движок, retries, размер. Отдельные документы - через span events.
func indexBatch(ctx context.Context, tracer trace.Tracer, offset, size int) (indexed, failed int, err error) {
	_, span := tracer.Start(ctx, "reindex.batch",
		trace.WithAttributes(
			attribute.Int("batch.offset", offset),
			attribute.Int("batch.size", size),
		),
	)
	defer span.End()

	// имитация bulk-запроса в поисковый движок
	time.Sleep(50 * time.Millisecond)

	for i := 0; i < size; i++ {
		if (offset+i)%137 == 0 {
			span.AddEvent("doc.failed", trace.WithAttributes(
				attribute.String("doc.id", fmt.Sprintf("doc-%d", offset+i)),
				attribute.String("reason", "validation"),
			))
			failed++
			continue
		}
		indexed++
	}

	span.SetAttributes(
		attribute.Int("batch.indexed", indexed),
		attribute.Int("batch.failed", failed),
	)
	return indexed, failed, nil
}
