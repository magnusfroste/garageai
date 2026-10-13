#!/bin/sh
# Installs GarageAI Bridge, the `garageai` command.
#
#   curl -fsSL https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.sh | sh
#
# One binary, no runtime, nothing else installed. Downloaded with curl, so macOS puts no quarantine
# flag on it and Gatekeeper does not stop it. The download is checked against the release's
# SHA256SUMS. It goes to /usr/local/bin (with sudo when needed), and only without a terminal or
# sudo to ~/.local/bin; the script says which, and whether that is on your PATH.
#
#   GARAGEAI_BRIDGE_VERSION=bridge-v0.1.0   a specific release instead of the newest
#
# With `sh -s -- connect` it goes on to connect this machine (garageai connect: approve it in the
# browser), so one line installs and connects:
#
#   curl -fsSL https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.sh | sh -s -- connect
#
# Options after `connect` go to it, so one line also connects with the Bridge tunnel (no NetBird):
#
#   curl -fsSL https://raw.githubusercontent.com/magnusfroste/garageai/main/cli/install.sh | sh -s -- connect --tunnel bridge
set -eu

then=""
if [ "${1:-}" = "connect" ]; then then=connect; shift; fi   # the rest are connect's options (e.g. --tunnel bridge)

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
  # The repository has other releases too: take the highest stable bridge-v* version; a pre-release
  # (bridge-v0.3.0-rc1) only when no stable release exists. GARAGEAI_BRIDGE_VERSION picks one.
  VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases?per_page=50" 2>/dev/null \
    | sed -n 's/.*"tag_name": *"\(bridge-v[^"]*\)".*/\1/p' \
    | awk '{ v=$0; sub(/^bridge-v/, "", v); pre=0; if (v ~ /-/) { pre=1; sub(/-.*/, "", v) }
            n=split(v, p, "."); printf "%d%06d%06d%06d %s\n", 1-pre, p[1], p[2], p[3], $0 }' \
    | sort -r | head -n 1 | awk '{print $2}')
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

# /usr/local/bin is on everyone's PATH; ~/.local/bin is the last resort (on macOS it is not on
# the PATH, and garage-m4's operator could not find `garageai` after installing, 2026-10-13).
# connect needs sudo anyway, so asking for it here costs nothing: sudo caches the password.
sudo_install() {
  if sudo -n true 2>/dev/null; then   # no password needed (CI, some servers)
    sudo mkdir -p /usr/local/bin && sudo install -m 755 "$tmp/$asset" "$dest"
  elif (: </dev/tty) 2>/dev/null; then   # curl | sh: stdin is the script, so sudo asks on the terminal
    echo "Installing to /usr/local/bin (sudo may ask for your password)..."
    { sudo mkdir -p /usr/local/bin && sudo install -m 755 "$tmp/$asset" "$dest"; } </dev/tty
  else
    return 1
  fi
}
dest=/usr/local/bin/garageai
if [ -w /usr/local/bin ]; then
  mv "$tmp/$asset" "$dest"
elif [ "$(id -u)" != 0 ] && command -v sudo >/dev/null 2>&1 && sudo_install; then
  :
else
  mkdir -p "$HOME/.local/bin"
  dest="$HOME/.local/bin/garageai"
  mv "$tmp/$asset" "$dest"
fi

echo "Installed $("$dest" version) -> $dest (checksum verified)"
case ":$PATH:" in
  *":$(dirname "$dest"):"*) ;;
  *)
    echo ""
    echo "$(dirname "$dest") is not on your PATH. Add it:"
    echo "  echo 'export PATH=\"$(dirname "$dest"):\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
    ;;
esac
# A garage already connected: the heartbeat service runs its own root-only copy of Bridge, so
# update that one too (and move a service that ran the copy in the home directory, before v0.2.4).
if [ -f /Library/LaunchDaemons/eu.garageai.bridge.plist ] || [ -f /etc/systemd/system/garageai-bridge.service ]; then
  echo ""
  echo "Updating the heartbeat service (it runs as root, so sudo may ask for your password)..."
  if [ "$(id -u)" = 0 ]; then "$dest" install-service; else sudo "$dest" install-service; fi ||
    echo "Could not update the heartbeat service. Run: sudo $dest install-service" >&2
fi
echo ""
if [ "$then" = connect ]; then
  # stdin is the script itself (curl | sh): connect asks questions, so give it the terminal.
  if (: </dev/tty) 2>/dev/null; then
    exec "$dest" connect "$@" </dev/tty
  fi
  echo "No terminal to ask questions in: run  garageai connect" >&2
  exit 1
fi
echo "Next:"
echo "  garageai connect    connect this machine to GarageAI (you approve it in the browser)"
echo "  garageai doctor     what runs on this machine, what is wrong, and how to fix it"
