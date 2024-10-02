package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/efekarakus/termcolor"
)

func ReadIoOnStderr(runCmd *exec.Cmd) {
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

	ReadIo(runCmd, os.Stderr, ioPipe)
}

func ReadIoOnStdout(runCmd *exec.Cmd) {
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
	ReadIo(runCmd, os.Stdout, ioPipe)
}

func ReadIo(runCmd *exec.Cmd, out io.Writer, ioPipe io.ReadCloser) {
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
			coloredLine := ColorizeLine(line, CmdRules.Rules)
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
			coloredLine := ColorizeLine(line, CmdRules.Rules)
			if len(coloredLine) > 0 {
				fmt.Fprint(out, coloredLine+"\n")
			} else {
				fmt.Fprint(out, line+"\n")
			}
			buffer.Reset()
		}
	}
}
