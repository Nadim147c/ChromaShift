#!/bin/sh

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

REQUIRED_COMMANDS="grep cut curl mktemp realpath sudo install"
for cmd in $REQUIRED_COMMANDS; do
    if ! command_exists "$cmd"; then
        echo "Error: Required command '$cmd' is not installed or not in PATH."
        exit 1
    fi
done

TEMP=$(mktemp -d)
cd "$TEMP" || exit

PLATFORM=$(uname | tr '[:upper:]' '[:lower:]')
if [ "$PLATFORM" = "linux" ]; then
    PLATFORM="linux"
elif [ "$PLATFORM" = "darwin" ]; then
    PLATFORM="darwin"
else
    echo "Unsupported platform: $PLATFORM"
    exit 1
fi

ARCHITECTURE=$(uname -m)
if [ "$ARCHITECTURE" = "x86_64" ]; then
    ARCHITECTURE="amd64"
elif [ "$ARCHITECTURE" = "arm64" ] || [ "$ARCHITECTURE" = "aarch64" ]; then
    ARCHITECTURE="arm64"
else
    echo "Unsupported architecture: $ARCHITECTURE"
    exit 1
fi

REPO="Nadim147c/chromashift"

RELEASE_URL="https://api.github.com/repos/$REPO/releases/latest"

ASSET_URL=$(curl -s $RELEASE_URL | grep "browser_download_url" | grep "$PLATFORM" | grep "$ARCHITECTURE" | cut -d '"' -f 4)

if [ -z "$ASSET_URL" ]; then
    echo "No matching release found for $ARCHITECTURE on $PLATFORM"
    exit 1
fi

DOWNLOAD_PATH="release-${PLATFORM}-${ARCHITECTURE}.tar.gz"
printf '%s\n' "Downloading from $ASSET_URL" " to $(realpath "$DOWNLOAD_PATH")..."

curl -L -o "$DOWNLOAD_PATH" "$ASSET_URL"

echo "Download completed."

echo "Extracting download archive..."
tar -xvf "$DOWNLOAD_PATH"

PREFIX="${PREFIX:-/usr/local}"
DEST_DIR="${PREFIX}/etc/chromashift"
BIN_DIR="${BIN_DIR:-$PREFIX/bin}"

maybe_sudo() {
    if [ -w "$PREFIX" ]; then
        "$@"
    else
        printf '%s\n\n' "sudo is require to run: $*"
        sudo "$@"
    fi
}

maybe_sudo mkdir -p "$DEST_DIR/rules"

maybe_sudo install -m 755 ./bin/* "$BIN_DIR"
maybe_sudo install -m 644 "scripts/alias.zsh" "$DEST_DIR"
