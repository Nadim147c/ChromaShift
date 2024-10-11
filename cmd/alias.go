package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	aliasCmd.AddCommand(aliasZshCmd)
	aliasCmd.AddCommand(aliasBashCmd)
	rootCmd.AddCommand(aliasCmd)
}

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Generate the aliases script for the specified shell",
}

var aliasZshCmd = &cobra.Command{
	Use: "zsh",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig()
		if err != nil {
			return err
		}
		script := `#!/bin/zsh
if ! tty -s || [ ! -n "$TERM" ] || [ "$TERM" = dumb ] || (( ! $+commands[cshift] )); then
    return
fi

alias csudo="sudo $commands[cshift] --"
`
		zshFunction := `
if (( $+commands[%s] )) ; then
    function %s {
        cshift -- %s "$@"
    }
fi
`

		fmt.Println(script)
		for cmd := range config {
			fmt.Printf(zshFunction, cmd, cmd, cmd)
		}

		return nil
	},
}

var aliasBashCmd = &cobra.Command{
	Use: "bash",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig()
		if err != nil {
			return err
		}
		script := `#!/bin/bash

if ! tty -s || [ -z "$TERM" ] || [ "$TERM" = "dumb" ] || ! command -v cshift >/dev/null; then
    exit 1
fi

alias csudo="sudo $(command -v cshift) --"
`
		bashFunction := `if command -v "%s" >/dev/null ; then
    function %s {
        cshift -- "%s" $@
    }
fi

`
		fmt.Println(script)
		for cmd := range config {
			fmt.Printf(bashFunction, cmd, cmd, cmd)
		}

		return nil
	},
}
