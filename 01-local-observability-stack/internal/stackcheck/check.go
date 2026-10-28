// Package stackcheck verifies stored HTTP traces against a real Jaeger backend.
// Пакет stackcheck проверяет сохранённые HTTP-трассы в настоящем Jaeger.
package stackcheck

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// EN: Expected identifies one request, including its simulated upstream parent.
// RU: Expected описывает запрос и его имитируемого внешнего родителя.
type Expected struct {
	TraceID    string
	ParentID   string
	Service    string
	ServerName string
	ChildName  string
	HTTPStatus int
}

type attribute struct {
	Key   string `json:"key"`
	Value struct {
		String  string          `json:"stringValue"`
		Integer json.RawMessage `json:"intValue"`
	} `json:"value"`
}

type span struct {
	TraceID    string          `json:"traceId"`
	SpanID     string          `json:"spanId"`
	ParentID   string          `json:"parentSpanId"`
	Name       string          `json:"name"`
	Kind       json.RawMessage `json:"kind"`
	Attributes []attribute     `json:"attributes"`
	Status     struct {
		Code json.RawMessage `json:"code"`
	} `json:"status"`
}

type traceResponse struct {
	Result struct {
		ResourceSpans []struct {
			Resource struct {
				Attributes []attribute `json:"attributes"`
			} `json:"resource"`
			ScopeSpans []struct {
				Spans []span `json:"spans"`
			} `json:"scopeSpans"`
		} `json:"resourceSpans"`
	} `json:"result"`
}

// EN: Probe sends a successful and a failing request with known sampled parents.
// RU: Probe отправляет успешный и ошибочный запрос с известными sampled-родителями.
func Probe(ctx context.Context, appURL, service string) ([]Expected, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	cases := []struct {
		path, child string
		status      int
	}{
		{"/work?delay_ms=10", "service.work", http.StatusOK},
		{"/error?status=503", "service.error", http.StatusServiceUnavailable},
	}
	expected := make([]Expected, 0, len(cases))
	for _, tc := range cases {
		var ids [24]byte
		if _, err := rand.Read(ids[:]); err != nil {
			return nil, err
		}
		want := Expected{
			TraceID: hex.EncodeToString(ids[:16]), ParentID: hex.EncodeToString(ids[16:]),
			Service: service, ChildName: tc.child, HTTPStatus: tc.status,
			ServerName: "GET " + strings.SplitN(tc.path, "?", 2)[0],
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(appURL, "/")+tc.path, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("traceparent", "00-"+want.TraceID+"-"+want.ParentID+"-01")
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("send request / отправка запроса: %w", err)
		}
		_, readErr := io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		closeErr := resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if resp.StatusCode != tc.status {
			return nil, fmt.Errorf("HTTP status / статус HTTP: got %d want %d", resp.StatusCode, tc.status)
		}
		expected = append(expected, want)
	}
	return expected, nil
}

// EN: Wait polls by trace ID until the full expected parent/child set is stored.
// RU: Wait опрашивает TraceID до появления полного набора родителя и потомка.
func Wait(ctx context.Context, queryURL string, expected []Expected) error {
	client := &http.Client{Timeout: 3 * time.Second}
	for _, want := range expected {
		endpoint := strings.TrimRight(queryURL, "/") + "/api/v3/traces/" + want.TraceID
		var lastErr error
		for {
			lastErr = readTrace(ctx, client, endpoint, want)
			if lastErr == nil {
				break
			}
			timer := time.NewTimer(200 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return fmt.Errorf("trace %s not verified / трасса не подтверждена: %w; last response / последний ответ: %v", want.TraceID, ctx.Err(), lastErr)
			case <-timer.C:
			}
		}
	}
	return nil
}

func readTrace(ctx context.Context, client *http.Client, endpoint string, want Expected) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Jaeger API status / статус API Jaeger: %d", resp.StatusCode)
	}
	var data traceResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&data); err != nil {
		return err
	}
	return validateTrace(data, want)
}

func validateTrace(data traceResponse, want Expected) error {
	var server, child *span
	for _, group := range data.Result.ResourceSpans {
		service := ""
		for _, a := range group.Resource.Attributes {
			if a.Key == "service.name" {
				service = a.Value.String
			}
		}
		if service != want.Service {
			continue
		}
		for _, scope := range group.ScopeSpans {
			for i := range scope.Spans {
				s := &scope.Spans[i]
				if s.TraceID != want.TraceID {
					continue
				}
				if s.Name == want.ServerName && enum(s.Kind) == 2 {
					server = s
				}
				if s.Name == want.ChildName {
					child = s
				}
			}
		}
	}
	if server == nil || child == nil {
		return fmt.Errorf("expected server and business span in service %q / ожидались серверный и бизнес-спан", want.Service)
	}
	if server.ParentID != want.ParentID || child.ParentID != server.SpanID {
		return fmt.Errorf("broken parent linkage / нарушена связь с родителем")
	}
	status := ""
	for _, a := range server.Attributes {
		if a.Key == "http.response.status_code" {
			status = strings.Trim(string(a.Value.Integer), "\"")
		}
	}
	if status != fmt.Sprint(want.HTTPStatus) {
		return fmt.Errorf("stored HTTP status / сохранённый HTTP-статус: got %s want %d", status, want.HTTPStatus)
	}
	wantCode := 0
	if want.HTTPStatus >= 500 {
		wantCode = 2
	}
	if enum(server.Status.Code) != wantCode || enum(child.Status.Code) != wantCode {
		return fmt.Errorf("unexpected stored span status / неожиданный статус сохранённого спана")
	}
	return nil
}

func enum(raw json.RawMessage) int {
	switch strings.Trim(string(raw), "\"") {
	case "", "0", "STATUS_CODE_UNSET":
		return 0
	case "1", "STATUS_CODE_OK", "SPAN_KIND_INTERNAL":
		return 1
	case "2", "STATUS_CODE_ERROR", "SPAN_KIND_SERVER":
		return 2
	default:
		return -1
	}
}
