package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	nsc "github.com/singl3focus/go-otel-workshop/04-non-standard-cases"
)

const (
	totalKeys  = 200
	numWorkers = 4
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Printf("warmup failed: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) (err error) {
	tp := nsc.NewTracerProvider("example-cache-warmup")
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(shutdownCtx)
	}()
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("04-non-standard-cases/cache-warmup")

	ctx, root := tracer.Start(ctx, "cache.warmup",
		trace.WithAttributes(
			attribute.Int("keys.total", totalKeys),
			attribute.Int("workers", numWorkers),
		),
	)
	defer func() {
		if err != nil {
			root.RecordError(err)
			root.SetStatus(codes.Error, "warmup failed")
		}
		root.End()
	}()

	tasks := make(chan string, numWorkers)
	var wg sync.WaitGroup
	var warmed, miss int64

	// Воркеры живут внутри run() - ctx валиден до wg.Wait(),
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			w, m := worker(ctx, tracer, id, tasks)
			atomic.AddInt64(&warmed, int64(w))
			atomic.AddInt64(&miss, int64(m))
		}(i)
	}

	for i := 0; i < totalKeys; i++ {
		tasks <- fmt.Sprintf("user:%d", i)
	}
	close(tasks)
	wg.Wait()

	root.SetAttributes(
		attribute.Int64("keys.warmed", warmed),
		attribute.Int64("keys.miss", miss),
	)
	log.Printf("warmup done: warmed=%d miss=%d", warmed, miss)
	return nil
}

// worker - один span на всю работу воркера, не на каждый ключ.
//
// Почему: 1M ключей × span = взрыв трассировки. Отдельные мисс-события идут
// через span.AddEvent, агрегаты - через атрибуты воркера. В root поднимаются
// суммарные warmed/miss - этого достаточно для SLI прогрева.
func worker(ctx context.Context, tracer trace.Tracer, id int, tasks <-chan string) (warmed, miss int) {
	_, span := tracer.Start(ctx, "cache.worker",
		trace.WithAttributes(attribute.Int("worker.id", id)),
	)
	defer span.End()

	for key := range tasks {
		if !warmOne(key) {
			miss++
			span.AddEvent("warm.miss", trace.WithAttributes(attribute.String("key", key)))
			continue
		}
		warmed++
	}

	span.SetAttributes(
		attribute.Int("worker.warmed", warmed),
		attribute.Int("worker.miss", miss),
	)
	return warmed, miss
}

func warmOne(key string) bool {
	time.Sleep(2 * time.Millisecond)
	// имитация: ~10% ключей отсутствуют в источнике
	return len(key)%10 != 0
}
