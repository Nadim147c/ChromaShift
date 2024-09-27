#!/bin/sh

PREFIX="${PREFIX:-/usr/local}"
DEST_DIR="${PREFIX}/etc/colorize"
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

maybe_sudo install -m 755 ./colorize "$BIN_DIR"
maybe_sudo install -m 644 "config.toml" "$DEST_DIR/config.toml"
maybe_sudo install -m 644 "scripts/alias.zsh" "$DEST_DIR"
maybe_sudo install -m 644 rules/* "$DEST_DIR/rules"
