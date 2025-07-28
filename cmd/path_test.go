package cmd_test

import (
	"cshift/cmd"
	"os"
	"strings"
	"testing"

	"github.com/gobwas/glob"
)

func printAnsi(s string) string {
	return strings.ReplaceAll(s, "\033", "\\033")
}

func TestGetLsColor(t *testing.T) {
	LsColorsMap := cmd.LsColorsMap
	DefaultLsColors := cmd.DefaultLsColors
	defer func() {
		cmd.LsColorsMap = LsColorsMap
		cmd.DefaultLsColors = DefaultLsColors
	}()

	t.Run("Get LS_COLORS from built-in LS_COLORS", func(t *testing.T) {
		os.Setenv("LS_COLORS", "")
		cmd.LsColorsMap = make([]cmd.LsColor, 0)

		lsColor, err := cmd.GetLsColor("main.go")

		if err != nil || (lsColor != "\033[36m") {
			t.Fatalf("expected %s, but got %s", printAnsi("\033[36m"), printAnsi(lsColor))
		}
	})

	t.Run("Get LS_COLORS from env LS_COLORS", func(t *testing.T) {
		os.Setenv("LS_COLORS", "*.go=31")
		cmd.LsColorsMap = make([]cmd.LsColor, 0)

		lsColor, err := cmd.GetLsColor("main.go")

		if err != nil || lsColor != "\033[31m" {
			t.Fatalf("expected %s, but got %s", printAnsi("\033[31m"), printAnsi(lsColor))
		}
	})

	t.Run("Get LS_COLORS from complied LS_COLORS", func(t *testing.T) {
		os.Setenv("LS_COLORS", "*.go=31") // even though go code 31

		// It should use complied go=32
		cmd.LsColorsMap = []cmd.LsColor{{
			Glob: glob.MustCompile("*.go"),
			Code: "32",
		}}

		lsColor, err := cmd.GetLsColor("main.go")

		if err != nil || lsColor != "\033[32m" {
			t.Fatalf("expected %s, but got %s", printAnsi("\033[32m"), printAnsi(lsColor))
		}
	})
}
