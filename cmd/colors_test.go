package cmd_test

import (
	"colorize/cmd"
	"testing"
)

func TestGetColor(t *testing.T) {
	type TestCase struct {
		colorName     string
		expectedValue string
	}

	validTestCases := []TestCase{
		{"reset", cmd.Ansi.Reset},
		{"bold", cmd.Ansi.Bold},
		{"underline", cmd.Ansi.Underline},
		{"blink", cmd.Ansi.Blink},
		{"reverse", cmd.Ansi.Reverse},
		{"conceal", cmd.Ansi.Conceal},
		{"black", cmd.Ansi.Black},
		{"red", cmd.Ansi.Red},
		{"green", cmd.Ansi.Green},
		{"yellow", cmd.Ansi.Yellow},
		{"blue", cmd.Ansi.Blue},
		{"magenta", cmd.Ansi.Magenta},
		{"cyan", cmd.Ansi.Cyan},
		{"white", cmd.Ansi.White},
		{"gray", cmd.Ansi.Gray},
		{"bgblack", cmd.Ansi.BgBlack},
		{"bgred", cmd.Ansi.BgRed},
		{"bggreen", cmd.Ansi.BgGreen},
		{"bgyellow", cmd.Ansi.BgYellow},
		{"bgblue", cmd.Ansi.BgBlue},
		{"bgmagenta", cmd.Ansi.BgMagenta},
	}

	t.Run("Valid colors", func(t *testing.T) {
		for _, tc := range validTestCases {
			t.Run(tc.colorName, func(t *testing.T) {
				result := cmd.Ansi.GetColor(tc.colorName)
				if result != tc.expectedValue {
					t.Errorf("Expected %q for color %q, but got %q", tc.expectedValue, tc.colorName, result)
				}
			})
		}
	})

	t.Run("Invalid or edge case colors", func(t *testing.T) {
		edgeCases := []struct {
			colorName     string
			expectedValue string
		}{
			{"unknownColor", ""},          // Invalid color name
			{"  Bold  ", cmd.Ansi.Bold},   // Case-insensitive and trim spaces
			{"BgCyan", cmd.Ansi.BgCyan},   // Test camel-case, assuming BgCyan is defined
			{"BGWHITE", cmd.Ansi.BgWhite}, // Uppercase test
			{"bggray", cmd.Ansi.BgGray},   // Lowercase test
		}

		for _, tc := range edgeCases {
			t.Run(tc.colorName, func(t *testing.T) {
				result := cmd.Ansi.GetColor(tc.colorName)
				if result != tc.expectedValue {
					t.Errorf("Expected %q for color %q, but got %q", tc.expectedValue, tc.colorName, result)
				}
			})
		}
	})
}
