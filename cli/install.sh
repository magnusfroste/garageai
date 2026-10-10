#!/bin/sh
# Installs GarageAI Bridge, the `garageai` command.
#
#   curl -fsSL https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.sh | sh
#
# One binary, no runtime, nothing else installed. Downloaded with curl, so macOS puts no quarantine
# flag on it and Gatekeeper does not stop it. The download is checked against the release's
# SHA256SUMS. It goes to /usr/local/bin when that is writable, otherwise to ~/.local/bin, and the
# script says which, and whether that is on your PATH.
#
#   GARAGEAI_BRIDGE_VERSION=bridge-v0.1.0   a specific release instead of the newest
set -eu

REPO="${GARAGEAI_BRIDGE_REPO:-magnusfroste/garageai}"
VERSION="${GARAGEAI_BRIDGE_VERSION:-latest}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "garageai: no build for $arch" >&2; exit 1 ;;
esac
case "$os" in
  darwin|linux) ;;
  *) echo "garageai: no build for $os (on Windows, download garageai-windows-amd64.exe from the release page)" >&2; exit 1 ;;
esac

if [ -n "${GARAGEAI_BRIDGE_DOWNLOAD_BASE:-}" ]; then VERSION=local; fi   # a local build server, for tests
if [ "$VERSION" = "latest" ]; then
  # The repository has other releases too: take the newest one whose tag starts with bridge-v.
  VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases?per_page=30" 2>/dev/null \
    | sed -n 's/.*"tag_name": *"\(bridge-v[^"]*\)".*/\1/p' | head -n 1)
  [ -n "$VERSION" ] || { echo "garageai: could not find a Bridge release in $REPO (is github.com reachable?)" >&2; exit 1; }
fi
base="https://github.com/$REPO/releases/download/$VERSION"
[ -n "${GARAGEAI_BRIDGE_DOWNLOAD_BASE:-}" ] && base="$GARAGEAI_BRIDGE_DOWNLOAD_BASE"
asset="garageai-$os-$arch"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
echo "Downloading GarageAI Bridge $VERSION for $os/$arch..."
curl -fsSL "$base/$asset" -o "$tmp/$asset" || { echo "garageai: download failed: $base/$asset" >&2; exit 1; }
curl -fsSL "$base/SHA256SUMS" -o "$tmp/SHA256SUMS" || { echo "garageai: could not get SHA256SUMS for $VERSION" >&2; exit 1; }
want=$(awk -v f="$asset" '$2 == f || $2 == "*"f {print $1}' "$tmp/SHA256SUMS")
if command -v sha256sum >/dev/null 2>&1; then have=$(sha256sum "$tmp/$asset" | awk '{print $1}')
else have=$(shasum -a 256 "$tmp/$asset" | awk '{print $1}'); fi
[ -n "$want" ] && [ "$want" = "$have" ] || { echo "garageai: checksum mismatch for $asset, not installing" >&2; exit 1; }
chmod +x "$tmp/$asset"

if [ -w /usr/local/bin ]; then
  dest=/usr/local/bin/garageai
else
  mkdir -p "$HOME/.local/bin"
  dest="$HOME/.local/bin/garageai"
fi
mv "$tmp/$asset" "$dest"

echo "Installed $("$dest" version) -> $dest (checksum verified)"
case ":$PATH:" in
  *":$(dirname "$dest"):"*) ;;
  *)
    echo ""
    echo "$(dirname "$dest") is not on your PATH. Add it:"
    echo "  echo 'export PATH=\"$(dirname "$dest"):\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
    ;;
esac
echo ""
echo "Next:"
echo "  garageai doctor     what runs on this machine, what is wrong, and how to fix it"
