#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
. "$repo_root/tools/versions.env"

cache_dir=${MINIPSQL_CACHE_DIR:-"$repo_root/.cache"}

fetch_repo() {
    name=$1
    repository=$2
    revision=$3
    directory="$cache_dir/$name"

    if [ ! -d "$directory/.git" ]; then
        mkdir -p "$cache_dir"
        git clone --filter=blob:none --no-checkout "$repository" "$directory"
    fi

    git -C "$directory" fetch --depth 1 origin "$revision"
    git -C "$directory" checkout --detach --force "$revision"
}

fetch_repo pgrust "$PGRUST_REPOSITORY" "$PGRUST_REVISION"
fetch_repo wasm2go "$WASM2GO_REPOSITORY" "$WASM2GO_REVISION"

for patch in "$repo_root"/patches/wasm2go/*.patch; do
    [ -f "$patch" ] || continue
    git -C "$cache_dir/wasm2go" apply "$patch"
done
