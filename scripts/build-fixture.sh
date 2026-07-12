#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cache_dir=${MINIPSQL_CACHE_DIR:-"$repo_root/.cache"}
pgrust_dir="$cache_dir/pgrust"
target_dir=${CARGO_TARGET_DIR:-"$cache_dir/cargo-target-native"}
data_dir="$cache_dir/initdb-native"
share_dir="$pgrust_dir/vendor/postgres-18.3/share"

if [ ! -d "$pgrust_dir/.git" ]; then
    "$repo_root/scripts/fetch-upstream.sh"
fi

PGRUST_PGSHAREDIR="$share_dir" CARGO_TARGET_DIR="$target_dir" \
cargo build --release --locked --manifest-path "$pgrust_dir/Cargo.toml" --bin postgres

rm -rf "$data_dir"
"$target_dir/release/postgres" --initdb \
    -D "$data_dir" \
    -L "$share_dir" \
    --no-locale \
    --encoding UTF8 \
    -U postgres

# The wasm artifact intentionally bakes this stable virtual installation path.
# It is resolved by pgvfs inside the same in-memory tree at runtime.
mkdir -p "$data_dir/usr/local/pgsql"
cp -R "$share_dir" "$data_dir/usr/local/pgsql/share"

mkdir -p "$repo_root/internal/fixture" "$cache_dir/go-build" "$cache_dir/tmp"
GOCACHE="$cache_dir/go-build" TMPDIR="$cache_dir/tmp" \
go run "$repo_root/tools/fixturepack" \
    -source "$data_dir" \
    -output "$repo_root/internal/fixture/pgdata.zip"
