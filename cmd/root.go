package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
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

func startRunWithoutColor(runCmd *exec.Cmd) {
	runCmd.Stderr = os.Stderr
	runCmd.Stdout = os.Stdout
	runCmd.Run()
	os.Exit(0)
}

var rootCmd = &cobra.Command{
	Use:     "colorize [OPTOINS] -- COMMAND [OPTIONS/ARGUMENTS]",
	Version: Version,
	Example: "colorize -- stat go.mod",
	Short:   "A colorizer for your favorite commands",
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

		if err != nil && Verbose {
			fmt.Fprintln(os.Stderr, "Failed to load config:", err)
		}

		cmdBaseName := filepath.Base(cmdName)
		ruleFileName := config[cmdBaseName].File
		if len(ruleFileName) <= 0 {
			for name, values := range config {
				if cmdName == name || cmdBaseName == name {
					ruleFileName = values.File
					break
				}

				commandStr := strings.Join(args, " ")
				if matched, _ := regexp.Match(values.Regexp, []byte(commandStr)); matched {
					ruleFileName = values.File
					break
				}
			}
		}

		if len(ruleFileName) <= 0 {
			if Verbose {
				fmt.Println("No config exists for current command")
			}
			startRunWithoutColor(runCmd)
		}

		CmdRules, err = LoadRules(ruleFileName)
		if Verbose && err != nil {
			fmt.Println("Failed to load rules for current command:", err)
		}

		if len(CmdRules.Rules) <= 0 {
			if Verbose {
				fmt.Println("No config exists for current command")
			}
			startRunWithoutColor(runCmd)
		}

		if len(CmdRules.SkipColor.Argument) > 0 {
			re, err := regexp.Compile(CmdRules.SkipColor.Argument)
			if err != nil {
				if Verbose {
					fmt.Println("failed to compile ignore argument", err)
				}
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
				if Verbose {
					fmt.Println("failed to compile ignore arguments", err)
				}
				startRunWithoutColor(runCmd)
			}

			if re.Match([]byte(strings.Join(cmdArgs, " "))) {
				startRunWithoutColor(runCmd)
				os.Exit(0)
			}
		}

		if Verbose {
			fmt.Printf("%d rules found.\n", len(CmdRules.Rules))
		}

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
			if Verbose {
				fmt.Fprintln(os.Stderr, "Error waiting for command:", err)
			}
			os.Exit(1)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetErrPrefix("Colorize Error:")
	rootCmd.Flags().StringVar(&ConfigFile, "config", "", "specify path to the config file")
	rootCmd.Flags().StringVar(&RulesDirectory, "rules-dir", "", "specify path to the rules directory")
	rootCmd.Flags().StringVar(&Color, "color", "auto", "whether use color or not (never, auto, always)")
	rootCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}
