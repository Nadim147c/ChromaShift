#!/usr/bin/env zsh

if ! tty -s || [ ! -n "$TERM" ] || [ "$TERM" = dumb ] || (( ! $+commands[cshift] )); then
    return
fi

CSHIFT_EXECUTABLE=$(command -v cshift)

alias csudo="sudo $CSHIFT_EXECUTABLE --"

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
    docker-compose
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
            cshift -- ${commands[$0]} "$@"
        }
    fi
done

unset cmds cmd
