package ansi

import (
	"slices"
	"strings"
)

// Match defines the interface for ANSI styling matches within a string.
// Implementations should represent spans of text that require specific styling.
type Match interface {
	// Len returns the number of style matches this object contains.
	Len() int
	// Match returns the start position, end position, and ANSI sequence string
	// for the match at index i. If start equals end, the match is ignored.
	Match(i int) (start, end int, sequence string)
}

// ansiWriter accumulates ANSI escape sequences and text content, flushing
// sequences into the output at appropriate times.
type ansiWriter struct {
	buf   strings.Builder
	codes []string
}

// WriteString writes plain text content to the buffer, flushing any pending
// sequences first.
func (w *ansiWriter) WriteString(s string) (int, error) {
	n, err := w.flushSequence()
	if err != nil {
		return n, err
	}
	n2, err := w.buf.WriteString(s)
	return n + n2, err
}

// WriteSequence appends an ANSI sequence code to be flushed later.
func (w *ansiWriter) WriteSequence(s string) {
	if s == "" {
		return
	}
	size := len(w.codes)
	// checks if last ansi sequence is the same as current
	if size > 0 && w.codes[size-1] == s {
		return
	}
	w.codes = append(w.codes, s)
}

func (w *ansiWriter) buildSequence() string {
	var style Style
	var codes []int
	for s := range slices.Values(w.codes) {
		c := parseCodes(s)
		codes = append(codes, c...)
	}

	applyCodes(&style, codes)

	return "\x1b[" + style.Sequence() + "m"
}

// flushSequence writes accumulated sequences as a single ANSI escape code in
// the format "\x1b[codes;joined;by;semicolonm".
func (w *ansiWriter) flushSequence() (int, error) {
	if len(w.codes) == 0 {
		return 0, nil
	}
	n, err := w.buf.WriteString(w.buildSequence())
	w.codes = w.codes[:0] // reset slice
	return n, err
}

// String returns the final formatted string with all ANSI codes applied.
func (w *ansiWriter) String() string {
	w.flushSequence()
	return w.buf.String()
}

// event represents a style application or removal event at a specific position.
type event struct {
	id       uint
	pos      int    // position in the string
	start    bool   // true for style start, false for style end
	sequence string // ANSI sequence code (only populated for starts)
}

// eventSlice is a sortable collection of events used for processing matches.
type eventSlice []event

// sortEvents arranges events by position, with style starts before ends at the
// same position, and shorter sequences prioritized for consistent ordering.
func sortEvents(events eventSlice) {
	slices.SortFunc(events, func(a, b event) int {
		if a.pos != b.pos {
			return a.pos - b.pos
		}
		// Start events precede end events at the same position
		if a.start != b.start {
			if a.start {
				return -1
			}
			return 1
		}
		// Shorter sequences are prioritized for deterministic ordering
		return len(a.sequence) - len(b.sequence)
	})
}

// Colorize applies ANSI styling to a string based on Match specifications.
//
// It handles overlapping and nested matches by maintaining a style stack. When
// matches overlap, the implementation properly resets and reapplies styles to
// ensure visual correctness.
//
// The function works by:
//  1. Creating start and end events for each match
//  2. Sorting events by position (starts before ends at same position)
//  3. Building output by writing string segments and injecting ANSI codes
//  4. Maintaining a stack to handle nested style resets
//
// Parameters:
//   - s: the original string to colorize
//   - matches: a slice of Match objects defining style spans
//
// Returns the input string with ANSI escape codes inserted for all matched
// ranges. Empty match ranges (start == end) are skipped.
func Colorize(s string, matches []Match) string {
	if len(matches) == 0 {
		return s
	}

	// Extract and create events from all matches
	events := buildEvents(matches)
	if len(events) == 0 {
		return s
	}

	// Sort events by position
	sortEvents(events)

	// Build result with proper interleaving of text and ANSI codes
	result := &ansiWriter{}
	stack := Stack{}
	lastPos := 0

	for _, event := range events {
		// Write text between last event and current event
		if event.pos > lastPos {
			result.WriteString(s[lastPos:event.pos])
		}

		if event.start {
			// Apply new style
			stack.PushRaw(event.id, event.sequence)
			result.WriteSequence(event.sequence)
		} else {
			reseter := stack.Kick(event.id).Reset()
			result.WriteSequence(reseter.Sequence())
			result.WriteSequence(stack.Current().Sequence())
		}

		lastPos = event.pos
	}

	// Write remaining string content
	if lastPos < len(s) {
		result.WriteString(s[lastPos:])
	}

	return result.String()
}

var counter uint

func getID() uint {
	counter++
	return counter
}

// buildEvents extracts all start and end events from a slice of matches.
func buildEvents(matches []Match) eventSlice {
	counter = 0
	events := make(eventSlice, 0)

	for _, m := range matches {
		for i := range m.Len() {
			start, end, seq := m.Match(i)

			// Skip empty ranges
			if start == end {
				continue
			}

			id := getID()
			se := event{id: id, pos: start, start: true, sequence: seq}
			ee := event{id: id, pos: end, start: false}
			events = append(events, se, ee)
		}
	}

	return events
}
