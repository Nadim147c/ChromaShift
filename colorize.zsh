#!/usr/bin/env zsh

if ! tty -s || [ ! -n "$TERM" ] || [ "$TERM" = dumb ] || (( ! $+commands[colorize] )); then
    return
fi

cmds=(
    ping stat traceroute df env
    cp mv rm ps lsblk mount
)

for cmd in $cmds ; do
    if (( $+commands[$cmd] )) ; then
        $cmd() {
            colorize -- ${commands[$0]} "$@"
        }
    fi
done

unset cmds cmd
