#!/bin/sh

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

maybe_sudo mkdir -p "$DEST_DIR" "$BIN_DIR"

maybe_sudo install -m 755 ./bin/* "$BIN_DIR"
maybe_sudo install -m 644 scripts/alias.* "$DEST_DIR"
