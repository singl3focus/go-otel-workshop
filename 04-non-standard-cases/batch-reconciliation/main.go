package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	nsc "github.com/singl3focus/go-otel-workshop/04-non-standard-cases"
)

const (
	tickInterval = 2 * time.Second
	tickTimeout  = 1500 * time.Millisecond
	pageSize     = 3
)

func main() {
	if err := run(); err != nil {
		log.Printf("reconciliation loop failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	tp := nsc.NewTracerProvider("example-batch-reconciliation")
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(shutdownCtx)
	}()
	otel.SetTracerProvider(tp)

	tracer := otel.Tracer("04-non-standard-cases/batch-reconciliation")

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	runOnce(rootCtx, tracer)
	for {
		select {
		case <-rootCtx.Done():
			return nil
		case <-ticker.C:
			runOnce(rootCtx, tracer)
		}
	}
}

// runOnce - один тик реконсиляции.
//
// Главное: КАЖДЫЙ тик - НОВЫЙ root span. Если обернуть всю for-петлю
// одним span'ом, трасса растёт бесконечно, экспортер не может её закрыть,
// в UI это выглядит как зависшая операция.
//
// WithNewRoot делает семантику явной: даже если в parent ctx вдруг окажется
// активный span, мы его не наследуем.
func runOnce(parent context.Context, tracer trace.Tracer) {
	ctx, cancel := context.WithTimeout(parent, tickTimeout)
	defer cancel()

	ctx, root := tracer.Start(ctx, "reconcile.tick",
		trace.WithNewRoot(),
		trace.WithAttributes(
			attribute.String("job.name", "reconciliation"),
			attribute.Int("job.page_size", pageSize),
		),
	)
	defer root.End()

	processed, failed, err := reconcile(ctx, tracer)
	root.SetAttributes(
		attribute.Int("reconcile.processed", processed),
		attribute.Int("reconcile.failed", failed),
	)
	if err != nil {
		root.RecordError(err)
		root.SetStatus(codes.Error, "tick failed")
	}
	log.Printf("tick done: processed=%d failed=%d err=%v", processed, failed, err)
}

func reconcile(ctx context.Context, tracer trace.Tracer) (processed, failed int, err error) {
	for page := 0; ; page++ {
		items, last, ferr := fetchPage(ctx, tracer, page)
		if ferr != nil {
			return processed, failed, ferr
		}
		for _, id := range items {
			if rerr := reconcileOne(ctx, tracer, id); rerr != nil {
				failed++
				continue
			}
			processed++
		}
		if last {
			return processed, failed, nil
		}
	}
}

func fetchPage(ctx context.Context, tracer trace.Tracer, page int) ([]string, bool, error) {
	_, span := tracer.Start(ctx, "reconcile.fetch_page",
		trace.WithAttributes(attribute.Int("page", page)),
	)
	defer span.End()

	time.Sleep(30 * time.Millisecond)
	ids := []string{"a", "b", "c"}
	span.SetAttributes(attribute.Int("page.size", len(ids)))
	return ids, page >= 1, nil
}

func reconcileOne(ctx context.Context, tracer trace.Tracer, id string) error {
	_, span := tracer.Start(ctx, "reconcile.one",
		trace.WithAttributes(attribute.String("entity.id", id)),
	)
	defer span.End()
	time.Sleep(10 * time.Millisecond)
	return nil
}
