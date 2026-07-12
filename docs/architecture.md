# Architecture

## Conversion boundary

The pinned pgrust source already has a browser-oriented build and a `pgvfs`
host ABI. It deliberately targets `wasm64-unknown-unknown`, not WASI:

- PostgreSQL data structures retain their native 64-bit pointer and alignment
  layout.
- filesystem, stdin/stdout, argv, process exit, and wall clock operations are
  imported from the `pgvfs` module;
- the wasm entry point runs PostgreSQL's single-user backend;
- initdb remains a native build-time operation.

The first generated engine therefore uses this flow:

```text
pgrust wasm64 module
        │
        ├── pgvfs imports ──> Go memory filesystem and session streams
        │
        └── Binaryen memory64/table64 lowering
                    │
                    └── goccy/wasm2go ──> committed portable pure Go
```

Changing pgrust to `wasm32-wasip1` is not a viable shortcut. A trial build at
the pinned revision fails ABI assertions across PostgreSQL structs because
wasm32 changes pointer size and alignment. Those assertions protect real
in-memory contracts and must not be disabled.

## Cluster and session model

An `Instance` owns one filesystem tree containing an initialized PostgreSQL
18.3 cluster. Each database/sql connection owns a guest backend/session while
all sessions in an instance share that tree and cluster coordination state.
Separate instances receive separate trees.

Each database/sql connection keeps one single-user guest alive and feeds it SQL
through the pgvfs stdin stream. This preserves transaction and temporary-object
state across calls. The pool is currently limited to one connection because
pgrust's single-user backends do not yet expose shared-memory coordination for
concurrent sessions. The library does not claim isolation-suite compatibility
until that is implemented and measured by the upstream suite.

## Storage invariant

All `pgvfs` path operations are resolved relative to an instance-owned root in
memory. The host implementation has no fallback to `os.Open`, `os.CreateTemp`,
or a host data directory. PostgreSQL temporary files, WAL, lock files, and sort
spill use the same memory filesystem.

The initialized cluster template is produced during regeneration and embedded
read-only in the generated package. `New` clones it into a new writable memory
tree. This avoids running pgrust's subprocess-based initdb at application
runtime and keeps consumer builds independent of Rust and PostgreSQL tools.

## Release gates

1. The pinned wasm64 module builds reproducibly.
2. Binaryen lowers pgrust's Memory64/Table64 indices without changing its LP64
   data layout, and wasm2go translates the resulting module.
3. A generated module boots an embedded initdb fixture and executes a query. ✓
4. The database/sql driver passes lifecycle, transaction, cancellation, and
   concurrency tests without host filesystem writes.
5. PostgreSQL 18.3 regression and isolation expected outputs match upstream.
6. `CGO_ENABLED=0` builds succeed for Linux, macOS, and Windows on amd64/arm64.
