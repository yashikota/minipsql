# SpiderMonkey flow and the pgrust adaptation

The reference implementation is a four-repository pipeline, not merely a
successful invocation of the wasm2go CLI.

1. `spidermonkey-wasm` links the prebuilt engine archive with ThinLTO and runs
   Binaryen optimization.
2. `protoc-gen-wasmify-go` invokes wasm2go and emits both the typed API bridge
   and the transpiled engine bundle.
3. `spidermonkeywasm2go` distributes the engine as an independent module:
   `base`, `p0` through `p4`, `data.bin`, asm trampolines, and pure-Go fallbacks.
4. `go-spidermonkey` wraps a generated single-module API in an instance-owned
   runtime. Each interpreter owns its linear memory and host state.

The properties we reuse for minipsql are:

- optimize before transpilation;
- one generated module instance per isolated database;
- a small hand-written public API over generated bindings;
- sidecar initialization data rather than giant Go byte literals;
- chained packages to keep the Go compiler's peak memory bounded;
- the portable pure-Go fallback for pgrust's substantially larger bundle;
- a separate generated-engine module so ordinary users do not run generation.

The material difference is the optimized input size. SpiderMonkey reaches
6,892 functions and 3,579 table entries. pgrust currently reaches 43,655
functions and 21,840 table entries even after ThinLTO and Binaryen. Because
indirect table roots dominate, export-only DCE cannot bridge that gap.

Therefore minipsql extends the proven output architecture instead of replacing
it. The committed wasm2go patches format each completed chunk before lowering
the next one and expose the existing pure-Go fallback for very large modules.
This reduced measured pgrust translation RSS from 8.17 GiB to about 2.48 GiB
and completes generation without gcasm's dual-architecture listings. pgrust's
LP64 module is lowered with Binaryen's Table64 and Memory64 passes, allowing the
established memory32 backend to be reused without changing PostgreSQL layouts.
The 33 chained packages keep normal Go compilation bounded and cross-platform.
