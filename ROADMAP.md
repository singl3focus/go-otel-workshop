# Workshop roadmap

[English](ROADMAP.md) | [Русский](ROADMAP.ru.md)

English is the original; the Russian translation is maintained alongside it.

This is a direction for future lessons, not a claim that they already exist.
The current supported path is described in the [README](README.md).

## Available in this update

- English originals and Russian translations across all lesson READMEs and deep dives.
- Bilingual source explanations and exercise diagnostics.
- A deterministic head-sampling lesson with parent decisions and error limitations.
- Executable reference solutions, cancellation regression checks, and per-module CI.
- Troubleshooting and contributor instructions.
- Finite worker-pool lifecycle demo: drain, job timeout, pool cancellation, and concurrent admission tests.
- Pinned Collector/Jaeger stack, HTTP readiness command, and opt-in stored-trace checks (live run pending).

## Next learning increments

| Priority | Topic | Concrete completion criterion |
| --- | --- | --- |
| Verify | Collector + trace UI | Configuration and tests are included; execute the full live check and capture a real trace screenshot |
| Available | Context-aware jobs | See the [worker lifecycle demo](03-internal-context-propagation/worker-pool-lifecycle/README.md): bounded admission, HTTP deadlines, worker draining, and race tests |
| Later | Database and RPC boundaries | One SQL or gRPC example with trace assertions and explicit external-service requirements |
| Later | Asynchronous messaging | Message carrier propagation, retries, and a reasoned parent-versus-link choice |
| Later | Tail sampling and sensitive data | A Collector exercise demonstrating a stated retention policy and attribute filtering |
| Later | Metrics and logs | Separate lessons clearly distinguished from this trace-focused path |

Prefer a complete small lesson with expected output over many untested scenarios.
The local backend is prepared; its live verification remains the first follow-up.

## Publication and reach

After local verification, publish the update, wait for the first GitHub Actions run,
and describe the actual runnable lessons in a release announcement. Keep About
and topics accurate. Consider a social preview and sharing one reproducible
scenario in relevant Go/OTel communities; publication and outreach need a separate
maintainer action.

Use feedback, completed exercises, issues, and contributions to evaluate usefulness.
Views/clones can help evaluate discovery when repository traffic data is available.
More stars are a possible outcome, not a promised result of translation.
