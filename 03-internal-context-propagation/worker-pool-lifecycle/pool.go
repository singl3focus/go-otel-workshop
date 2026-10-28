package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	errQueueFull  = errors.New("queue full / очередь заполнена")
	errPoolClosed = errors.New("pool closed / пул закрыт")
)

type job struct {
	id     string
	parent trace.SpanContext
}

type poolConfig struct {
	workers    int
	capacity   int
	jobTimeout time.Duration
}

type poolStats struct {
	accepted  int64
	completed int64
	failed    int64
	full      int64
}

type pool struct {
	tracer    trace.Tracer
	client    *http.Client
	endpoint  string
	timeout   time.Duration
	jobs      chan job
	done      chan struct{}
	runCtx    context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
	closed    bool
	accepted  atomic.Int64
	completed atomic.Int64
	failed    atomic.Int64
	full      atomic.Int64
}

func newPool(cfg poolConfig, tracer trace.Tracer, client *http.Client, endpoint string) (*pool, error) {
	if cfg.workers < 1 || cfg.capacity < 1 || cfg.jobTimeout <= 0 {
		return nil, fmt.Errorf("positive worker, queue and timeout limits required / нужны положительные лимиты воркеров, очереди и времени")
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &pool{tracer: tracer, client: client, endpoint: endpoint, timeout: cfg.jobTimeout,
		jobs: make(chan job, cfg.capacity), done: make(chan struct{}), runCtx: ctx, cancel: cancel}
	var wg sync.WaitGroup
	for i := 0; i < cfg.workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range p.jobs {
				p.process(id, j)
			}
		}(i)
	}
	go func() { wg.Wait(); close(p.done) }()
	return p, nil
}

// EN: Store only the parent IDs. Accepted jobs do not inherit request cancellation.
// RU: Сохраняем только ID родителя. Принятые задачи не наследуют отмену запроса.
func (p *pool) Submit(ctx context.Context, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return errPoolClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case p.jobs <- job{id: id, parent: trace.SpanContextFromContext(ctx)}:
		p.accepted.Add(1)
		return nil
	default:
		p.full.Add(1)
		return errQueueFull
	}
}

// EN: The same lock protects send and close, preventing send-on-closed-channel races.
// RU: Общая блокировка защищает отправку и close от гонки send-on-closed-channel.
func (p *pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.closed {
		p.closed = true
		close(p.jobs)
	}
}

// EN: Drain first. When the budget ends, cancel I/O and join workers before returning.
// Joining is cooperative: the downstream transport must honor its request context.
// RU: Сначала дожидаемся очереди. По окончании бюджета отменяем I/O и ждём воркеров.
// Ожидание совместное: downstream transport должен учитывать контекст запроса.
func (p *pool) Shutdown(ctx context.Context) error {
	p.Close()
	defer p.cancel()
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		p.cancel()
		<-p.done
		return ctx.Err()
	}
}

func (p *pool) stats() poolStats {
	return poolStats{accepted: p.accepted.Load(), completed: p.completed.Load(), failed: p.failed.Load(), full: p.full.Load()}
}

func (p *pool) process(workerID int, j job) {
	// EN: Job timeout starts on dequeue; time in the queue is a separate budget.
	// RU: Таймаут задачи начинается при извлечении; ожидание в очереди — отдельный бюджет.
	parent := trace.ContextWithSpanContext(p.runCtx, j.parent)
	ctx, cancel := context.WithTimeout(parent, p.timeout)
	defer cancel()
	ctx, span := p.tracer.Start(ctx, "worker.process", trace.WithAttributes(
		attribute.String("job.id", j.id), attribute.Int("worker.id", workerID)))
	defer span.End()
	if err := p.call(ctx); err != nil {
		outcome := "error"
		if errors.Is(err, context.DeadlineExceeded) {
			outcome = "deadline"
		}
		if errors.Is(err, context.Canceled) {
			outcome = "canceled"
		}
		span.SetAttributes(attribute.String("job.outcome", outcome))
		span.RecordError(err)
		span.SetStatus(codes.Error, outcome)
		p.failed.Add(1)
		return
	}
	span.SetAttributes(attribute.String("job.outcome", "completed"))
	p.completed.Add(1)
}

func (p *pool) call(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if _, err := io.Copy(io.Discard, io.LimitReader(resp.Body, 32<<10)); err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("downstream HTTP status / статус downstream HTTP: %d", resp.StatusCode)
	}
	return nil
}
