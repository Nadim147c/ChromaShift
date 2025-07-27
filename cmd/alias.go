package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	aliasCmd.AddCommand(aliasZshCmd)
	aliasCmd.AddCommand(aliasBashCmd)
	aliasCmd.AddCommand(aliasFishCmd)
	aliasCmd.AddCommand(aliasNuCmd)
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

var aliasFishCmd = &cobra.Command{
	Use: "fish",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig()
		if err != nil {
			return err
		}
		cmds := make([]string, 0, len(config))
		for cmd := range config {
			cmds = append(cmds, cmd)
		}
		fmt.Printf(`#!/bin/fish

set cshift_cmd_list %s

for executable in $cshift_cmd_list
    if type -q $executable
        function $executable --inherit-variable executable --wraps=$executable
            if isatty 1
                cshift -- $executable $argv
            else
                eval command $executable $argv
            end
        end
    end
end
`, strings.Join(cmds, " "))

		return nil
	},
}

var aliasNuCmd = &cobra.Command{
	Use: "nu",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig()
		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stderr, "The nushell alias script is experimental!!!")

		banned := map[string]bool{
			"ps": true, "last": true, "find": true,
			"cp": true, "mv": true, "rm": true,
		}

		script := `#!/bin/nu

if ($env.TERM == "dumb") and (which cshift | is-not-empty) {
    exit 1
}
`
		fmt.Println(script)
		for cmd := range config {
			if _, ok := banned[cmd]; ok {
				continue
			}
			fmt.Printf("def --wrapped %s [...p] { cshift -- %s ...$p }\n", cmd, cmd)
		}

		return nil
	},
}
