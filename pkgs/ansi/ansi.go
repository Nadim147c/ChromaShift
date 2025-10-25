// Package ansi provides context-aware ANSI coloring based on regex match
// patterns. It supports complex text styling with foreground/background colors,
// bold, italic, underline, and inverse attributes using ANSI SGR (Select
// Graphic Rendition) codes.
package ansi

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/muesli/termenv"
)

// crazy hack?
func boolptr(b bool) *bool     { return &b }
func invBoolptr(b *bool) *bool { a := !(*b); return &a }

// Pre-allocated boolean pointers for efficient reuse.
var (
	trueptr  = boolptr(true)
	falseptr = boolptr(false)
)

// resetColor is a sentinel type that represents a color reset operation.
type resetColor struct{}

// Sequence returns the ANSI code to reset color.
// If bg is true, returns the background color reset code, otherwise foreground.
func (r resetColor) Sequence(bg bool) string {
	if bg {
		return "49" // reset background color
	}
	return "39" // reset foreground color
}

// Style represents a complete ANSI text styling configuration. Fields with
// pointer types (Bold, Italic, Underline, Inverse) use nil to indicate "no
// change" (attribute not specified), false to indicate "disable", and true to
// indicate "enable".
type Style struct {
	// If true, resets all attributes
	Reset bool
	// Foreground color (nil if not set)
	Fg termenv.Color
	// Background color (nil if not set)
	Bg termenv.Color
	// Bold attribute (nil=unset, true=bold, false=not bold)
	Bold *bool
	// Italic attribute (nil=unset, true=italic, false=not italic)
	Italic *bool
	// Underline attribute (nil=unset, true=underline, false=not underline)
	Underline *bool
	// Inverse attribute (nil=unset, true=inverse, false=not inverse)
	Inverse *bool
}

// Sequence returns the ANSI SGR sequence string without escape prefix/suffix.
// Example output: "1;38;2;255;100;0;48;5;42"
func (s *Style) Sequence() string {
	if s.Reset {
		return "0"
	}

	var parts []string

	// Add text attribute codes
	if s.Bold != nil {
		if *s.Bold {
			parts = append(parts, "1")
		} else {
			parts = append(parts, "22") // reset bold
		}
	}
	if s.Italic != nil {
		if *s.Italic {
			parts = append(parts, "3")
		} else {
			parts = append(parts, "23") // reset italic
		}
	}
	if s.Underline != nil {
		if *s.Underline {
			parts = append(parts, "4")
		} else {
			parts = append(parts, "24") // reset underline
		}
	}
	if s.Inverse != nil {
		if *s.Inverse {
			parts = append(parts, "7")
		} else {
			parts = append(parts, "27") // reset inverse
		}
	}

	// Add color codes
	if s.Fg != nil {
		seq := s.Fg.Sequence(false)
		parts = append(parts, seq)
	}
	if s.Bg != nil {
		seq := s.Bg.Sequence(true)
		parts = append(parts, seq)
	}

	// Remove empty sequences and return
	parts = slices.DeleteFunc(parts, func(p string) bool {
		return p == ""
	})

	if len(parts) == 0 {
		return "0"
	}

	return strings.Join(parts, ";")
}

// FullEscape returns the complete ANSI escape sequence including escape prefix
// and suffix. Example output: "\x1b[1;31m"
func (s *Style) FullEscape() string {
	return "\x1b[" + s.Sequence() + "m"
}

// Invert returns a new Style with inverted attributes that would undo this
// style. If s.Reset is true, returns an empty style. For boolean attributes:
// true becomes false and vice versa. For colors: they are replaced with reset
// color codes.
func (s *Style) Invert() *Style {
	inverted := new(Style)

	// If this is a full reset, return an empty style
	if s.Reset {
		return inverted
	}

	// Invert boolean attributes
	if s.Bold != nil {
		inverted.Bold = invBoolptr(s.Bold)
	}
	if s.Italic != nil {
		inverted.Italic = invBoolptr(s.Italic)
	}
	if s.Underline != nil {
		inverted.Underline = invBoolptr(s.Underline)
	}
	if s.Inverse != nil {
		inverted.Inverse = invBoolptr(s.Inverse)
	}

	// Replace colors with reset codes
	if s.Fg != nil {
		inverted.Fg = resetColor{}
	}
	if s.Bg != nil {
		inverted.Bg = resetColor{}
	}

	return inverted
}

