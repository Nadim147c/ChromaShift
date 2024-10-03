package cmd_test

import (
	"bytes"
	"colorize/cmd"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"testing"
)

func TestReadIo(t *testing.T) {
	type Test struct {
		name     string
		stderr   bool
		output   string
		rules    cmd.CommandRules
		expected string
	}

	tests := []Test{
		{
			name:   "Test stdout handling with colorization",
			stderr: false,
			output: "test output",
			rules: cmd.CommandRules{
				Rules: []cmd.Rule{
					{Regexp: regexp.MustCompile("(test)"), Colors: "black"},
				},
			},
			expected: fmt.Sprintf("%stest%s output%s\n", cmd.Ansi.Black, cmd.Ansi.Reset, cmd.Ansi.Reset),
		},
		{
			name:   "Test stderr handling with no colorization",
			stderr: true,
			output: "error output",
			rules: cmd.CommandRules{
				Rules: []cmd.Rule{
					{Regexp: regexp.MustCompile("do not match"), Colors: "black"},
				},
			},
			expected: fmt.Sprintf("error output%s\n", cmd.Ansi.Reset),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defautlCmdRules := cmd.CmdRules
			defer func() { cmd.CmdRules = defautlCmdRules }()
			cmd.CmdRules = tt.rules

			var ioPipe io.ReadCloser
			var runCmd *exec.Cmd

			if tt.stderr {
				runCmd = exec.Command("sh", "-c", "echo '"+tt.output+"' >&2")
				ioPipe, _ = runCmd.StderrPipe()
			} else {
				runCmd = exec.Command("sh", "-c", "echo '"+tt.output+"'")
				ioPipe, _ = runCmd.StdoutPipe()
			}
			runCmd.Start()

			outputBuf := new(bytes.Buffer)

			cmd.DefaultReadIo(runCmd, outputBuf, ioPipe)

			runCmd.Wait()

			result := outputBuf.String()
			if result != tt.expected {
				t.Errorf("Expected output %q, but got %q", tt.expected, result)
			}
		})
	}
}
