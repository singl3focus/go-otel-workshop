# OpenTelemetry for Go: a hands-on workshop

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

Learn distributed tracing by running small Go programs, breaking trace context,
and repairing instrumentation. Start with one span in your terminal, then follow
HTTP requests, background jobs, worker pools, and cross-service calls.

You need familiarity with Go, `context.Context`, and `net/http`. This workshop
focuses on **traces**; metrics and logs pipelines are outside its current scope.

## First span

Install Go **1.25.1 or newer**. The modules currently pin OTel SDK **1.43.0**
and, where used, `otelhttp` **0.68.0**. Initial runs download Go dependencies.

```sh
git clone https://github.com/singl3focus/go-otel-workshop.git
cd go-otel-workshop/00-otel-basics/01-minimal
go run .
```

The terminal prints a JSON span named `hello`, including its `TraceID`, `SpanID`,
start time, and end time. IDs and timestamps change on every run.
No Docker or observability backend is required for this first example.

## What you will learn

| Module | Material | Outcome |
| --- | --- | --- |
| [00 · Basics](00-otel-basics/README.md) | Ten small examples | Create spans; use attributes, events, resources, propagation, links, semantic conventions, and sampling |
| [01 · HTTP instrumentation](01-local-observability-stack/README.md) | Manual middleware alongside `otelhttp` | Understand the SDK setup, request spans, trace/log correlation, and OTLP export |
| [02 · Common mistakes](02-base-issues/README.md) | Four exercises with tests | Repair missing `End`, missing `Shutdown`, late mutations, and lost parents |
| [03 · In-process propagation](03-internal-context-propagation/README.md) | Three HTTP demos and a finite worker-lifecycle demo | Propagate context, bound queues, cancel I/O, and drain workers |
| [04 · Jobs and CLI](04-non-standard-cases/README.md) | Four runnable scenarios | Trace migrations, periodic reconciliation, batch reindexing, and cache warmup |
| [05 · Cross-service propagation](05-cross-service-propagation/README.md) | Two local HTTP servers and one exercise | Carry `traceparent` across an outbound HTTP request |
| [Deep dives](docs/README.md) | Explanations and troubleshooting | Understand span lifecycle, cancellation, and missing traces |

Follow **00 → 01 → 02 → 03 → 04 → 05**, or go directly to the problem you need.
Each numbered directory is an independent Go module. Run commands inside that
module or its example directory; there is no root `go.mod` or `go.work`.

## Practice and reference solutions

From the repository root:

```sh
cd 02-base-issues/missing-span-end
go run .
go test .
```

The exercise test **fails initially**. Edit `exercise.go`, rerun the test, then
compare with `solution.go`. `problem.go` remains a separate broken demonstration;
editing the exercise does not change `go run .`.

To validate the reference implementation without changing your answer:

```sh
go test -tags=solution .
```

The `solution` tag selects `solution.go` instead of `exercise.go`. The same
workflow applies to all four exercises in module 02 and the exercise in module 05.

## HTTP examples and export

Modules 00, 02, 03, 04, and 05 export to stdout. Module 01 uses **OTLP/gRPC**;
it includes two applications and a pinned [Collector + Jaeger Compose stack](01-local-observability-stack/stack/README.md).
Their default endpoint is `http://localhost:4317`, matching the local Collector.
The stack comes with readiness and real-backend test commands; a live stack run
is pending. Set `OTEL_EXPORTER_OTLP_ENDPOINT` when using another receiver.

Start with [module 01](01-local-observability-stack/README.md) for shell-specific
commands, or use [troubleshooting](docs/troubleshooting.md) for missing spans.
Most HTTP examples bind port 8080; module 05 also binds 8081. Run one at a time.

## Checks and contributions

For any module, `go build ./...` checks compilation and
`go test -tags=solution ./...` checks its tests against reference solutions.
Ordinary `go test ./...` remains red in untouched exercise modules by design.
The [CI workflow](.github/workflows/ci.yml) builds and tests every module and
checks documentation links. A local check does not prove that OTLP delivery works.

Found a confusing lesson or want to add a scenario? See
[CONTRIBUTING](CONTRIBUTING.md) and the [roadmap](ROADMAP.md).
All READMEs are available in English and Russian; code comments and exercise
diagnostics include both languages. Keep both versions in sync in the same change.

## License

Code and accompanying documentation are available under the [MIT License](LICENSE).
