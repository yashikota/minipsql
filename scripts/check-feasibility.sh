#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cache_dir=${MINIPSQL_CACHE_DIR:-"$repo_root/.cache"}
wasm="$repo_root/dist/upstream/pgrust-postgres.wasm"

if [ ! -f "$wasm" ]; then
    "$repo_root/scripts/build-wasm.sh"
fi

# This is deliberately a hard gate. pgrust is compiled as LP64 wasm64, but the
# canonical transpiler input must have its memory and table instructions lowered
# to their 32-bit forms before goccy wasm2go sees it.
(
    cd "$cache_dir/wasm2go"
    dump=$(go run ./cmd/wasm2go -dump -i "$wasm")
    printf '%s\n' "$dump"
    printf '%s\n' "$dump" | grep -q 'memory64=false'
)
