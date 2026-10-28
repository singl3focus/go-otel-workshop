# Migration runs

[English](README.md) | [Русский](README.ru.md)

English is the original; the Russian translation is maintained alongside it.

## Run

```sh
go run .
```

## Expected trace

Three simulated SQL migrations run sequentially. Expect four spans in one trace: `migrations.run` and three `migration.apply` children. The root has `db.system=postgres`, `migrations.total=3`, and `migrations.applied=3`; children have `migration.name`.

The root retains the applied count if a migration fails. Normal completion leaves status Unset. Errors are recorded on the failed child and root. Shutdown flushes the batch processor before the short CLI exits.

Spans use the stdout batch exporter; allow its export delay or wait for normal
process completion. See [all job examples](../README.md).
