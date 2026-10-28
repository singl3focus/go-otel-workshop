# Contributing

[English](CONTRIBUTING.md) | [Русский](CONTRIBUTING.ru.md)

English is the original; the Russian translation is maintained alongside it.

Help learners understand tracing by contributing a small, runnable example,
clearer explanation, bug report, or translation correction.

## Local checks

Each numbered directory is an independent Go module. From its directory:

```sh
go build -mod=readonly ./...
go test -mod=readonly -count=1 -timeout=60s -tags=solution ./...
```

On a platform with a C toolchain, also run the race detector:

```sh
go test -mod=readonly -race -count=1 -timeout=60s -tags=solution ./...
```

Ordinary exercise tests fail before the participant fixes them. Do not commit
answers into `exercise.go` just to make those tests green. `-tags=solution`
selects the reference implementation; the CI tests that path.
When changing an exercise, confirm the untouched student version still exposes
the intended mistake and the reference passes the same assertions.

From the repository root, check documentation:

```sh
python scripts/check_docs.py
```

The script checks local Markdown targets, heading anchors, language switches,
and matching runnable code blocks in paired READMEs. It does not verify external
links or semantic translation quality. CI runs it on Linux from a clean checkout.

Format changed Go files with `gofmt`. The root `.golangci.yml` is optional local
tooling: run the installed golangci-lint from inside each Go module with
`golangci-lint run --config ../.golangci.yml`. That configuration currently
enables fixes; review the resulting diff. It is not part of the CI contract.

## English originals and Russian translations

- `README.md` is the English original; `README.ru.md` is its Russian counterpart.
- Keep the language switch near the top and link Russian navigation to Russian files.
- Update both versions in the same pull request when changing commands, prerequisites,
  lesson order, or expected results. Match runnable code blocks exactly.
- Use an English explanation followed by Russian in shared source comments (`EN:`/`RU:`).
  Keep build tags, linter directives, identifiers, span names, and attribute keys intact.
- Exercise diagnostics should explain a failure in both languages. Stable technical
  fields such as `TraceID`, `SpanID`, `got`, and `want` need no translated identifier.
- Translations must preserve the technical meaning. Explain a real mismatch instead
  of blindly reproducing an incorrect statement from either language.

## A useful new lesson

Give it one learning objective, an exact command, observable trace properties,
and a short explanation of how it differs from adjacent lessons. Prefer stdout
and in-process test servers when the topic does not require an external backend.
State any required ports and services. For an exercise, provide `problem.go`,
an initially broken `exercise.go`, tests, and a `solution.go` behind the solution tag.

Keep demonstrations intentionally small. Distinguish deliberate anti-patterns
from bugs in the reference implementation. Avoid claims such as "all errors are
sampled" or "the timeout stops all work" unless the example actually establishes them.

Open a focused pull request with the problem, resulting learner behavior,
and commands you ran. English or Russian issues and discussion are welcome.
Be respectful and keep feedback specific to the work.

## License

This project uses the [MIT License](LICENSE), including its accompanying documentation.
Contributions are made under the same license.

## Local backend integration

Follow the [local stack guide](01-local-observability-stack/stack/README.md).
The `integration` tag requires live Collector and Jaeger; ordinary tests stay self-contained.
The dedicated CI job checks actual stored traces, not just readiness responses.
