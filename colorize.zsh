#!/usr/bin/env zsh

if ! tty -s || [ ! -n "$TERM" ] || [ "$TERM" = dumb ] || (( ! $+commands[colorize] )); then
    return
fi

cmds=(
    ping stat traceroute df env
    cp mv rm ps lsblk mount lsmod
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
