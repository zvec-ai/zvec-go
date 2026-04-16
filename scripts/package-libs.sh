#!/usr/bin/env bash
# package-libs.sh — Build and package the zvec C-API library for vendor distribution.
#
# This script builds the zvec C-API library from the submodule and copies the
# resulting shared library and header to the lib/ directory for the current platform.
#
# Usage:
#   ./scripts/package-libs.sh              # Build and package for current platform
#   ./scripts/package-libs.sh --all        # Show instructions for all platforms
#   ./scripts/package-libs.sh --sync-header # Only sync the C-API header file

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# --- Helpers ----------------------------------------------------------------

print_header() { echo -e "\n=== $1 ==="; }
print_ok()     { echo "  ✓ $1"; }
print_err()    { echo "  ✗ $1" >&2; }

detect_platform() {
    local os arch
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    arch="$(uname -m)"

    case "$os" in
        linux)          os="linux" ;;
        darwin)         os="darwin" ;;
        mingw*|msys*)   os="windows" ;;
        *)              print_err "Unsupported OS: $os"; exit 1 ;;
    esac

    case "$arch" in
        x86_64|amd64)   arch="amd64" ;;
        aarch64|arm64)  arch="arm64" ;;
        *)              print_err "Unsupported architecture: $arch"; exit 1 ;;
    esac

    echo "${os}_${arch}"
}

detect_lib_ext() {
    case "$(uname -s)" in
        Darwin)         echo "dylib" ;;
        MINGW*|MSYS*)   echo "dll" ;;
        *)              echo "so" ;;
    esac
}

detect_nproc() {
    nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 2
}

# --- Sync header ------------------------------------------------------------

sync_header() {
    print_header "Syncing C-API header"
    local src="$REPO_ROOT/zvec/src/include/zvec/c_api.h"
    local dst="$REPO_ROOT/lib/include/zvec/c_api.h"

    if [[ ! -f "$src" ]]; then
        print_err "C-API header not found at $src"
        print_err "Make sure the zvec submodule is initialized: git submodule update --init --recursive"
        exit 1
    fi

    mkdir -p "$(dirname "$dst")"
    cp "$src" "$dst"
    print_ok "Header synced to lib/include/zvec/c_api.h"
}

# --- Build and package ------------------------------------------------------

build_and_package() {
    local platform lib_ext nproc_count
    platform="$(detect_platform)"
    lib_ext="$(detect_lib_ext)"
    nproc_count="$(detect_nproc)"

    print_header "Building zvec C-API for $platform"

    # Ensure submodule is initialized
    cd "$REPO_ROOT"
    git submodule update --init --recursive

    # Build
    cd "$REPO_ROOT/zvec"
    mkdir -p build && cd build

    if [[ "$platform" == windows_* ]]; then
        # Windows: use MSVC generator (no Ninja required)
        cmake .. \
            -DCMAKE_BUILD_TYPE=Release \
            -DBUILD_C_BINDINGS=ON 2>&1 | tail -5
        cmake --build . --config Release -j"$nproc_count" --target zvec_c_api 2>&1 | tail -5
    else
        # Unix: use Ninja generator
        cmake .. \
            -DCMAKE_BUILD_TYPE=Release \
            -DBUILD_C_BINDINGS=ON \
            -G Ninja 2>&1 | tail -5
        cmake --build . -j"$nproc_count" --target zvec_c_api 2>&1 | tail -5
    fi

    # Locate built library
    local built_lib=""
    local built_implib=""
    if [[ "$platform" == windows_* ]]; then
        # Windows: DLL may be in lib/Release/ or lib/
        for search_dir in "$REPO_ROOT/zvec/build/lib/Release" "$REPO_ROOT/zvec/build/lib" "$REPO_ROOT/zvec/build/bin/Release"; do
            if [[ -f "$search_dir/zvec_c_api.dll" ]]; then
                built_lib="$search_dir/zvec_c_api.dll"
                break
            fi
        done
        # Import library (.lib) for linking
        for search_dir in "$REPO_ROOT/zvec/build/lib/Release" "$REPO_ROOT/zvec/build/lib"; do
            if [[ -f "$search_dir/zvec_c_api.lib" ]]; then
                built_implib="$search_dir/zvec_c_api.lib"
                break
            fi
        done
    else
        built_lib="$REPO_ROOT/zvec/build/lib/libzvec_c_api.$lib_ext"
    fi

    if [[ -z "$built_lib" || ! -f "$built_lib" ]]; then
        print_err "Build failed: library not found"
        echo "  Searched for: libzvec_c_api.$lib_ext"
        echo "  Build directory contents:"
        find "$REPO_ROOT/zvec/build/lib" -type f 2>/dev/null || echo "  (lib directory not found)"
        find "$REPO_ROOT/zvec/build/bin" -type f 2>/dev/null || echo "  (bin directory not found)"
        exit 1
    fi
    print_ok "Library built: $(du -h "$built_lib" | cut -f1) ($built_lib)"

    # Copy to lib/ directory
    local target_dir="$REPO_ROOT/lib/$platform"
    mkdir -p "$target_dir"
    cp "$built_lib" "$target_dir/"
    print_ok "Library copied to lib/$platform/$(basename "$built_lib")"

    # Copy import library on Windows
    if [[ -n "$built_implib" && -f "$built_implib" ]]; then
        cp "$built_implib" "$target_dir/"
        print_ok "Import library copied to lib/$platform/$(basename "$built_implib")"
    fi

    # Sync header
    sync_header

    # Summary
    print_header "Package complete"
    echo "  Platform: $platform"
    echo "  Library:  lib/$platform/$(basename "$built_lib") ($(du -h "$target_dir/$(basename "$built_lib")" | cut -f1))"
    echo "  Header:   lib/include/zvec/c_api.h"
    echo ""
    echo "  To verify: go build -tags integration ./..."
    echo "  To commit: git add lib/ && git commit -m 'chore: update vendor libs for $platform'"
}

# --- Main -------------------------------------------------------------------

case "${1:-}" in
    --sync-header)
        sync_header
        ;;
    --all)
        echo "To package libraries for all platforms, run this script on each target:"
        echo ""
        echo "  macOS ARM64:  ./scripts/package-libs.sh"
        echo "  Linux x64:    ./scripts/package-libs.sh  (on linux-x64 machine)"
        echo "  Linux ARM64:  ./scripts/package-libs.sh  (on linux-arm64 machine)"
        echo ""
        echo "Or use the CI workflow to build all platforms automatically."
        ;;
    -h|--help)
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  (none)          Build and package for current platform"
        echo "  --all           Show instructions for all platforms"
        echo "  --sync-header   Only sync the C-API header file"
        echo "  -h, --help      Show this help message"
        ;;
    *)
        build_and_package
        ;;
esac
