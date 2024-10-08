package cmd_test

import (
	"cshift/cmd"
	"testing"
)

func TestGetColor(t *testing.T) {
	t.Run("Get color", func(t *testing.T) {
		ansiColor := cmd.Ansi.GetColor("red")
		if ansiColor != cmd.Ansi.Red {
			t.Fatalf("expected %s, but got %s", printAnsi(cmd.Ansi.Red), printAnsi(ansiColor))
		}
	})

	t.Run("Get color with all upper case", func(t *testing.T) {
		ansiColor := cmd.Ansi.GetColor("CYAN")
		if ansiColor != cmd.Ansi.Cyan {
			t.Fatalf("expected %s, but got %s", printAnsi(cmd.Ansi.Cyan), printAnsi(ansiColor))
		}
	})

	t.Run("Get color with ambiguous case", func(t *testing.T) {
		ansiColor := cmd.Ansi.GetColor("bLUe")
		if ansiColor != cmd.Ansi.Blue {
			t.Fatalf("expected %s, but got %s", printAnsi(cmd.Ansi.Blue), printAnsi(ansiColor))
		}
	})

	t.Run("Get color with ambiguous text", func(t *testing.T) {
		ansiColor := cmd.Ansi.GetColor(" GrEEn   ")
		if ansiColor != cmd.Ansi.Green {
			t.Fatalf("expected %s, but got %s", printAnsi(cmd.Ansi.Green), printAnsi(ansiColor))
		}
	})

	t.Run("Get color empty on invalid_color", func(t *testing.T) {
		ansiColor := cmd.Ansi.GetColor("invalid_color")
		if ansiColor != "" {
			t.Fatalf("expected %s, but got %s", "Empty String", printAnsi(ansiColor))
		}
	})
}
