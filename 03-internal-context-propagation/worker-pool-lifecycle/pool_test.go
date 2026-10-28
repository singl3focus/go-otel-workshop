package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return fn(r) }

func testTracer(t *testing.T) (*sdktrace.TracerProvider, *tracetest.SpanRecorder) {
	t.Helper()
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	return tp, sr
}

func stopPool(t *testing.T, p *pool) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = p.Shutdown(ctx)
}

func waitSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("synchronization timed out / таймаут синхронизации")
	}
}

func endedJobs(sr *tracetest.SpanRecorder) []sdktrace.ReadOnlySpan {
	var result []sdktrace.ReadOnlySpan
	for _, s := range sr.Ended() {
		if s.Name() == "worker.process" {
			result = append(result, s)
		}
	}
	return result
}

func outcome(s sdktrace.ReadOnlySpan) string {
	for _, a := range s.Attributes() {
		if a.Key == "job.outcome" {
			return a.Value.AsString()
		}
	}
	return ""
}

// EN: A real HTTP call outlives the canceled request; bounded admission and drain are observed.
// RU: Настоящий HTTP-вызов переживает отменённый запрос; проверяются лимит приёма и обработка очереди.
func TestAcceptedJobsSurviveRequestAndDrainBeforeTelemetryShutdown(t *testing.T) {
	tp, sr := testTracer(t)
	entered := make(chan struct{})
	release := make(chan struct{})
	var first sync.Once
	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		first.Do(func() { close(entered) })
		select {
		case <-release:
			w.WriteHeader(http.StatusNoContent)
		case <-r.Context().Done():
		}
	}))
	defer downstream.Close()
	p, err := newPool(poolConfig{workers: 1, capacity: 1, jobTimeout: 2 * time.Second}, tp.Tracer("test"), downstream.Client(), downstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer stopPool(t, p)
	parentCtx, parent := tp.Tracer("test").Start(context.Background(), "request.parent")
	requestCtx, cancelRequest := context.WithCancel(parentCtx)
	defer cancelRequest()
	if err := p.Submit(requestCtx, "active"); err != nil {
		t.Fatal(err)
	}
	waitSignal(t, entered)
	if err := p.Submit(requestCtx, "queued"); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(requestCtx, "overflow"); !errors.Is(err, errQueueFull) {
		t.Fatalf("queue admission / приём в очередь: %v", err)
	}
	parent.End()
	cancelRequest()
	p.Close()
	if err := p.Submit(context.Background(), "late"); !errors.Is(err, errPoolClosed) {
		t.Fatalf("closed admission / приём после закрытия: %v", err)
	}
	close(release)
	drain, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := p.Shutdown(drain); err != nil {
		t.Fatal(err)
	}
	jobs := endedJobs(sr)
	if len(jobs) != 2 {
		t.Fatalf("ended jobs / завершённые задачи: %d", len(jobs))
	}
	for _, s := range jobs {
		if s.Parent().SpanID() != parent.SpanContext().SpanID() || s.SpanContext().TraceID() != parent.SpanContext().TraceID() {
			t.Fatal("worker lost parent / воркер потерял родителя")
		}
		if s.Status().Code != codes.Unset || outcome(s) != "completed" {
			t.Fatalf("job should complete / задача должна завершиться: %v", s.Status())
		}
	}
	if got := p.stats(); got.accepted != 2 || got.completed != 2 || got.failed != 0 || got.full != 1 {
		t.Fatalf("pool totals / итоги пула: %+v", got)
	}
	// EN: At the actual shutdown call all worker spans have already ended.
	// RU: При реальном shutdown телеметрии все спаны воркеров уже завершены.
	if err := tp.Shutdown(drain); err != nil {
		t.Fatal(err)
	}
}

