#!/usr/bin/env sh
# Installs the latest nox release binary for this machine's OS/architecture.
#
#   curl -fsSL https://raw.githubusercontent.com/hbasria/nox/main/install.sh | sh
set -eu

REPO="hbasria/nox"
INSTALL_DIR="${NOX_INSTALL_DIR:-/usr/local/bin}"

os=$(uname -s)
case "$os" in
    Darwin) os="darwin" ;;
    Linux) os="linux" ;;
    *)
        echo "nox: unsupported OS: $os" >&2
        exit 1
        ;;
esac

arch=$(uname -m)
case "$arch" in
    x86_64 | amd64) arch="amd64" ;;
    arm64 | aarch64) arch="arm64" ;;
    *)
        echo "nox: unsupported architecture: $arch" >&2
        exit 1
        ;;
esac

asset="nox-${os}-${arch}"
url="https://github.com/${REPO}/releases/latest/download/${asset}"

echo "nox: downloading ${asset} from the latest ${REPO} release..."
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
curl -fsSL "$url" -o "$tmp"
chmod +x "$tmp"

mkdir -p "$INSTALL_DIR"
if [ -w "$INSTALL_DIR" ]; then
    mv "$tmp" "$INSTALL_DIR/nox"
else
    echo "nox: sudo is required to write to $INSTALL_DIR"
    sudo mv "$tmp" "$INSTALL_DIR/nox"
fi

echo "nox: installed to $INSTALL_DIR/nox"
case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *) echo "nox: note - $INSTALL_DIR isn't on your PATH" ;;
esac
