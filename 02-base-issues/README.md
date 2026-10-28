# 02 · Common tracing mistakes

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

Four exercises isolate common SDK mistakes. All dependencies are in this module's
`go.mod`; each subdirectory is a separate `package main`.

| Exercise | Repair | Test observes |
| --- | --- | --- |
| [missing-span-end](missing-span-end/README.md) | Finish the operation | One ended span with its attribute |
| [missing-tracerprovider-shutdown](missing-tracerprovider-shutdown/README.md) | Flush and shut down | One exported span and exporter shutdown |
| [mutate-after-end](mutate-after-end/README.md) | Record final state before End | `result=success` and Ok status |
| [new-root-context](new-root-context/README.md) | Keep incoming context | TraceID and parent/child linkage |

## Workflow

From this directory:

```sh
cd missing-span-end
go run .
go test .
go test -tags=solution .
```

`problem.go` demonstrates the mistake; `exercise.go` is your answer.
The ordinary test starts red. Repair the exercise, then compare with `solution.go`.
The `solution` build tag excludes your answer and runs the **same tests** on the
reference implementation. It does not turn `problem.go` into a corrected demo.

Module-wide reference check: `go test -tags=solution ./...`.
For shutdown tests, a retaining test exporter preserves the exported spans after
shutdown; the SDK's stock in-memory exporter clears them at shutdown.
