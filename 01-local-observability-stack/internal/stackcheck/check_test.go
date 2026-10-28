package stackcheck

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// EN: Model the real Jaeger v3 OTLP JSON format, including string intValue fields.
// RU: Модель реального OTLP JSON Jaeger v3, включая строковые значения intValue.
const storedTrace = `{"result":{"resourceSpans":[{"resource":{"attributes":[{"key":"service.name","value":{"stringValue":"demo"}}]},"scopeSpans":[{"spans":[{"traceId":"0123456789abcdef0123456789abcdef","spanId":"1111111111111111","parentSpanId":"2222222222222222","name":"GET /error","kind":2,"attributes":[{"key":"http.response.status_code","value":{"intValue":"503"}}],"status":{"code":2}},{"traceId":"0123456789abcdef0123456789abcdef","spanId":"3333333333333333","parentSpanId":"1111111111111111","name":"service.error","kind":1,"status":{"code":2}}]}]}]}}`

func expectedTrace() Expected {
	return Expected{TraceID: "0123456789abcdef0123456789abcdef", ParentID: "2222222222222222", Service: "demo", ServerName: "GET /error", ChildName: "service.error", HTTPStatus: 503}
}

func TestStoredTraceContract(t *testing.T) {
	for _, tc := range []struct{ name, old, replacement string }{
		{name: "complete trace"},
		{"wrong parent", `"parentSpanId":"1111111111111111"`, `"parentSpanId":"9999999999999999"`},
		{"wrong service", `"stringValue":"demo"`, `"stringValue":"other"`},
		{"lost child", `"name":"service.error"`, `"name":"different"`},
		{"lost trace context", `"traceId":"0123456789abcdef0123456789abcdef"`, `"traceId":"ffffffffffffffffffffffffffffffff"`},
		{"lost HTTP status", `"intValue":"503"`, `"intValue":"200"`},
		{"lost error status", `"code":2`, `"code":0`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := storedTrace
			if tc.old != "" {
				body = strings.ReplaceAll(body, tc.old, tc.replacement)
			}
			var data traceResponse
			if err := json.Unmarshal([]byte(body), &data); err != nil {
				t.Fatal(err)
			}
			err := validateTrace(data, expectedTrace())
			if (err == nil) != (tc.old == "") {
				t.Fatalf("stored trace contract / контракт сохранённой трассы: %v", err)
			}
		})
	}
}

func TestWaitForPartiallyStoredTrace(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/traces/"+expectedTrace().TraceID {
			t.Errorf("wrong trace endpoint / неверный адрес трассы: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if requests.Add(1) == 1 {
			_, _ = w.Write([]byte(`{"result":{"resourceSpans":[]}}`))
			return
		}
		_, _ = w.Write([]byte(storedTrace))
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := Wait(ctx, server.URL, []Expected{expectedTrace()}); err != nil {
		t.Fatal(err)
	}
	if requests.Load() < 2 {
		t.Fatal("partial trace was accepted / принята неполная трасса")
	}
}

func TestProbeCreatesIndependentSampledParents(t *testing.T) {
	var headers []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = append(headers, r.Header.Get("traceparent"))
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	defer server.Close()
	got, err := Probe(context.Background(), server.URL, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].TraceID == got[1].TraceID {
		t.Fatal("requests must have separate traces / нужны отдельные трассы запросов")
	}
	for i, want := range got {
		if headers[i] != "00-"+want.TraceID+"-"+want.ParentID+"-01" {
			t.Fatal("parent context not injected / родительский контекст не передан")
		}
	}
}
