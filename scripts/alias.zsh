#!/usr/bin/env zsh

if ! tty -s || [ ! -n "$TERM" ] || [ "$TERM" = dumb ] || (( ! $+commands[colorize] )); then
    return
fi

COLORIZE_EXECUTABLE=$(command -v colorize)

alias csudo="sudo COLORIZE_CONFIG=\$COLORIZE_CONFIG COLORIZE_RULES=\$COLORIZE_RULES $COLORIZE_EXECUTABLE --"

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
)

for cmd in $cmds ; do
    if (( $+commands[$cmd] )) ; then
        unalias $cmd 2>/dev/null
        unfunction $cmd 2>/dev/null
        $cmd() {
            colorize -- ${commands[$0]} "$@"
        }
    fi
done

unset cmds cmd
