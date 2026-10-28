# Troubleshooting

[English](troubleshooting.md) | [Русский](troubleshooting.ru.md)

English is the original; the Russian translation is maintained alongside it.

## Go says there is no main module

The repository has six Go modules and no root module. Change into the numbered
module, then run `go build ./...` or a command from its README. Use Go 1.25.1 or
newer. Downloads need network access on the first run; errors from `GOPROXY`
or a module cache are environment failures, not proof of an exercise mistake.

## Tests fail immediately

All five exercise implementations start broken. `go test .` checks your answer;
`go test -tags=solution .` selects the reference. Run from the exercise directory.
If reference tests fail, record the exact command, Go version, and output in an issue.
Do not expect `go run .` to call the exercise function: it runs `problem.go`.

## A span is missing from stdout

Check that the provider is configured, sampling permits recording, `End()` is
called, and the batch processor has exported. Short CLIs must shut down/flush.
Example 00/05 deliberately omits shutdown, and module 02 deliberately demonstrates
missing End. A dropped sampling decision is not changed by `RecordError`.
An ended child can export even if its parent never ends.

## A trace splits into different TraceIDs

Inside a process, pass the returned context from `tracer.Start` down the chain.
Across HTTP, instrument both the server handler and outbound transport, and
configure a compatible propagator. Request context alone does not write headers.
Inspect `traceparent` at both ends. Module 05 demonstrates this failure.

## Background work reports context canceled

Returning from `ServeHTTP` cancels its request context. For independent work,
detach cancellation and supply a bounded lifetime owned by the worker, or carry
only SpanContext in the job. `defer cancel()` in the handler cancels the worker's
context too. `time.Sleep` does not honor context cancellation.
See the [deep dive](internal-context-propagation/detached-background-context.md).

## HTTP works but module 01 exports no traces

The default `http://localhost:4317` needs a running receiver. Start the
[local stack](../01-local-observability-stack/stack/README.md), or configure an existing receiver using the shell commands
in [module 01](../01-local-observability-stack/README.md).
The app uses OTLP/gRPC; an OTLP/HTTP endpoint on 4318 is a different protocol.
An application inside a container has its own localhost. Verify the receiver's
configured bind address and routing, not only that a process/container is running.
Read asynchronous exporter errors and the receiver's logs. HTTP 200 and nonempty
trace IDs in app logs do not demonstrate delivery to a backend.

## The port is busy or curl behaves differently

Run one demo at a time. Most HTTP examples use 8080; module 05 also needs 8081.
Module 01 supports `HTTP_PORT`. In Windows PowerShell, use `curl.exe` to avoid
the `Invoke-WebRequest` alias. Quote URLs containing query parameters.

## The disk is full during a build

Go uses a temporary build directory and a cache. Free disk space before rerunning;
do not classify a filesystem error as a code or test failure. Keep local reports,
module caches, and credentials out of commits.
