package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitRetriesStartingService(t *testing.T) {
	var count atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if count.Add(1) < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := wait(ctx, server.URL); err != nil {
		t.Fatal(err)
	}
	if count.Load() < 2 {
		t.Fatal("no readiness retry / не было повторной проверки готовности")
	}
}

func TestWaitReportsDeadlineForUnavailableService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	if err := wait(ctx, server.URL); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected readiness deadline / ожидался таймаут готовности: %v", err)
	}
}
