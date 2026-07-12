# Feasibility report

Measurements below use the revisions in `tools/versions.env` on Linux/amd64.

## Confirmed

- A direct `wasm32-wasip1` build is invalid for pgrust. It fails 29 compile-time
  PostgreSQL ABI layout assertions because wasm32 is ILP32 and pgrust's current
  port preserves LP64 layouts.
- The upstream `wasm64-unknown-unknown` build succeeds with the documented
  host-import model. The build script supplies empty ICU archive names required
  by wasm-ld and keeps actual symbols unresolved with `--allow-undefined`.
- The unoptimized-for-transpilation `wasm-boot` artifact was 171.6 MiB and
  contained:
  - 147,673 defined functions;
  - 225 host imports;
  - one Memory64 linear memory starting at 138 pages / 8 MiB;
  - 23,056 table entries;
  - 7,657,688 bytes of initialized data.
- The committed wasm2go patch parses the Memory64 limits flag and `-dump`
  correctly reports `memory64=true` for this artifact.

## SpiderMonkey control measurement

The released `spidermonkey.wasm` is not a raw representation of the source
tree. The successful pipeline first performs a ThinLTO link and Binaryen
optimization, then invokes wasm2go through `protoc-gen-wasmify-go`. The released
input to wasm2go is 6.1 MiB and contains 6,892 defined functions and 3,579 table
entries. wasm2go emits a `base` package plus five chained packages (`p0` through
`p4`), with gcasm output on amd64/arm64 and pure-Go fallbacks on other targets.

This means source-tree size is not the useful comparison. The relevant values
are the optimized wasm function-body size, retained indirect-call table, and
the size of each generated Go package.

## pgrust optimized measurement

Using the same ordering as the SpiderMonkey pipeline:

1. `wasm-prod` with ThinLTO and 16 codegen units;
2. `wasm-opt -O2`, with Memory64/bulk-memory enabled and SIMD generation
   disabled;
3. wasm2go with `main` as the explicit export root.

produces a 37.0 MiB module with:

- 43,655 defined functions;
- 224 host imports;
- 21,840 table entries;
- 6,540,816 bytes of initialized data.

This is a 4.6x reduction in bytes and a 3.4x reduction in functions from the
bring-up artifact. ThinLTO completes on the 12 GiB development host where fat
LTO was killed. Binaryen peaks at about 2.7 GiB RSS.

The optimized non-SIMD module enters wasm2go's existing multi-package/gcasm
pipeline successfully. DCE removes only 3 functions because Rust's indirect
function table retains most of the program. Unmodified wasm2go then reaches
about 8.17 GiB RSS and is killed before serialization.

The second wasm2go patch preserves the SpiderMonkey `base`/`pN`/gcasm output
layout but formats each completed chunk and releases its AST before lowering
the next chunk. With that patch, the same translation:

- lowers all 43,652 reachable functions;
- reduces peak RSS to about 2.1 GiB;
- reaches the gcasm Go compiler capture stage;
- reports 10,259,268 SSA instructions reduced to 7,431,768;
- stops at the first concrete Memory64 type mismatch: bulk-memory helpers still
  accept i32 addresses and lengths while the guest supplies i64 values.

## Binaryen Memory64/Table64 lowering

Building pgrust directly for `wasm32-wasip1` remains invalid: doing so changes
the Rust/C ABI to ILP32 and fails PostgreSQL's compile-time layout assertions.
Binaryen's lowering passes operate after compilation and have different
semantics. They preserve the LP64 struct layout and i64 values used as guest
pointers, but rewrite Memory64/Table64 index operations to the 32-bit WebAssembly
forms. This is valid for minipsql while linear memory remains below 4 GiB.

`--memory64-lowering` alone is insufficient because rustc's wasm64 target also
emits a 64-bit table. The working sequence is:

```sh
wasm-opt \
  --all-features --disable-simd --disable-relaxed-simd \
  --table64-lowering --memory64-lowering -O2 \
  pgrust-postgres.memory64.wasm \
  -o pgrust-postgres.wasm
```

The lowered artifact is 39.1 MiB, contains 43,651 defined functions and 21,840
table entries, and is reported by wasm2go as `memory64=false`. With the bounded
AST patch, goccy wasm2go lowers all 43,648 reachable functions and passes the
previous i64 bulk-memory helper mismatch. The next high-water mark is gcasm's
`go build -gcflags=-S` capture: its monolithic `CombinedOutput` retains several
gigabytes of assembly listing. Streaming that listing removes the immediate
buffer duplication, but retaining both architecture captures still peaks near
5.9 GiB for pgrust.

The production path therefore selects wasm2go's existing pure-Go fallback with
`WASM2GO_PURE_GO=1`. Full generation completes in 1m42s at about 2.48 GiB RSS,
emits 33 chained packages plus `base` and `data.bin` (about 387 MiB of source),
and compiles successfully. The generated engine boots the embedded PostgreSQL
18 fixture and executes `SELECT 1`, transactions, and parameterized statements
through database/sql.

## ncruces/wasm2go

[`ncruces/wasm2go`](https://github.com/ncruces/wasm2go) is a separate AOT
Wasm-to-Go translator, not a runtime. Its current upstream explicitly supports
64-bit address spaces, bulk memory, reference types, multi-value results,
atomics, and other post-MVP features. It emits a self-contained Go package with
no non-standard-library runtime dependency.

It is useful here as a semantic reference for Memory64 and for differential
testing small extracted fixtures. It is not the primary generator because it
emits one Go source file, while the measured pgrust input retains more than 43k
functions. The goccy pipeline already provides the SpiderMonkey-proven chained
`base`/`pN` package layout, gcasm acceleration, pure-Go fallbacks, and wasmify
bridge integration. The project therefore remains based on goccy/wasm2go, with
Binaryen lowering removing Memory64 from its critical path.

## Full-translation measurement (obsolete input)

The following trial was intentionally stopped before the host was exhausted:

```sh
go run ./cmd/wasm2go \
  -i pgrust-postgres.wasm \
  -pkg pgrust \
  -import github.com/yashikota/minipsql/internal/generated/pgrust \
  -out-dir /tmp/minipsql-generated \
  -entry-exports main
```

- Reachability removed only 4 of 147,673 functions because the entry point and
  indirect function table retain almost the entire program.
- RSS reached 8.4 GiB and all 4 GiB of swap was consumed before output files
  were emitted.
- wasm2go's multi-package emitter first compiles every chunk into retained Go
  AST (`translateLinknameMulti`, step 1) and serializes chunks only afterwards.
  This is suitable for the 6.9k-function SpiderMonkey bundle but is the dominant
  peak-memory term for pgrust.
- This measurement used the wrong pre-optimization input and is retained only
  as a baseline. It is not the production generation path.

## Next compatibility work

1. Adapt the upstream pgrust/PostgreSQL regression runner to the persistent
   single-user stream and record per-file expected-output diffs.
2. Replace optional ICU/libxml/OpenSSL no-op imports as their regression groups
   are enabled.
3. Add shared coordination before allowing more than one database/sql
   connection per instance, then run the upstream isolation suite.
