#!/usr/bin/env bash
# Build NoGrassWrapper for all platforms
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
BUILD="$DIR/build"
SRC="$DIR/src"

echo "==> NoGrassWrapper Builder"
echo "    Output dir: ${BUILD}"
echo ""

# Generate icon PNG from SVG for Windows fallback
if command -v rsvg-convert &>/dev/null; then
    rsvg-convert -w 64 -h 64 "$DIR/assets/Icon.svg" > "$SRC/assets/Icon.png"
    echo "  Generated src/assets/Icon.png"
elif command -v convert &>/dev/null; then
    convert -background none -size 64x64 "$DIR/assets/Icon.svg" "$SRC/assets/Icon.png"
    echo "  Generated src/assets/Icon.png (via ImageMagick)"
else
    echo "  [WARN] No SVG converter found — using existing src/assets/Icon.png"
fi

build() {
    local os="$1" arch="$2" suffix="$3" cgo="${4:-0}"
    local name="nograsswrapper${suffix}"
    local out="${BUILD}/${name}"
    local ver
    ver="$(git describe --tags 2>/dev/null || echo dev)"
    local ldflags="-s -w -X main.Version=${ver}"
    if [[ "$os" == "windows" ]]; then
        ldflags="-H windowsgui -s -w -X main.Version=${ver}"
    fi
    echo "  Building ${os}/${arch}  (version: ${ver}, cgo: ${cgo})..."
    GOOS="$os" GOARCH="$arch" CGO_ENABLED="$cgo" CC="${CC:-gcc}" go build -ldflags="${ldflags}" -o "${out}" ./src 2>&1 | sed 's/^/    /'
    local size
    size=$(du -h "${out}" 2>/dev/null | cut -f1)
    echo "    -> ${out}  (${size})"
}

# Ensure build dir exists
mkdir -p "$BUILD"

# Linux — requires CGo for systray GTK
build linux amd64 "" 1
if command -v aarch64-linux-gnu-gcc &>/dev/null; then
    # Cross-compile arm64 with CGo (requires aarch64-linux-gnu-gcc)
    CC=aarch64-linux-gnu-gcc build linux arm64 "-arm64" 1
else
    echo "  [SKIP] linux/arm64 — install aarch64-linux-gnu-gcc for CGo cross-build"
fi

# Windows — pure Go, no CGo needed
build windows amd64 ".exe"
build windows arm64 "-arm64.exe"

# macOS — requires CGo for systray (native ObjC)
if [[ "$(uname)" == "Darwin" ]]; then
    build darwin amd64 "-mac" 1
    build darwin arm64 "-mac-arm64" 1
else
    echo "  [SKIP] darwin/amd64 — build natively on macOS"
    echo "  [SKIP] darwin/arm64 — build natively on macOS"
fi

echo ""
echo "==> Done! Binaries in build/:"
ls -lh "$BUILD"
