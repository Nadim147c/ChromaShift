#!/usr/bin/env zsh

if ! tty -s || [ ! -n "$TERM" ] || [ "$TERM" = dumb ] || (( ! $+commands[colorize] )); then
    return
fi

COLORIZE_EXECUTABLE=$(command -v colorize)

alias csudo="sudo $COLORIZE_EXECUTABLE --"

cmds=(
    ping
    stat
    traceroute
    df
    du
    env
    cp
    mv
    rm
    ps
    lsblk
    mount
    lsmod
    free
    docker
    yt-dlp
    go
    id
    strace
    find
    curl
    wget
)

for cmd in $cmds ; do
    if (( $+commands[$cmd] )) ; then
        $cmd() {
            colorize -- ${commands[$0]} "$@"
        }
    fi
done

unset cmds cmd
