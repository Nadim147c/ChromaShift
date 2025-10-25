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

// Option represents a tri-state flag: Active, Inactive, or Unset.
type Option int8

const (
	False     Option = -1
	Undefined Option = 0
	True      Option = 1
)

// Switch returns `a` if True, `b` if False, and an empty string if Undefined.
func (o Option) Switch(a, b string) string {
	switch o {
	case True:
		return a
	case False:
		return b
	default:
		return ""
	}
}

// Inverted toggles between True and Flase. Undefined remains unchanged.
func (o Option) Inverted() Option {
	switch o {
	case True:
		return False
	case False:
		return True
	default:
		return Undefined
	}
}

// IsSet reports whether the Optional has a definite (non-Unset) value.
func (o Option) IsSet() bool {
	return o != Undefined
}

// String implements fmt.Stringer for easier debugging/logging.
func (o Option) String() string {
	switch o {
	case True:
		return "true"
	case False:
		return "false"
	default:
		return "undefined"
	}
}

// resetColor is a sentinel type that represents a color reset operation.
type resetColor struct{}

var _ termenv.Color = (*resetColor)(nil)

// Sequence returns the ANSI code to reset color.
// If bg is true, returns the background color reset code, otherwise foreground.
func (r resetColor) Sequence(bg bool) string {
	if bg {
		return "49" // reset background color
	}
	return "39" // reset foreground color
}

// sequencedColor is ansi true color as sequence
type sequencedColor struct {
	r, g, b uint8
}

var _ termenv.Color = (*sequencedColor)(nil)

func (s sequencedColor) Sequence(bg bool) string {
	prefix := termenv.Foreground
	if bg {
		prefix = termenv.Background
	}
	return fmt.Sprintf("%s;2;%d;%d;%d", prefix, s.r, s.g, s.b)
}

func newSequencedColor(r, g, b int) sequencedColor {
	return sequencedColor{uint8(r), uint8(g), uint8(b)}
}

// Style represents a complete ANSI text styling configuration. Fields with
// pointer types (Bold, Italic, Underline, Inverse) use nil to indicate "no
// change" (attribute not specified), false to indicate "disable", and true to
// indicate "enable".
type Style struct {
	ID uint
	// If true, resets all attributes
	HardReset bool
	// Foreground color (nil if not set)
	Fg termenv.Color
	// Background color (nil if not set)
	Bg termenv.Color
	// Bold attribute (nil=unset, true=bold, false=not bold)
	Bold Option
	// Italic attribute (nil=unset, true=italic, false=not italic)
	Italic Option
	// Underline attribute (nil=unset, true=underline, false=not underline)
	Underline Option
	// Inverse attribute (nil=unset, true=inverse, false=not inverse)
	Inverse Option
}

func NewStyle(id uint, raw string) Style {
	s := Style{ID: id}
	codes := parseCodes(raw)
	if len(codes) == 0 {
		return s
	}
	applyCodes(&s, codes)
	return s
}

// Sequence returns the ANSI SGR sequence string without escape prefix/suffix.
// Example output: "1;38;2;255;100;0;48;5;42"
func (s Style) Sequence() string {
	if s.HardReset {
		return "0"
	}

	parts := make([]string, 0, 6)

	parts = append(parts, s.Bold.Switch("1", "22"))
	parts = append(parts, s.Italic.Switch("3", "23"))
	parts = append(parts, s.Underline.Switch("4", "24"))
	parts = append(parts, s.Inverse.Switch("7", "27"))

	// Add color codes
	if s.Fg != nil {
		parts = append(parts, s.Fg.Sequence(false))
	}
	if s.Bg != nil {
		parts = append(parts, s.Fg.Sequence(false))
	}

	// Remove empty sequences and return
	parts = slices.DeleteFunc(parts, func(p string) bool { return p == "" })

	if len(parts) == 0 {
		return "0"
	}

	return strings.Join(parts, ";")
}

// Reset returns a new Style with inverted attributes that would undo this
// style. If s.HardReset is true, returns an empty style. For Option attributes:
// true becomes false and vice versa. For colors: they are replaced with reset
// color codes.
func (s Style) Reset() Style {
	inverted := Style{ID: s.ID}

	// If this is a full reset, return an empty style
	if s.HardReset {
		return inverted
	}

	// Invert boolean attributes
	inverted.Bold = s.Bold.Inverted()
	inverted.Italic = s.Italic.Inverted()
	inverted.Underline = s.Underline.Inverted()
	inverted.Inverse = s.Inverse.Inverted()

	// Replace colors with reset codes
	if s.Fg != nil {
		inverted.Fg = resetColor{}
	}
	if s.Bg != nil {
		inverted.Bg = resetColor{}
	}

	return inverted
}

// Stack manages a stack of Style objects to handle nested and overlapping
// style regions. It tracks the current active style based on the stack state.
type Stack struct {
	layers []Style
}

// Len returns the number of styles currently in the stack.
func (s Stack) Len() int {
	return len(s.layers)
}

// Push adds a new style to the top of the stack.
func (s *Stack) Push(id uint, style Style) Style {
	style.ID = id
	s.layers = append(s.layers, style)
	return style
}

// Kick removes the Style with the given id and returns it. If not found, it
// returns the zero value.
func (s *Stack) Kick(id uint) (out Style) {
	for i := len(s.layers) - 1; i >= 0; i-- {
		if s.layers[i].ID == id {
			out = s.layers[i]
			s.layers = append(s.layers[:i], s.layers[i+1:]...)
			return out
		}
	}
	return out
}

// Current returns the currently active style at the top of the stack. Returns
// an empty Style if the stack is empty.
func (s Stack) Current() Style {
	if len(s.layers) == 0 {
		return Style{HardReset: true}
	}
	return s.layers[len(s.layers)-1]
}

// PushRaw parses an ANSI SGR sequence string and pushes the resulting Style
// onto the stack. The raw string should be in the format used by ANSI codes,
// e.g., "1;31;4" for bold red underline.
func (s *Stack) PushRaw(id uint, raw string) Style {
	return s.Push(id, NewStyle(id, raw))
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
func applyCodes(st *Style, codes []int) {
	var hasReset bool
	// only reset if the codes include reset code 0. Importanly, true color can
	// have 0 as code RGB value. Thus, we have to check if there is 'actual'
	// reset code to enable reset. If not than we set reset to false
	defer func() { st.HardReset = hasReset }()
	for i := 0; i < len(codes); i++ {
		c := codes[i]
		// Text attributes
		switch {
		case c == 0:
			hasReset = true
		case c == 1:
			st.Bold = True
		case c == 3:
			st.Italic = True
		case c == 4:
			st.Underline = True
		case c == 7:
			st.Inverse = True
		case c == 22:
			st.Bold = False
		case c == 23:
			st.Italic = False
		case c == 24:
			st.Underline = False
		case c == 27:
			st.Inverse = False
		case c == 39:
			st.Fg = resetColor{}
		case c == 49:
			st.Bg = resetColor{}
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
						col := newSequencedColor(r, g, b)
						if isFg {
							st.Fg = col
						} else {
							st.Bg = col
						}
						i += 4
					}
				}
			}
		}
	}
}
