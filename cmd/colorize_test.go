package cmd_test

import (
	"colorize/cmd"
	"os"
	"reflect"
	"regexp"
	"testing"
)

func TestExtentColorMapFromMatches(t *testing.T) {
	t.Run("Basic color map extension with matches", func(t *testing.T) {
		colorMap := make(map[int]string)
		matches := [][]int{
			{0, 1, 2, 4, 10, 13},
			{1, 3, 7, 10},
		}
		colors := []string{"red", "blue"}

		expectedColorMap := map[int]string{
			2:  cmd.Ansi.Red,                   // 1st match 1st group
			4:  cmd.Ansi.Reset,                 // end of 1st match 1st group
			7:  cmd.Ansi.Red,                   // 2n match
			10: cmd.Ansi.Reset + cmd.Ansi.Blue, // end of 2nd match and start of 1st match's 2 group
			13: cmd.Ansi.Reset,
		}

		cmd.ExtentColorMapFromMatches(colorMap, matches, colors)

		if !reflect.DeepEqual(colorMap, expectedColorMap) {
			t.Errorf("Expected color map %v, but got %v", expectedColorMap, colorMap)
		}
	})
}

func TestExtentColorMapWithLsColors(t *testing.T) {
	envLsColors := os.Getenv("LS_COLORS")
	os.Setenv("LS_COLORS", "")
	defer func() { os.Setenv("LS_COLORS", envLsColors) }()
	t.Run("Basic color map extension with matches using for path type", func(t *testing.T) {
		path := "from '/path/to/main.go' to '/another/path/to/main.go'"

		colorMap := make(map[int]string)

		matches := [][]int{
			{0, 54, 6, 22},
		}

		expectedColorMap := map[int]string{
			6:  cmd.Ansi.Blue,
			15: cmd.Ansi.Cyan,
			22: cmd.Ansi.Reset,
		}

		cmd.ExtentColorMapWithLsColors(colorMap, matches, path)

		if !reflect.DeepEqual(colorMap, expectedColorMap) {
			t.Errorf("Expected color map %+v, but got %+v", expectedColorMap, colorMap)
		}
	})
}

func TestColorizeLine(t *testing.T) {
	// Subtest 1: Basic colorization
	t.Run("Basic colorization with rules", func(t *testing.T) {
		line := "hello world"
		rules := []cmd.Rule{
			{
				Regexp: regexp.MustCompile("(hello)"),
				Colors: "red",
			},
			{
				Regexp: regexp.MustCompile("(world)"),
				Colors: "blue",
			},
		}

		expectedOutput := "\033[31mhello\033[0m \033[34mworld\033[0m"

		coloredLine := cmd.ColorizeLine(line, rules)

		if coloredLine != expectedOutput {
			t.Errorf("Expected %q, but got %q", expectedOutput, coloredLine)
		}
	})

	// Subtest 2: Overwrite rule
	t.Run("Overwrite rule", func(t *testing.T) {
		line := "hello world"
		rules := []cmd.Rule{
			{
				Regexp:    regexp.MustCompile("(hello)"),
				Colors:    "red",
				Overwrite: true,
			},
			{
				Regexp: regexp.MustCompile("(world)"),
				Colors: "blue",
			},
		}

		expectedOutput := "\033[31mhello\033[0m world\033[0m"

		coloredLine := cmd.ColorizeLine(line, rules)

		if coloredLine != expectedOutput {
			t.Errorf("Expected %q, but got %q", expectedOutput, coloredLine)
		}
	})

	t.Run("No matches", func(t *testing.T) {
		line := "hello world"
		rules := []cmd.Rule{
			{
				Regexp: regexp.MustCompile("nothing"),
				Colors: "green",
			},
		}

		expectedOutput := "hello world\033[0m"

		coloredLine := cmd.ColorizeLine(line, rules)

		if coloredLine != expectedOutput {
			t.Errorf("Expected %q, but got %q", expectedOutput, coloredLine)
		}
	})
}
