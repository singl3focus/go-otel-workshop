package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// EN: This checks HTTP readiness only. Integration tests prove trace delivery.
// RU: Здесь проверяется только HTTP-готовность. Доставку трасс проверяют интеграционные тесты.
func main() {
	timeout := flag.Duration("timeout", 45*time.Second, "readiness timeout / таймаут ожидания готовности")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	endpoints := []struct{ name, url string }{
		{"Collector", env("WORKSHOP_COLLECTOR_HEALTH", "http://127.0.0.1:13133/")},
		{"Jaeger", env("WORKSHOP_JAEGER_HEALTH", "http://127.0.0.1:13134/status")},
		{"Jaeger UI", env("WORKSHOP_JAEGER_URL", "http://127.0.0.1:16686")},
	}
	for _, endpoint := range endpoints {
		if err := wait(ctx, endpoint.url); err != nil {
			log.Fatalf("%s not ready / не готов: %v", endpoint.name, err)
		}
		fmt.Printf("ready / готов: %s %s\n", endpoint.name, endpoint.url)
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func wait(ctx context.Context, endpoint string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	var lastErr error
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil {
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			err = fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		lastErr = err
		timer := time.NewTimer(200 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("%s: %w (last / последнее: %v)", endpoint, ctx.Err(), lastErr)
		case <-timer.C:
		}
	}
}