func TestJobDeadlineCancelsRealHTTP(t *testing.T) {
	tp, sr := testTracer(t)
	remoteCanceled := make(chan struct{})
	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		close(remoteCanceled)
	}))
	defer downstream.Close()
	p, err := newPool(poolConfig{workers: 1, capacity: 1, jobTimeout: 100 * time.Millisecond}, tp.Tracer("test"), downstream.Client(), downstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer stopPool(t, p)
	if err := p.Submit(context.Background(), "deadline"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := p.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	waitSignal(t, remoteCanceled)
	jobs := endedJobs(sr)
	if len(jobs) != 1 || jobs[0].Status().Code != codes.Error || outcome(jobs[0]) != "deadline" {
		t.Fatalf("expected one deadline error / ожидалась одна ошибка таймаута: %v", jobs)
	}
	if len(jobs[0].Events()) != 1 || jobs[0].Events()[0].Name != "exception" {
		t.Fatal("missing error event / нет события ошибки")
	}
}

func TestShutdownBudgetCancelsActiveAndQueuedJobs(t *testing.T) {
	tp, sr := testTracer(t)
	entered := make(chan struct{})
	var calls atomic.Int32
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		close(entered)
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	p, err := newPool(poolConfig{workers: 1, capacity: 1, jobTimeout: time.Hour}, tp.Tracer("test"), client, "http://downstream.local")
	if err != nil {
		t.Fatal(err)
	}
	defer stopPool(t, p)
	if err := p.Submit(context.Background(), "active"); err != nil {
		t.Fatal(err)
	}
	waitSignal(t, entered)
	if err := p.Submit(context.Background(), "queued"); err != nil {
		t.Fatal(err)
	}
	stop, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.Shutdown(stop); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected shutdown cancellation / ожидалась отмена остановки: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatal("queued canceled job started I/O / отменённая задача очереди начала I/O")
	}
	jobs := endedJobs(sr)
	if len(jobs) != 2 {
		t.Fatalf("unfinished job spans / незавершённые спаны: got %d", len(jobs))
	}
	for _, s := range jobs {
		if outcome(s) != "canceled" || s.Status().Code != codes.Error {
			t.Fatalf("cancel outcome / результат отмены: %v", s.Status())
		}
	}
}

func TestConcurrentSubmitAndClose(t *testing.T) {
	tp, sr := testTracer(t)
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}
	p, err := newPool(poolConfig{workers: 2, capacity: 4, jobTimeout: time.Second}, tp.Tracer("test"), client, "http://downstream.local")
	if err != nil {
		t.Fatal(err)
	}
	defer stopPool(t, p)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			err := p.Submit(context.Background(), "race")
			if err != nil && !errors.Is(err, errQueueFull) && !errors.Is(err, errPoolClosed) {
				t.Errorf("unexpected admission result / неожиданный результат приёма: %v", err)
			}
		}()
	}
	wg.Add(1)
	go func() { defer wg.Done(); <-start; p.Close() }()
	close(start)
	wg.Wait()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := p.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.stats(); got.completed != got.accepted || int64(len(endedJobs(sr))) != got.accepted {
		t.Fatalf("accepted jobs lost / принятые задачи потеряны: %+v", got)
	}
}

func TestRejectCanceledSubmission(t *testing.T) {
	tp, _ := testTracer(t)
	p, err := newPool(poolConfig{workers: 1, capacity: 1, jobTimeout: time.Second}, tp.Tracer("test"), http.DefaultClient, "http://unused.local")
	if err != nil {
		t.Fatal(err)
	}
	defer stopPool(t, p)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.Submit(ctx, "canceled"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled submission / отменённая отправка: %v", err)
	}
	if p.stats().accepted != 0 {
		t.Fatal("canceled request was accepted / отменённый запрос принят")
	}
}

func TestRejectInvalidLimits(t *testing.T) {
	tp, _ := testTracer(t)
	for _, cfg := range []poolConfig{{0, 1, time.Second}, {1, 0, time.Second}, {1, 1, 0}} {
		if _, err := newPool(cfg, tp.Tracer("test"), http.DefaultClient, "http://unused.local"); err == nil {
			t.Fatal("invalid limits accepted / приняты неверные лимиты")
		}
	}
}
