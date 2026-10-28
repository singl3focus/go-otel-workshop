package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"

	icp "github.com/singl3focus/go-otel-workshop/03-internal-context-propagation"
)

func main() {
	mode := flag.String("mode", "drain", "drain, timeout or cancel / обработать очередь, таймаут или отмена")
	flag.Parse()
	if err := run(*mode); err != nil {
		log.Fatal(err)
	}
}

// EN: A finite demo uses a real loopback HTTP server, with channels controlling progress.
// RU: Конечная демонстрация использует настоящий HTTP-сервер loopback и каналы для управления.
func run(mode string) (result error) {
	if mode != "drain" && mode != "timeout" && mode != "cancel" {
		return fmt.Errorf("unknown mode / неизвестный режим: %s", mode)
	}
	tp := icp.NewTracerProvider("example-worker-pool-lifecycle")
	defer func() {
		flush, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		result = errors.Join(result, tp.Shutdown(flush))
	}()
	tracer := tp.Tracer("03-internal-context-propagation/worker-pool-lifecycle")
	rootCtx, root := tracer.Start(context.Background(), "demo.run")
	defer root.End()

	started := make(chan struct{})
	release := make(chan struct{})
	var first sync.Once
	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		first.Do(func() { close(started) })
		select {
		case <-release:
			w.WriteHeader(http.StatusNoContent)
		case <-r.Context().Done():
		}
	}))
	defer downstream.Close()
	timeout := 2 * time.Second
	if mode == "timeout" {
		timeout = 50 * time.Millisecond
	}
	p, err := newPool(poolConfig{workers: 1, capacity: 1, jobTimeout: timeout}, tracer, downstream.Client(), downstream.URL)
	if err != nil {
		return err
	}
	defer func() {
		cleanup, cancel := context.WithCancel(context.Background())
		cancel()
		_ = p.Shutdown(cleanup)
	}()

	parentCtx, parent := tracer.Start(rootCtx, "request.enqueue")
	requestCtx, cancelRequest := context.WithCancel(parentCtx)
	defer cancelRequest()
	if err := p.Submit(requestCtx, "first"); err != nil {
		parent.End()
		return err
	}
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		parent.End()
		return errors.New("downstream did not start / downstream не запустился")
	}
	if err := p.Submit(requestCtx, "queued"); err != nil {
		parent.End()
		return err
	}
	if err := p.Submit(requestCtx, "overflow"); !errors.Is(err, errQueueFull) {
		parent.End()
		return fmt.Errorf("expected queue rejection / ожидался отказ очереди: %v", err)
	}
	parent.SetAttributes(attribute.Int("queue.accepted", 2), attribute.Int("queue.full", 1))
	parent.End()
	cancelRequest()

	p.Close()
	if err := p.Submit(context.Background(), "after-close"); !errors.Is(err, errPoolClosed) {
		return fmt.Errorf("expected closed pool / ожидался закрытый пул: %v", err)
	}
	if mode == "drain" {
		close(release)
	}
	stop, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if mode == "cancel" {
		cancel()
	}
	shutdownErr := p.Shutdown(stop)
	if mode == "cancel" {
		if !errors.Is(shutdownErr, context.Canceled) {
			return fmt.Errorf("expected shutdown cancellation / ожидалась отмена остановки: %v", shutdownErr)
		}
	} else if shutdownErr != nil {
		return shutdownErr
	}
	totals := p.stats()
	root.SetAttributes(attribute.String("demo.mode", mode), attribute.Int64("jobs.completed", totals.completed), attribute.Int64("jobs.failed", totals.failed))
	fmt.Printf("mode=%s accepted=%d completed=%d failed=%d queue_full=%d after_close=closed\n",
		mode, totals.accepted, totals.completed, totals.failed, totals.full)
	return nil
}
