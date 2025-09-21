package cmd_test

import (
	"cshift/cmd"
	"os"
	"testing"

	"github.com/gobwas/glob"
)

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

		if err != nil || (lsColor != "36") {
			t.Fatalf("expected %s, but got %s", "36", lsColor)
		}
	})

	t.Run("Get LS_COLORS from env LS_COLORS", func(t *testing.T) {
		os.Setenv("LS_COLORS", "*.go=31")
		cmd.LsColorsMap = make([]cmd.LsColor, 0)

		lsColor, err := cmd.GetLsColor("main.go")

		if err != nil || lsColor != "31" {
			t.Fatalf("expected %s, but got %s", "31", lsColor)
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

		if err != nil || lsColor != "32" {
			t.Fatalf("expected %s, but got %s", "32", lsColor)
		}
	})
}
