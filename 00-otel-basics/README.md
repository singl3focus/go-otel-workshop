# 00 · OpenTelemetry basics

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

Ten runnable programs introduce one tracing concept at a time. Each directory
contains its own `package main`; dependencies are shared through this module's
`go.mod`. No backend is needed: spans go to `stdouttrace`.

| Example | Concept | What to look for |
| --- | --- | --- |
| [01-minimal](01-minimal/) | Exporter → provider → tracer → span | One ended `hello` span |
| [02-attributes-events-status](02-attributes-events-status/) | Attributes, events, errors, status | `ok.span` and `bad.span`; error event and Error status are separate |
| [03-nested-spans](03-nested-spans/) | Parent/child context | Same TraceID; child's parent is the parent's SpanID |
| [04-resource](04-resource/) | Resource metadata | `service.name=basics-resource-demo` |
| [05-batcher-vs-syncer](05-batcher-vs-syncer/) | Export buffering | `buffered.span` can be lost without shutdown |
| [06-propagator-mapcarrier](06-propagator-mapcarrier/) | Inject/Extract | A `traceparent` carrier and matching TraceIDs |
| [07-links-vs-parent](07-links-vs-parent/) | Links versus hierarchy | A child shares the trace; a linked span starts another |
| [08-global-vs-explicit-tracerprovider](08-global-vs-explicit-tracerprovider/) | Provider ownership | Distinct service resources for explicit and global providers |
| [09-semconv-vs-manual](09-semconv-vs-manual/) | Attribute conventions | Handwritten mistakes export, typed conventions avoid them |
| [10-sampling](10-sampling/README.md) | Head sampling and parent decisions | Recording, sampling flags, and exported span counts |

From this directory:

```sh
go run ./01-minimal
go run ./10-sampling
```

Use any example's directory with `go run .` instead. Stdout export order follows
span completion, not the trace tree. A parent can be printed after its child.
Successful spans often have **Unset**, not Ok, unless the code sets it explicitly.

Try changing one attribute or removing a parent context, then compare the IDs.
Continue with [HTTP instrumentation](../01-local-observability-stack/README.md)
or [common mistakes](../02-base-issues/README.md).
