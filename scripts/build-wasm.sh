#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cache_dir=${MINIPSQL_CACHE_DIR:-"$repo_root/.cache"}
pgrust_dir="$cache_dir/pgrust"
output_dir="$repo_root/dist/upstream"
profile=${PGRUST_WASM_PROFILE:-wasm-prod}
target_dir=${CARGO_TARGET_DIR:-"$cache_dir/cargo-target/pgrust"}
stub_dir="$cache_dir/wasm-native-stubs"
raw_wasm="$output_dir/pgrust-postgres.raw.wasm"
memory64_wasm="$output_dir/pgrust-postgres.memory64.wasm"
wasm="$output_dir/pgrust-postgres.wasm"

if [ ! -d "$pgrust_dir/.git" ]; then
    "$repo_root/scripts/fetch-upstream.sh"
fi

rustup component add rust-src --toolchain nightly

# Some pgrust crates retain native ICU link directives even though the
# single-user wasm path never calls them. wasm-ld still requires the named
# archives to exist before --allow-undefined can preserve their symbols as
# imports. Empty deterministic archives satisfy that lookup.
mkdir -p "$stub_dir"
llvm_ar=$(command -v llvm-ar || command -v ar)
for library in icui18n icuuc icudata; do
    "$llvm_ar" crs "$stub_dir/lib$library.a"
done

# SpiderMonkey's successful large-module pipeline links with ThinLTO before
# handing the module to Binaryen and wasm2go. Rust's `lto = true` selects fat
# LTO and exceeds the memory available on ordinary CI/developer machines for
# pgrust, while ThinLTO preserves the important whole-program elimination and
# completes within the same resource class. Callers can still override either
# setting explicitly.
PGRUST_PGSHAREDIR=/usr/local/pgsql/share \
CARGO_PROFILE_WASM_PROD_LTO=${CARGO_PROFILE_WASM_PROD_LTO:-thin} \
CARGO_PROFILE_WASM_PROD_CODEGEN_UNITS=${CARGO_PROFILE_WASM_PROD_CODEGEN_UNITS:-16} \
CARGO_TARGET_DIR="$target_dir" cargo +nightly rustc \
    --manifest-path "$pgrust_dir/Cargo.toml" \
    -Zbuild-std=std,panic_abort \
    --locked \
    --package init \
    --bin postgres \
    --target wasm64-unknown-unknown \
    --profile "$profile" \
    -- \
    -L "native=$stub_dir" \
    -C link-arg=--allow-undefined

mkdir -p "$output_dir"
cp "$target_dir/wasm64-unknown-unknown/$profile/postgres.wasm" \
    "$raw_wasm"

# wasm2go currently has no SIMD lowering. `--all-features` lets Binaryen read
# Memory64, bulk-memory and non-trapping conversions used by rustc; the two
# explicit disables prevent -O2 from introducing SIMD into an otherwise scalar
# input module. This mirrors wasmify's optimize-before-transpile ordering.
wasm-opt -O2 \
    --all-features \
    --disable-simd \
    --disable-relaxed-simd \
    --strip-debug \
    "$raw_wasm" \
    -o "$memory64_wasm"

# pgrust must be compiled with an LP64 ABI: compiling the source as wasm32
# breaks PostgreSQL's layout assertions. Binaryen lowering is different: it
# keeps the already-compiled 64-bit C/Rust data layout and narrows only the
# linear-memory/table index instructions. Both passes are required because the
# wasm64 Rust target emits a 64-bit table as well as Memory64. The resulting
# module stays below wasm32's 4 GiB address-space limit and can use wasm2go's
# mature memory32 codegen path.
wasm-opt \
    --all-features \
    --disable-simd \
    --disable-relaxed-simd \
    --table64-lowering \
    --memory64-lowering \
    -O2 \
    --strip-debug \
    "$memory64_wasm" \
    -o "$wasm"

if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$wasm" > "$wasm.sha256"
else
    shasum -a 256 "$wasm" > "$wasm.sha256"
fi
