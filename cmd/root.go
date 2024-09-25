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

	"github.com/efekarakus/termcolor"
	"github.com/spf13/cobra"
)

var Version = "dev"

var (
	color    string
	cfgFile  string
	rulesDir string
	verbose  bool
	useColor bool
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
		useColor = true

		if len(args) < 1 {
			cmd.Help()
			os.Exit(0)
		}

		cmdName := args[0]
		cmdArgs := args[1:]

		runCmd := exec.Command(cmdName, cmdArgs...)

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigChan
			if err := runCmd.Process.Signal(sig); err != nil {
				fmt.Fprintln(os.Stderr, "Error sending signal to process:", err)
			}
		}()

		switch color {
		case "never":
			useColor = false
		case "always":
			useColor = true
		default:
			useColor = termcolor.SupportsBasic(os.Stdout)
		}

		if !useColor {
			startRunWithoutColor(runCmd)
		}

		config, err := loadConfig()

		if err != nil && verbose {
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
			if verbose {
				fmt.Println("No config exists for current command")
			}
			startRunWithoutColor(runCmd)
		}

		cmdRules, err := loadRules(ruleFileName)
		if verbose && err != nil {
			fmt.Println("Failed to load rules for current command:", err)
		}

		if len(cmdRules.Rules) <= 0 {
			if verbose {
				fmt.Println("No config exists for current command")
			}
			startRunWithoutColor(runCmd)
		}

		if len(cmdRules.SkipColor.Argument) > 0 {
			re, err := regexp.Compile(cmdRules.SkipColor.Argument)
			if err != nil {
				if verbose {
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
				if verbose {
					fmt.Println("failed to compile ignore arguments", err)
				}
				startRunWithoutColor(runCmd)
			}

			if re.Match([]byte(strings.Join(cmdArgs, " "))) {
				startRunWithoutColor(runCmd)
				os.Exit(0)
			}
		}

		if verbose {
			fmt.Printf("%d rules found.\n", len(cmdRules.Rules))
		}

		runCmd.Stderr = os.Stderr

		stdout, err := runCmd.StdoutPipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating stdout pipe:", err)
			os.Exit(1)
		}
		if err := runCmd.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "Error starting command:", err)
			os.Exit(1)
		}

		reader := bufio.NewReader(stdout)
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
				coloredLine := colorizeLine(line, cmdRules.Rules)
				if len(coloredLine) > 0 {
					fmt.Print(coloredLine + "\r")
				} else {
					fmt.Print(line + "\r")
				}
				buffer.Reset()
			} else {
				buffer.WriteByte(char)
			}

			if char == '\n' {
				line := strings.TrimRightFunc(buffer.String(), unicode.IsSpace)
				coloredLine := colorizeLine(line, cmdRules.Rules)
				if len(coloredLine) > 0 {
					fmt.Print(coloredLine + "\n")
				} else {
					fmt.Print(line + "\n")
				}
				buffer.Reset()
			}
		}

		if err := runCmd.Wait(); err != nil {
			fmt.Fprintln(os.Stderr, "Error waiting for command:", err)
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
	rootCmd.Flags().StringVar(&cfgFile, "config", "", "specify path to the config file")
	rootCmd.Flags().StringVar(&rulesDir, "rules-dir", "", "specify path to the rules directory")
	rootCmd.Flags().StringVar(&color, "color", "auto", "whether use color or not (never, auto, always)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
