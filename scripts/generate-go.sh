#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cache_dir=${MINIPSQL_CACHE_DIR:-"$repo_root/.cache"}
wasm="$repo_root/dist/upstream/pgrust-postgres.wasm"
wasm2go_dir="$cache_dir/wasm2go"
output_dir="$repo_root/internal/generated/pgrust"

if [ ! -f "$wasm" ]; then
    "$repo_root/scripts/build-wasm.sh"
fi
if [ ! -d "$wasm2go_dir/.git" ]; then
    "$repo_root/scripts/fetch-upstream.sh"
fi

mkdir -p "$cache_dir/go-build" "$cache_dir/tmp" "$output_dir"

# pgrust has more than 43k reachable functions. gcasm's dual-architecture
# assembly capture is useful for smaller modules, but its listings dominate
# generation memory here. wasm2go's pure fallback keeps the same semantics and
# produces portable source for every GOOS/GOARCH without cgo.
(
    cd "$wasm2go_dir"
    WASM2GO_PURE_GO=1 \
    GOCACHE="$cache_dir/go-build" \
    TMPDIR="$cache_dir/tmp" \
    go run ./cmd/wasm2go \
        -i "$wasm" \
        -pkg pgrust \
        -import github.com/yashikota/minipsql/internal/generated/pgrust \
        -out-dir "$output_dir" \
        -entry-exports main
)

GOCACHE="$cache_dir/go-build" TMPDIR="$cache_dir/tmp" \
go run "$repo_root/tools/envstubgen" \
    -in "$output_dir/base/base.go" \
    -out "$output_dir/host/env.go" \
    -package host \
    -base-import github.com/yashikota/minipsql/internal/generated/pgrust/base
