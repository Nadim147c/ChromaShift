package cmd

import (
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/MatusOllah/slogcolor"
	"github.com/carapace-sh/carapace"
	termcolor "github.com/fatih/color"
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

var Version = "dev"

var (
	Color          string
	ConfigFile     string
	RulesDirectory string
	Debug          bool
	UseColor       bool
)

func init() {
	rootCmd.SetErrPrefix("ChromaShift Error:")
	rootCmd.Flags().StringVar(&ConfigFile, "config", "", "specify path to the config file")
	rootCmd.Flags().StringVar(&RulesDirectory, "rules-dir", "", "specify path to the rules directory")
	rootCmd.Flags().StringVar(&Color, "color", "auto", "whether use color or not (never, auto, always)")
	rootCmd.Flags().BoolVarP(&Debug, "debug", "d", false, "verbose output")
	carapace.Gen(rootCmd)
}

func isTerminal(f *os.File) bool {
	return termenv.NewOutput(f).EnvColorProfile() == termenv.TrueColor
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
	PreRun: func(cmd *cobra.Command, args []string) {
		switch Color {
		case "never":
			UseColor = false
		case "always":
			UseColor = true
		default:
			UseColor = isTerminal(os.Stdout)
		}

		opts := slogcolor.DefaultOptions
		if Debug {
			opts.Level = slog.LevelDebug
		} else {
			opts.Level = slog.LevelInfo + 1000
			return
		}

		termcolor.NoColor = termenv.NewOutput(os.Stderr).EnvNoColor()
		opts.NoTime = true
		opts.SrcFileMode = 0
		opts.LevelTags = map[slog.Level]string{
			slog.LevelDebug: termcolor.New(termcolor.FgGreen).Sprint("ChromaShift"),
			slog.LevelInfo:  termcolor.New(termcolor.FgCyan).Sprint("ChromaShift"),
			slog.LevelWarn:  termcolor.New(termcolor.FgYellow).Sprint("ChromaShift"),
			slog.LevelError: termcolor.New(termcolor.FgRed).Sprint("ChromaShift"),
		}

		slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, opts)))
	},
	Run: func(cmd *cobra.Command, args []string) {
		UseColor = true

		if len(args) < 1 {
			cmd.Help()
			os.Exit(0)
		}

		cmdName := args[0]
		cmdArgs := args[1:]

		runCmd := exec.Command(cmdName, cmdArgs...)

		if !UseColor {
			startRunWithoutColor(runCmd)
		}

		config, err := LoadConfig()
		if err != nil {
			slog.Debug("Failed to load config", "error", err)
		}

		ruleFileName, err := GetRuleFileName(config, args)
		if err != nil {
			slog.Debug("No config exists for current command")
			startRunWithoutColor(runCmd)
		} else {
			slog.Debug("Rules file name", "name", ruleFileName)
		}

		cmdRules, err := LoadRules(ruleFileName)
		if err != nil {
			slog.Debug("Failed to load rules for current command", "error", err)
			startRunWithoutColor(runCmd)
		}

		if cmdRules.Rules == nil {
			slog.Debug("No config exists for current command")
			startRunWithoutColor(runCmd)
		}

		slog.Debug("Rules found", "count", len(cmdRules.Rules))

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigChan
			if err := runCmd.Process.Signal(sig); err != nil {
				slog.Debug("Error sending signal to process", "error", err)
			}
		}()

		if cmdRules.PTY {
			outputReader := NewOutput(runCmd, cmdRules.Rules, cmdRules.Stderr)
			outputReader.StartWithPTY(cmdRules.Stderr)
		} else {
			runCmd.Stdin = os.Stdin
			outputReader := NewOutput(runCmd, cmdRules.Rules, cmdRules.Stderr)
			outputReader.Start(cmdRules.Stderr)
		}

		if err := runCmd.Wait(); err != nil {
			slog.Debug("Error waiting for command", "error", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	cc.Init(&cc.Config{
		RootCmd:         rootCmd,
		Headings:        cc.Cyan + cc.Bold + cc.Underline,
		Commands:        cc.Yellow + cc.Bold,
		CmdShortDescr:   cc.Bold,
		Example:         cc.Italic,
		ExecName:        cc.Bold,
		Flags:           cc.Green + cc.Bold,
		FlagsDataType:   cc.Red + cc.Bold,
		NoExtraNewlines: true,
	})

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
