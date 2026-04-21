package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	nsc "github.com/singl3focus/go-otel-workshop/04-non-standard-cases"
)

type migration struct {
	name string
	sql  string
}

var schema = []migration{
	{name: "001_create_users.sql", sql: "CREATE TABLE users (...)"},
	{name: "002_add_email_index.sql", sql: "CREATE INDEX users_email_idx ON users (email)"},
	{name: "003_backfill_status.sql", sql: "UPDATE users SET status = 'active' WHERE status IS NULL"},
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Printf("migrations failed: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) (err error) {
	tp := nsc.NewTracerProvider("example-migrations")
	// CLI короткоживущий - без Shutdown батчер не успевает вытолкнуть последние спаны.
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(shutdownCtx)
	}()

	otel.SetTracerProvider(tp)
	tracer := otel.Tracer("04-non-standard-cases/migrations")

	// Нет входящего ctx - root span создаём сами.
	ctx, root := tracer.Start(ctx, "migrations.run",
		trace.WithAttributes(
			attribute.String("db.system", "postgres"),
			attribute.Int("migrations.total", len(schema)),
		),
	)
	defer func() {
		if err != nil {
			root.RecordError(err)
			root.SetStatus(codes.Error, "migration run failed")
		}
		root.End()
	}()

	var applied int
	for _, m := range schema {
		if err = apply(ctx, tracer, m); err != nil {
			// Счётчик applied оседает на root даже при падении - по одному спану
			// видно, на какой миграции порвалось.
			root.SetAttributes(attribute.Int("migrations.applied", applied))
			return err
		}
		applied++
	}

	root.SetAttributes(attribute.Int("migrations.applied", applied))
	log.Printf("migrations done: %d applied", applied)
	return nil
}

func apply(ctx context.Context, tracer trace.Tracer, m migration) (err error) {
	_, span := tracer.Start(ctx, "migration.apply",
		trace.WithAttributes(attribute.String("migration.name", m.name)),
	)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	// имитация исполнения SQL драйвером БД
	time.Sleep(80 * time.Millisecond)
	return nil
}
