package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"unicode"

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
)

var (
	Stat           = os.Stat
	DecodeTomlFile = toml.DecodeFile
)

func startRunWithoutColor(runCmd *exec.Cmd) {
	runCmd.Stderr = os.Stderr
	runCmd.Stdout = os.Stdout
	runCmd.Run()
	os.Exit(0)
}

var out = os.Stdout

func ReadIo(runCmd *exec.Cmd, cmdRules CommandRules, out io.Writer, ioPipe io.ReadCloser) {
	reader := bufio.NewReader(ioPipe)
	var buffer bytes.Buffer

	for {
		char, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintln(os.Stderr, "Error reading stdout:", err)
			break
		}

		if char == '\r' {
			line := buffer.String()
			coloredLine := ColorizeLine(line, cmdRules.Rules)
			if len(coloredLine) > 0 {
				fmt.Fprint(out, coloredLine+"\r")
			} else {
				fmt.Fprint(out, line+"\r")
			}
			buffer.Reset()
		} else {
			buffer.WriteByte(char)
		}

		if char == '\n' {
			line := strings.TrimRightFunc(buffer.String(), unicode.IsSpace)
			coloredLine := ColorizeLine(line, cmdRules.Rules)
			if len(coloredLine) > 0 {
				fmt.Fprint(out, coloredLine+"\n")
			} else {
				fmt.Fprint(out, line+"\n")
			}
			buffer.Reset()
		}
	}
}

type (
	StatFunc       func(string) (os.FileInfo, error)
	DecodeFileFunc func(string, interface{}) (toml.MetaData, error)
)

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

		runCmd.Stdin = os.Stdin

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigChan
			if err := runCmd.Process.Signal(sig); err != nil {
				fmt.Fprintln(os.Stderr, "Error sending signal to process:", err)
			}
		}()

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

		var ruleFileName string
		for name, values := range config {
			if cmdName == name {
				ruleFileName = values.File
				break
			}

			commandStr := strings.Join(args, " ")
			if matched, _ := regexp.Match(values.Regexp, []byte(commandStr)); matched {
				ruleFileName = values.File
				break
			}
		}

		if len(ruleFileName) <= 0 {
			if Verbose {
				fmt.Println("No config exists for current command")
			}
			startRunWithoutColor(runCmd)
		}

		cmdRules, err := LoadRules(ruleFileName)
		if Verbose && err != nil {
			fmt.Println("Failed to load rules for current command:", err)
		}

		if len(cmdRules.Rules) <= 0 {
			if Verbose {
				fmt.Println("No config exists for current command")
			}
			startRunWithoutColor(runCmd)
		}

		if len(cmdRules.SkipColor.Argument) > 0 {
			re, err := regexp.Compile(cmdRules.SkipColor.Argument)
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

		if len(cmdRules.SkipColor.Arguments) > 0 {
			re, err := regexp.Compile(cmdRules.SkipColor.Arguments)
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
			fmt.Printf("%d rules found.\n", len(cmdRules.Rules))
		}

		if cmdRules.Stderr {
			if Color != "always" && !termcolor.SupportsBasic(os.Stderr) {
				startRunWithoutColor(runCmd)
				os.Exit(0)
			}
			runCmd.Stdout = os.Stdout
			ioPipe, err := runCmd.StderrPipe()
			if err != nil {
				if Verbose {
					fmt.Fprintln(os.Stderr, "Error creating stdout pipe:", err)
				}
				os.Exit(1)
			}
			if err := runCmd.Start(); err != nil {
				if Verbose {
					fmt.Fprintln(os.Stderr, "Error starting command:", err)
				}
				os.Exit(1)
			}

			ReadIo(runCmd, cmdRules, os.Stderr, ioPipe)
		} else {
			if Color != "always" && !termcolor.SupportsBasic(os.Stdout) {
				startRunWithoutColor(runCmd)
				os.Exit(0)
			}
			runCmd.Stderr = os.Stderr
			ioPipe, err := runCmd.StdoutPipe()
			if err != nil {
				if Verbose {
					fmt.Fprintln(os.Stderr, "Error creating stdout pipe:", err)
				}
				os.Exit(1)
			}
			if err := runCmd.Start(); err != nil {
				if Verbose {
					fmt.Fprintln(os.Stderr, "Error starting command:", err)
				}
				os.Exit(1)
			}
			ReadIo(runCmd, cmdRules, os.Stdout, ioPipe)
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
