# minipsql

`minipsql` is an experimental, embedded PostgreSQL 18.3-compatible database
for Go tests. It is built from [pgrust](https://github.com/malisper/pgrust),
compiled to WebAssembly, and translated to cgo-free Go with
[wasm2go](https://github.com/goccy/wasm2go).

The non-negotiable design constraints are:

- every database cluster is isolated and stored in memory;
- the runtime never falls back to a host data directory or temporary file;
- consumers build with `CGO_ENABLED=0` and do not need Rust or a Wasm runtime;
- Linux, macOS, and Windows on amd64 and arm64 are supported;
- PostgreSQL 18.3 regression and isolation tests are the compatibility oracle.

## Status

The end-to-end engine is working. pgrust is compiled as
`wasm64-unknown-unknown` to preserve PostgreSQL's LP64 ABI, optimized with
ThinLTO/Binaryen, then passed through Binaryen's Table64 and Memory64 lowering
passes. This retains the LP64 guest data layout while producing the WebAssembly
1.0 memory/table instruction shape consumed by goccy/wasm2go. wasm2go lowers
43,648 reachable functions to committed, portable pure Go.

The public `database/sql` API boots an embedded PostgreSQL 18 cluster entirely
in memory. `SELECT 1`, parameterized statements, persistent sessions, and
transactions are covered by integration tests. Full PostgreSQL regression and
isolation-suite compatibility remains the next release gate; it is not yet
claimed.

## Intended API

```go
instance, err := minipsql.New(ctx, minipsql.Options{})
if err != nil {
    t.Fatal(err)
}
t.Cleanup(func() { _ = instance.Close() })

db := instance.DB()
```

`Instance.DB()` is a regular `*sql.DB`. Connections opened by the pool share
one cluster; separate instances never share state.

## Development

Requirements for regeneration are Rust nightly, `rust-src`, Binaryen's
`wasm-opt`, and Go. The pinned upstream revisions are in `tools/versions.env`.

```sh
task fetch
task build:wasm
task feasibility
task generate:go
task generate:fixture
task test
task check:cgo
```

The local pipeline defaults to pgrust's `wasm-prod` profile with ThinLTO and 16
codegen units, followed by Binaryen optimization and 64-to-32-bit memory/table
instruction lowering. The pre-lowered LP64 module is retained as a build
artifact for diagnostics. `generate:fixture` runs pgrust's native build-time
initdb and packs both the cluster and its share/timezone tree into the embedded
read-only template; runtime instances clone that template into Go memory.

Generated Go source is committed, so these tools are not required by
library consumers.

## License

This project is distributed under AGPL-3.0 because it incorporates and adapts
pgrust. Dependency and generated-source notices will be retained with the
generated engine.
