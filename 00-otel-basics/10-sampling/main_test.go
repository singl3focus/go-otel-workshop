package main

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestHeadSamplingAndParentDecision(t *testing.T) {
	tests := []struct {
		name        string
		sampler     sdktrace.Sampler
		parent      context.Context
		wantSampled bool
	}{
		{"always root", sdktrace.AlwaysSample(), context.Background(), true},
		{"never root", sdktrace.NeverSample(), context.Background(), false},
		{"parent based root", sdktrace.ParentBased(sdktrace.NeverSample()), context.Background(), false},
		{"sampled remote parent", sdktrace.ParentBased(sdktrace.NeverSample()), remoteParent(true), true},
		{"unsampled remote parent", sdktrace.ParentBased(sdktrace.NeverSample()), remoteParent(false), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := observe(tc.parent, tc.sampler)
			if err != nil {
				t.Fatal(err)
			}
			if !got.spanContext.IsValid() {
				t.Fatal("invalid propagation context / невалидный контекст propagation")
			}
			if got.recording != tc.wantSampled || got.spanContext.IsSampled() != tc.wantSampled {
				t.Fatalf("sampling decision / решение sampling: recording=%t sampled=%t want=%t",
					got.recording, got.spanContext.IsSampled(), tc.wantSampled)
			}
			wantExported := 0
			if tc.wantSampled {
				wantExported = 1
			}
			if got.exported != wantExported {
				t.Fatalf("exported spans / экспортированные спаны: got %d want %d", got.exported, wantExported)
			}
			parent := trace.SpanContextFromContext(tc.parent)
			if parent.IsValid() && got.spanContext.TraceID() != parent.TraceID() {
				t.Fatal("parent TraceID lost / потерян TraceID родителя")
			}
		})
	}
}
