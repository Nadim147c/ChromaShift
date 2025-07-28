package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"unicode"
	"unicode/utf8"

	"github.com/creack/pty"
)

type Output struct {
	Command *exec.Cmd
	Out     *os.File
	Buffer  bytes.Buffer
	Rules   []Rule
}

func NewOutput(cmd *exec.Cmd, rules []Rule, strerr bool) *Output {
	var out *os.File
	if strerr {
		out = os.Stderr
	} else {
		out = os.Stdout
	}
	output := Output{Command: cmd, Rules: rules, Out: out}
	return &output
}

func (o *Output) Write(char rune) {
	if char == '\r' {
		line := o.Buffer.String()
		coloredLine := ColorizeLine(line, o.Rules)
		if len(coloredLine) > 0 {
			fmt.Fprint(o.Out, coloredLine+"\r")
		} else {
			fmt.Fprint(o.Out, line+"\r")
		}
		o.Buffer.Reset()
	} else {
		runeBytes := make([]byte, utf8.RuneLen(char))
		utf8.EncodeRune(runeBytes, char)
		for _, b := range runeBytes {
			o.Buffer.WriteByte(b)
		}
	}

	if char == '\n' {
		line := strings.TrimRightFunc(o.Buffer.String(), unicode.IsSpace)
		coloredLine := ColorizeLine(line, o.Rules)
		if len(coloredLine) > 0 {
			fmt.Fprint(o.Out, coloredLine+"\n")
		} else {
			fmt.Fprint(o.Out, line+"\n")
		}
		o.Buffer.Reset()
	}
}

func (o *Output) Start(stderr bool) {
	if Color != "always" && !isTerminal(o.Out) {
		startRunWithoutColor(o.Command)
		os.Exit(0)
	}

	var ioPipe io.ReadCloser
	var err error
	if !stderr {
		o.Command.Stderr = os.Stderr
		ioPipe, err = o.Command.StdoutPipe()
		if err != nil {
			slog.Debug("Error creating stdout pipe", "error", err)
			os.Exit(1)
		}
	} else {
		o.Command.Stdout = os.Stdout
		ioPipe, err = o.Command.StderrPipe()
		if err != nil {
			slog.Debug("Error creating stderr pipe", "error", err)
			os.Exit(1)
		}
	}

	if err := o.Command.Start(); err != nil {
		slog.Debug("Error starting command", "error", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(ioPipe)
	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			slog.Debug("Error reading from pipe", "error", err)
			break
		}

		o.Write(char)
	}
}

func (o *Output) StartWithPTY(stderr bool) {
	if Color != "always" && (!isTerminal(os.Stderr) || !isTerminal(os.Stdout)) {
		startRunWithoutColor(o.Command)
		os.Exit(0)
	}

	ptmx, err := pty.Start(o.Command)
	if err != nil {
		slog.Debug("Error starting command with pty", "error", err)
	} else {
		defer ptmx.Close()
	}

	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	defer func() { _ = ptmx.Close() }()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				slog.Debug("Error resizing pty", "error", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH
	defer func() { signal.Stop(ch); close(ch) }()

	buf := make([]byte, 1024)
	var partialBuf []byte

	for {
		n, err := ptmx.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			slog.Debug("Error reading from pty", "error", err)
			break
		}

		partialBuf = append(partialBuf, buf[:n]...)

		for len(partialBuf) > 0 {
			r, size := utf8.DecodeRune(partialBuf)

			if r == utf8.RuneError && size == 1 {
				break
			}

			o.Write(r)
			partialBuf = partialBuf[size:]
		}
	}
}
