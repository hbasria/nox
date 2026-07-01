export GOWORK := "off"

build:
    go build -o nox .

# Cross-compile release binaries for macOS/Linux, Intel/ARM into dist/.
release:
    #!/usr/bin/env bash
    set -euo pipefail
    rm -rf dist
    mkdir -p dist
    targets=("darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64")
    for t in "${targets[@]}"; do
        os="${t%/*}"
        arch="${t#*/}"
        echo "building $os/$arch"
        GOOS="$os" GOARCH="$arch" go build -o "dist/nox-$os-$arch" .
    done
    (cd dist && shasum -a 256 nox-* > checksums.txt)
