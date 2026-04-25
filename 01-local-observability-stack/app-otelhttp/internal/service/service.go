package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	appName string
	tracer  trace.Tracer
}

func New(appName string) *Service {
	return &Service{
		appName: appName,
		tracer:  otel.Tracer("local-observability-stack/service"),
	}
}

func (s *Service) Health() map[string]any {
	return map[string]any{
		"status":  "ok",
		"service": s.appName,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}
}

func (s *Service) Work(ctx context.Context, delayMsRaw string) map[string]any {
	_, span := s.tracer.Start(ctx, "service.work")
	defer span.End()

	delayMs := parseDelayMs(delayMsRaw, 120)
	span.SetAttributes(attribute.Int("app.work.delay_ms", delayMs))

	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	return map[string]any{
		"service":   s.appName,
		"operation": "work",
		"delay_ms":  delayMs,
		"result":    "done",
	}
}

func (s *Service) Error(ctx context.Context, statusRaw string) (int, map[string]any) {
	_, span := s.tracer.Start(ctx, "service.error")
	defer span.End()

	status := parseStatusCode(statusRaw, 500)
	span.SetAttributes(attribute.Int("http.response.status_code", status))
	if status >= 500 {
		span.SetStatus(codes.Error, "forced error")
	}

	return status, map[string]any{
		"service": s.appName,
		"error":   fmt.Sprintf("forced error with status=%d", status),
	}
}

func parseDelayMs(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	delay, err := strconv.Atoi(value)
	if err != nil || delay < 0 {
		return defaultValue
	}
	if delay > 10000 {
		return 10000
	}

	return delay
}

func parseStatusCode(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	status, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	if status < 400 || status > 599 {
		return defaultValue
	}

	return status
}
