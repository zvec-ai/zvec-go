#!/usr/bin/env bash
# sync-zvec.sh — Update the zvec submodule and verify C-API compatibility.
#
# Usage:
#   ./scripts/sync-zvec.sh                  # update to latest main
#   ./scripts/sync-zvec.sh v0.4.0           # update to a specific tag/commit
#   ./scripts/sync-zvec.sh --check-only     # only check for C-API changes (no update)
#   ./scripts/sync-zvec.sh --build          # update + rebuild C-API + run tests

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
C_API_HEADER="zvec/src/include/zvec/c_api.h"

# --- Helpers ----------------------------------------------------------------

print_header() { echo -e "\n=== $1 ==="; }
print_ok()     { echo "  ✓ $1"; }
print_warn()   { echo "  ⚠ $1"; }
print_err()    { echo "  ✗ $1" >&2; }

detect_nproc() {
    nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 2
}

# --- C-API change detection -------------------------------------------------

# Snapshot the C-API header before updating so we can diff later.
snapshot_c_api() {
    if [[ -f "$REPO_ROOT/$C_API_HEADER" ]]; then
        cp "$REPO_ROOT/$C_API_HEADER" /tmp/zvec_c_api_before.h
        return 0
    fi
    return 1
}

# Compare the C-API header before/after and report changes.
diff_c_api() {
    if [[ ! -f /tmp/zvec_c_api_before.h ]]; then
        print_warn "No previous C-API snapshot — skipping diff"
        return 0
    fi

    if ! diff -q /tmp/zvec_c_api_before.h "$REPO_ROOT/$C_API_HEADER" >/dev/null 2>&1; then
        print_header "C-API header changes detected"
        diff --color=auto -u /tmp/zvec_c_api_before.h "$REPO_ROOT/$C_API_HEADER" || true

        echo ""
        echo "  The following symbols may need Go binding updates:"
        # Show added/removed function declarations
        diff /tmp/zvec_c_api_before.h "$REPO_ROOT/$C_API_HEADER" \
            | grep -E '^[<>].*zvec_' \
            | sed 's/^< /  REMOVED: /; s/^> /  ADDED:   /' || true
        echo ""
        return 1
    else
        print_ok "C-API header unchanged"
        return 0
    fi
}

# --- Build C-API library ----------------------------------------------------

build_c_api() {
    print_header "Building C-API library"
    local nproc
    nproc=$(detect_nproc)

    cd "$REPO_ROOT/zvec"
    mkdir -p build && cd build
    cmake .. \
        -DCMAKE_BUILD_TYPE=Release \
        -DBUILD_C_BINDINGS=ON \
        -G Ninja 2>&1 | tail -5
    cmake --build . -j"$nproc" --target zvec_c_api 2>&1 | tail -5
    cd "$REPO_ROOT"

    if [[ -f zvec/build/lib/libzvec_c_api.a ]] || [[ -f zvec/build/lib/libzvec_c_api.dylib ]] || [[ -f zvec/build/lib/libzvec_c_api.so ]]; then
        print_ok "C-API library built successfully"
        ls -lh zvec/build/lib/libzvec_c_api.* 2>/dev/null
    else
        print_err "C-API library build failed"
        return 1
    fi
}

# --- Run Go tests -----------------------------------------------------------

run_tests() {
    print_header "Running Go tests"
    cd "$REPO_ROOT"
    CGO_ENABLED=1 go test -tags integration -count=1 -v ./... 2>&1
}

# --- Main -------------------------------------------------------------------

MODE="sync"
TARGET_REF="main"

for arg in "$@"; do
    case "$arg" in
        --check-only) MODE="check" ;;
        --build)      MODE="build" ;;
        -h|--help)
            echo "Usage: $0 [TARGET_REF] [--check-only|--build]"
            echo ""
            echo "  TARGET_REF     Git ref to sync to (default: main)"
            echo "  --check-only   Only check for C-API changes, don't update"
            echo "  --build        Update + rebuild C-API + run tests"
            echo ""
            exit 0
            ;;
        *)            TARGET_REF="$arg" ;;
    esac
done

cd "$REPO_ROOT"

if [[ "$MODE" == "check" ]]; then
    print_header "Checking for C-API changes (remote vs local)"
    git submodule update --init --recursive
    snapshot_c_api

    cd zvec && git fetch origin && cd "$REPO_ROOT"
    local_hash=$(git -C zvec rev-parse HEAD)
    remote_hash=$(git -C zvec rev-parse origin/main)

    if [[ "$local_hash" == "$remote_hash" ]]; then
        print_ok "Already up to date ($local_hash)"
        exit 0
    fi

    echo "  Local:  $local_hash"
    echo "  Remote: $remote_hash"
    echo ""
    echo "  Changes in C-API header:"
    git -C zvec diff "$local_hash".."$remote_hash" -- src/include/zvec/c_api.h || print_ok "No C-API changes"
    exit 0
fi

# --- Sync mode (default) or Build mode --------------------------------------

print_header "Syncing zvec submodule to '${TARGET_REF}'"

# Ensure submodule is initialized
git submodule update --init --recursive

# Snapshot C-API before update
snapshot_c_api
OLD_HASH=$(git -C zvec rev-parse --short HEAD 2>/dev/null || echo "none")

# Fetch latest from upstream
cd zvec
git fetch origin
git checkout "$TARGET_REF"

# If it's a branch (not a tag/commit), pull latest
if git symbolic-ref -q HEAD >/dev/null 2>&1; then
    git pull origin "$TARGET_REF"
fi
cd "$REPO_ROOT"

NEW_HASH=$(git -C zvec rev-parse --short HEAD)

print_header "zvec submodule updated"
echo "  Before: $OLD_HASH"
echo "  After:  $NEW_HASH"
git -C zvec log --oneline -1

# Check for C-API changes
C_API_CHANGED=0
diff_c_api || C_API_CHANGED=1

if [[ "$MODE" == "build" ]]; then
    build_c_api
    run_tests
fi

# --- Summary ----------------------------------------------------------------

print_header "Summary"
if [[ "$OLD_HASH" == "$NEW_HASH" ]]; then
    print_ok "Already up to date"
else
    echo "  Updated: $OLD_HASH → $NEW_HASH"
    if [[ "$C_API_CHANGED" -eq 1 ]]; then
        print_warn "C-API header changed — review Go bindings for compatibility"
    fi
fi

echo ""
echo "Next steps:"
if [[ "$MODE" != "build" ]]; then
    echo "  1. Build C-API:  make build-zvec"
    echo "  2. Run tests:    make test"
fi
echo "  3. Stage:        git add zvec"
echo "  4. Commit:       git commit -m 'chore(deps): update zvec submodule to $(git -C zvec describe --tags --always 2>/dev/null || echo "$NEW_HASH")'"
echo "  5. Push:         git push"
