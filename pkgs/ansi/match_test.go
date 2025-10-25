package ansi

import (
	"fmt"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func printDiff(t *testing.T, expected, got string) {
	dmp := diffmatchpatch.New()

	t.Logf("expected: '%s\x1b[0m' got: '%s\x1b[0m'", expected, got)
	expected = fmt.Sprintf("%q", expected)
	got = fmt.Sprintf("%q", got)
	diffs := dmp.DiffMain(expected, got, false)

	t.Errorf(
		"expected: %s got: %s\ndiff %s",
		expected,
		got,
		dmp.DiffPrettyText(diffs),
	)
}

// MockMatches implements the Matches interface for testing
type MockMatches struct {
	matches []struct {
		start, end int
		sequence   string
	}
}

func (m *MockMatches) Len() int {
	return len(m.matches)
}

func (m *MockMatches) Match(i int) (start, end int, sequence string) {
	if i < 0 || i >= len(m.matches) {
		return 0, 0, ""
	}
	return m.matches[i].start, m.matches[i].end, m.matches[i].sequence
}

func (m *MockMatches) AddMatch(start, end int, sequence string) {
	m.matches = append(m.matches, struct {
		start, end int
		sequence   string
	}{start, end, sequence})
}

func TestColorizeEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		matches  *MockMatches
		expected string
	}{
		{
			name:     "no matches",
			input:    "hello world",
			matches:  &MockMatches{},
			expected: "hello world",
		},
		{
			name:     "empty string",
			input:    "",
			matches:  &MockMatches{},
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Colorize(tt.input, []Match{tt.matches})
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestColorizeSingleMatch(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		matches  *MockMatches
		expected string
	}{
		{
			name:  "single bold match at start",
			input: "hello world",
			matches: func() *MockMatches {
				m := &MockMatches{}
				m.AddMatch(0, 5, "1") // bold "hello"
				return m
			}(),
			expected: "\x1b[1mhello\x1b[0m world",
		},
		{
			name:  "single red foreground match in middle",
			input: "hello world",
			matches: func() *MockMatches {
				m := &MockMatches{}
				m.AddMatch(6, 11, "31") // red "world"
				return m
			}(),
			expected: "hello \x1b[31mworld\x1b[0m",
		},
		{
			name:  "single match covering entire string",
			input: "hello",
			matches: func() *MockMatches {
				m := &MockMatches{}
				m.AddMatch(0, 5, "1;4") // bold and underline
				return m
			}(),
			expected: "\x1b[1;4mhello\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Colorize(tt.input, []Match{tt.matches})
			if result != tt.expected {
				printDiff(t, tt.expected, result)
			}
		})
	}
}

func TestColorizeNested(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		matches  *MockMatches
		expected string
	}{
		{
			name:  "nested matches:fully contained",
			input: "hello world",
			matches: func() *MockMatches {
				m := &MockMatches{}
				m.AddMatch(0, 11, "1")  // bold entire string
				m.AddMatch(6, 11, "31") // red "world"
				return m
			}(),
			expected: "\x1b[1mhello \x1b[31mworld\x1b[0m",
		},
		{
			name:  "overlapping matches",
			input: "hello world",
			matches: func() *MockMatches {
				m := &MockMatches{}
				m.AddMatch(0, 8, "1")   // bold "hello wo"
				m.AddMatch(6, 11, "31") // red "world"
				return m
			}(),
			expected: "\x1b[1mhello \x1b[31mwo\x1b[22;31mrld\x1b[0m",
		},
		{
			name:  "three level nesting",
			input: "abcdefghij",
			matches: func() *MockMatches {
				m := &MockMatches{}
				m.AddMatch(0, 10, "1") // bold all
				m.AddMatch(3, 7, "31") // red "defg"
				m.AddMatch(4, 6, "4")  // underline "ef"
				return m
			}(),
			expected: "\x1b[1mabc\x1b[31md\x1b[4mef\x1b[24;31mg\x1b[1;39mhij\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Colorize(tt.input, []Match{tt.matches})
			if result != tt.expected {
				printDiff(t, tt.expected, result)
			}
		})
	}
}

func TestColorizeMultipleNonOverlapping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		matches  *MockMatches
		expected string
	}{
		{
			name:  "multiple separate matches",
			input: "hello world foo bar",
			matches: func() *MockMatches {
				m := &MockMatches{}
				m.AddMatch(0, 5, "1")    // bold "hello"
				m.AddMatch(6, 11, "31")  // red "world"
				m.AddMatch(12, 15, "32") // green "foo"
				return m
			}(),
			expected: "\x1b[1mhello\x1b[0m \x1b[31mworld\x1b[0m \x1b[32mfoo\x1b[0m bar",
		},
		{
			name:  "adjacent matches",
			input: "ab",
			matches: func() *MockMatches {
				m := &MockMatches{}
				m.AddMatch(0, 1, "1")  // bold "a"
				m.AddMatch(1, 2, "31") // red "b"
				return m
			}(),
			expected: "\x1b[1ma\x1b[22;31mb\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Colorize(tt.input, []Match{tt.matches})
			if result != tt.expected {
				printDiff(t, tt.expected, result)
			}
		})
	}
}

func TestColorizeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		matches  *MockMatches
		expected string
	}{
		{
			name:  "zero-width match",
			input: "hello",
			matches: func() *MockMatches {
				m := &MockMatches{}
				m.AddMatch(2, 2, "1") // zero-width match
				return m
			}(),
			expected: "hello",
		},
		{
			name:  "match at string boundary - start",
			input: "hello",
			matches: func() *MockMatches {
				m := &MockMatches{}
				m.AddMatch(0, 0, "1")
				return m
			}(),
			expected: "hello",
		},
		{
			name:  "match at string boundary - end",
			input: "hello",
			matches: func() *MockMatches {
				m := &MockMatches{}
				m.AddMatch(5, 5, "1")
				return m
			}(),
			expected: "hello",
		},
		{
			name:  "complex sequence with 24-bit color",
			input: "text",
			matches: func() *MockMatches {
				m := &MockMatches{}
				m.AddMatch(0, 4, "1;38;2;255;100;0") // bold + RGB color
				return m
			}(),
			expected: "\x1b[1;38;2;255;100;0mtext\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Colorize(tt.input, []Match{tt.matches})
			if result != tt.expected {
				printDiff(t, tt.expected, result)
			}
		})
	}
}
