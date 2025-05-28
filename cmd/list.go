package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCommandsCmd)
}

var listCommandsCmd = &cobra.Command{
	Use:   "list",
	Short: "List available commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig()
		if err != nil {
			return err
		}

		for name, values := range config {
			fmt.Printf("%s[%s%s%s] %s%s%s%s = %s%s%s\n",
				Ansi.Bold, Ansi.Yellow, values.File, Ansi.Reset,
				Ansi.Bold, Ansi.Green, name, Ansi.Reset,
				Ansi.Cyan, values.Regexp, Ansi.Reset)
		}
		return nil
	},
}