// StyleStack manages a stack of Style objects to handle nested and overlapping
// style regions. It tracks the current active style based on the stack state.
type StyleStack struct {
	layers []*Style
}

// Len returns the number of styles currently in the stack.
func (s *StyleStack) Len() int {
	return len(s.layers)
}

// Push adds a new style to the top of the stack.
func (s *StyleStack) Push(st *Style) {
	s.layers = append(s.layers, st)
}

// Pop removes the top style from the stack, if present.
func (s *StyleStack) Pop() {
	if len(s.layers) > 0 {
		s.layers = s.layers[:len(s.layers)-1]
	}
}

// Current returns the currently active style at the top of the stack. Returns
// an empty Style if the stack is empty.
func (s *StyleStack) Current() *Style {
	if len(s.layers) == 0 {
		return &Style{}
	}
	return s.layers[len(s.layers)-1]
}

// PushRaw parses an ANSI SGR sequence string and pushes the resulting Style
// onto the stack. The raw string should be in the format used by ANSI codes,
// e.g., "1;31;4" for bold red underline.
func (s *StyleStack) PushRaw(raw string) {
	codes := parseCodes(raw)
	if len(codes) == 0 {
		return
	}
	st := applyCodes(codes)
	s.Push(st)
}

// parseCodes converts an ANSI SGR sequence string into a slice of integer
// codes. Handles semicolon-separated values. Returns [0] for empty input.
func parseCodes(seq string) []int {
	if seq == "" {
		return []int{0}
	}

	parts := strings.Split(seq, ";")
	codes := make([]int, 0, len(parts))

	for _, p := range parts {
		if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
			codes = append(codes, n)
		}
	}

	return codes
}

// applyCodes converts ANSI SGR codes into a Style object.
// Supports:
//   - Text attributes: bold (1), italic (3), underline (4), inverse (7)
//   - Standard colors: 30-37 (fg), 40-47 (bg), 90-97 (bright fg), 100-107
//     (bright bg)
//   - 8-bit colors: 38;5;N (fg), 48;5;N (bg)
//   - 24-bit truecolor: 38;2;R;G;B (fg), 48;2;R;G;B (bg)
func applyCodes(codes []int) *Style {
	st := new(Style)

	for i := 0; i < len(codes); i++ {
		c := codes[i]

		switch {
		// Text attributes
		case c == 1:
			st.Bold = trueptr
		case c == 3:
			st.Italic = trueptr
		case c == 4:
			st.Underline = trueptr
		case c == 7:
			st.Inverse = trueptr
		case c == 22:
			st.Bold = falseptr
		case c == 23:
			st.Italic = falseptr
		case c == 24:
			st.Underline = falseptr

		// Standard foreground colors (30-37)
		case 30 <= c && c <= 37:
			st.Fg = termenv.ANSIColor(c - 30)

		// Standard background colors (40-47)
		case 40 <= c && c <= 47:
			st.Bg = termenv.ANSIColor(c - 40)

		// Bright foreground colors (90-97)
		case 90 <= c && c <= 97:
			st.Fg = termenv.ANSIColor(c - 82)

		// Bright background colors (100-107)
		case 100 <= c && c <= 107:
			st.Bg = termenv.ANSIColor(c - 92)

		// 8-bit and 24-bit color modes (38 for fg, 48 for bg)
		case c == 38 || c == 48:
			isFg := c == 38

			if i+1 < len(codes) {
				switch codes[i+1] {
				case 5: // 8-bit color mode
					if i+2 < len(codes) {
						colorIdx := codes[i+2]
						if isFg {
							st.Fg = termenv.ANSI256Color(colorIdx)
						} else {
							st.Bg = termenv.ANSI256Color(colorIdx)
						}
						i += 2
					}

				case 2: // 24-bit truecolor mode
					if i+4 < len(codes) {
						r := codes[i+2]
						g := codes[i+3]
						b := codes[i+4]
						col := termenv.RGBColor(
							fmt.Sprintf("#%0X%0X%0X", r, g, b),
						)
						if isFg {
							st.Fg = col
						} else {
							st.Bg = col
						}
						i += 4
					}
				}
			}

		// Unknown codes are silently ignored
		default:
		}
	}

	return st
}
