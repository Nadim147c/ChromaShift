package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/efekarakus/termcolor"
	"github.com/spf13/cobra"
)

var Version = "dev"

var (
	Color          string
	ConfigFile     string
	RulesDirectory string
	Verbose        bool
	UseColor       bool
	CmdRules       CommandRules

	Stat           = os.Stat
	DecodeToml     = toml.Decode
	DecodeTomlFile = toml.DecodeFile
)

func Debug(a ...any) {
	if Verbose {
		fmt.Fprintln(os.Stderr, a...)
	}
}

func startRunWithoutColor(runCmd *exec.Cmd) {
	runCmd.Stderr = os.Stderr
	runCmd.Stdout = os.Stdout
	runCmd.Run()
	os.Exit(0)
}

var rootCmd = &cobra.Command{
	Use:     "cshift",
	Version: Version,
	Short:   "A output colorizer for your favorite commands",
	Run: func(cmd *cobra.Command, args []string) {
		UseColor = true

		if len(args) < 1 {
			cmd.Help()
			os.Exit(0)
		}

		cmdName := args[0]
		cmdArgs := args[1:]

		runCmd := exec.Command(cmdName, cmdArgs...)

		switch Color {
		case "never":
			UseColor = false
		case "always":
			UseColor = true
		default:
			UseColor = termcolor.SupportsBasic(os.Stdout) || termcolor.SupportsBasic(os.Stderr)
		}

		if !UseColor {
			startRunWithoutColor(runCmd)
		}

		config, err := LoadConfig()
		if err != nil {
			Debug("Failed to load config:", err)
		}

		ruleFileName, err := GetRuleFileName(config, args)
		if err != nil {
			Debug("No config exists for current command")
			startRunWithoutColor(runCmd)
		} else {
			Debug("Rules file name", ruleFileName)
		}

		CmdRules, err = LoadRules(ruleFileName)
		if err != nil {
			Debug("Failed to load rules for current command:", err)
		}

		if len(CmdRules.Rules) <= 0 {
			Debug("No config exists for current command")
			startRunWithoutColor(runCmd)
		}

		if len(CmdRules.SkipColor.Argument) > 0 {
			re, err := regexp.Compile(CmdRules.SkipColor.Argument)
			if err != nil {
				Debug("failed to compile ignore argument", err)
				startRunWithoutColor(runCmd)
			}
			for _, arg := range cmdArgs {
				if re.Match([]byte(arg)) {
					startRunWithoutColor(runCmd)
					os.Exit(0)
				}
			}
		}

		if len(CmdRules.SkipColor.Arguments) > 0 {
			re, err := regexp.Compile(CmdRules.SkipColor.Arguments)
			if err != nil {
				Debug("failed to compile ignore arguments", err)
				startRunWithoutColor(runCmd)
			}

			if re.Match([]byte(strings.Join(cmdArgs, " "))) {
				startRunWithoutColor(runCmd)
				os.Exit(0)
			}
		}

		Debug("rules found:", len(CmdRules.Rules))

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigChan
			if err := runCmd.Process.Signal(sig); err != nil {
				fmt.Fprintln(os.Stderr, "Error sending signal to process:", err)
			}
		}()

		if CmdRules.PTY {
			var out *os.File
			if CmdRules.Stderr {
				out = os.Stderr
			} else {
				out = os.Stdout
			}
			outputReader := Output{Command: runCmd, Out: out}
			outputReader.StartWithPTY(CmdRules.Stderr)
		} else {
			runCmd.Stdin = os.Stdin
			var out *os.File
			if CmdRules.Stderr {
				out = os.Stderr
			} else {
				out = os.Stdout
			}
			outputReader := Output{Command: runCmd, Out: out}
			outputReader.Start(CmdRules.Stderr)
		}

		if err := runCmd.Wait(); err != nil {
			Debug("Error waiting for command:", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	cobra.AddTemplateFunc("Heading", func(s interface{}) string {
		if color, _ := rootCmd.Flags().GetString("color"); termcolor.SupportsBasic(os.Stdout) || color == "always" {
			return Ansi.Yellow + Ansi.Bold + fmt.Sprint(s) + Ansi.Reset
		} else {
			return fmt.Sprint(s)
		}
	})
	cobra.AddTemplateFunc("CommandName", func(s interface{}) string {
		if color, _ := rootCmd.Flags().GetString("color"); termcolor.SupportsBasic(os.Stdout) || color == "always" {
			return Ansi.Green + fmt.Sprint(s) + Ansi.Reset
		} else {
			return fmt.Sprint(s)
		}
	})
	cobra.AddTemplateFunc("Option", func(s interface{}) string {
		if color, _ := rootCmd.Flags().GetString("color"); termcolor.SupportsBasic(os.Stdout) || color == "always" {
			return Ansi.Red + fmt.Sprint(s) + Ansi.Reset
		} else {
			return fmt.Sprint(s)
		}
	})

	rootCmd.SetHelpTemplate("{{ Heading .Short }}\n\n{{.UsageString}}")

	usage := `{{ Heading "Usage" }}:
  {{ CommandName "cshift" }} [{{ Option "CHROMASHIFT_OPTIONS" }}] {{ Option "--" }} <{{ CommandName "COMMAND" }}> [{{ Option "OPTIONS" }}]

{{ Heading "Examples" }}:
  {{ CommandName "cshift" }} {{ Option "--" }} {{ CommandName "stat" }} {{ Option "go.mod" }}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

{{ Heading "Available Commands" }}:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{ CommandName (rpad .Name .NamePadding) }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{ CommandName (rpad .Name .NamePadding) }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

{{ Heading "Additional Commands" }}:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{ CommandName (rpad .Name .NamePadding) }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{ Heading "Flags"}}:
{{ Option (.LocalFlags.FlagUsages | trimTrailingWhitespaces) }}{{end}}{{if .HasAvailableSubCommands}}

{{ Heading "Use" }}: "{{ CommandName .CommandPath }} <{{ CommandName "COMMAND" }}> {{ Option "--help" }}" for more information about a command.{{end}}
`
	rootCmd.SetUsageTemplate(usage)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetErrPrefix("ChromaShift Error:")
	rootCmd.Flags().StringVar(&ConfigFile, "config", "", "specify path to the config file")
	rootCmd.Flags().StringVar(&RulesDirectory, "rules-dir", "", "specify path to the rules directory")
	rootCmd.Flags().StringVar(&Color, "color", "auto", "whether use color or not (never, auto, always)")
	rootCmd.Flags().BoolVarP(&Verbose, "debug", "d", false, "verbose output")
}
