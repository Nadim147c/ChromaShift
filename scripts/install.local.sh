#!/bin/sh

DESTDIR="${PREFIX}/etc/colorize"

mkdir -p "$DESTDIR/rules"

install -m 644 rules/* "$DESTDIR/rules"
install -m 644 "colorize.toml" "$DESTDIR/config.toml"
install -m 644 "colorize.zsh" "$DESTDIR"
