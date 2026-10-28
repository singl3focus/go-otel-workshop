# Head sampling and parent decisions

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

## Run

```sh
go run .
```

This deterministic example uses an explicit provider and an in-memory exporter.
Each row prints whether the span is recording, its sampled flag, and the number
of exported spans. No backend or environment variables are needed.

| Case | Recording | Sampled | Exported |
| --- | --- | --- | --- |
| AlwaysSample root | true | true | 1 |
| NeverSample root | false | false | 0 |
| ParentBased(NeverSample), root | false | false | 0 |
| Same sampler, sampled remote parent | true | true | 1 |
| Same sampler, unsampled remote parent | false | false | 0 |

`ParentBased` uses the delegate for roots and honors a parent's decision by
default. In practice, a root delegate can be `TraceIDRatioBased(0.1)`; short runs
do not guarantee exactly 10% sampled traces, so this lesson uses fixed decisions.
An unsampled span still carries a valid TraceID/SpanID for propagation.

Head sampling decides at span creation. Calling `RecordError` later does not
resurrect a dropped span. Collecting traces first and deciding using completed
trace data is a separate tail-sampling design, outside this example.

Source: [OTel Go sampling API](https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#ParentBased).
