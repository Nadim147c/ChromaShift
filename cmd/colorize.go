package cmd

import (
	"iter"
	"log/slog"
	"slices"
	"strings"

	"github.com/muesli/termenv"
)

type Match [3]int

func (m Match) Values() (int, int, int) {
	return m[0], m[1], m[2]
}

func RegexMatches(matches [][]int) iter.Seq[Match] {
	return func(yield func(Match) bool) {
		for _, match := range matches {
			n := len(match) / 2
			for i := range n {
				start := match[i*2]
				end := match[i*2+1]
				if start < 0 || end < 0 {
					continue // skip unmatched groups
				}
				if !yield(Match{i, start, end}) {
					return
				}
			}
		}
	}
}

type Index map[int][]string

func (i *Index) Reset() {
	*i = make(Index)
}

func (i Index) AddStyle(idx int, style ...string) {
	i[idx] = append(i[idx], style...)
}

func (i Index) ResetStyle(idx int) {
	i.AddStyle(idx, termenv.ResetSeq)
}

func (i Index) Extent(line string, matches [][]int, colors []string) {
loop:
	for match := range RegexMatches(matches) {
		idx, start, end := match.Values()

		cfgStyle := strings.TrimSpace(colors[idx%len(colors)])

		for style := range strings.SplitSeq(cfgStyle, " ") {
			style = strings.ToLower(strings.TrimSpace(style))
			if style == "path" {
				i.ExtentPath(line, start, end)
				continue loop
			}

			seq := GetColorCode(style)
			i.AddStyle(start, seq)
		}

		i.ResetStyle(end)
	}
}

func (i Index) ExtentPath(line string, start, end int) {
	path := line[start:end]

	slog.Debug("Path", "value", path)

	basePathIndex := start
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			basePathIndex = start + i + 1
			break
		}
	}

	i.AddStyle(start, termenv.ANSIBlue.Sequence(false))

	meta, metaErr := GetFileMetadata(path)

	if metaErr == nil && meta.IsEveyone {
		i.AddStyle(basePathIndex, termenv.ANSIGreen.Sequence(false), termenv.BoldSeq)
		i.AddStyle(end, termenv.ResetSeq)
		return
	}

	if metaErr == nil && meta.IsExecutable {
		i.AddStyle(basePathIndex, termenv.ANSIRed.Sequence(false), termenv.BoldSeq)
		i.AddStyle(end, termenv.ResetSeq)
		return
	}

	defer i.ResetStyle(end)

	style, err := GetLsColor(line[basePathIndex:end])
	if err == nil {
		slog.Debug("GetLsColor (LS_COLORS) failed", "error", err)
		i.AddStyle(basePathIndex, style)
		return
	}

	if metaErr != nil {
		i.AddStyle(basePathIndex, termenv.ANSIBrightBlack.Sequence(false))
		return
	}
	if meta.IsSymlink {
		i.AddStyle(basePathIndex, termenv.ANSIMagenta.Sequence(false))
		return
	}
	if meta.IsDirectory {
		i.AddStyle(basePathIndex, termenv.BoldSeq, termenv.ANSIBlue.Sequence(false))
		return
	}

	if line[basePathIndex] == '.' {
		i.AddStyle(basePathIndex, termenv.ANSIBrightBlack.Sequence(false))
		return
	}
	i.AddStyle(basePathIndex, termenv.ResetSeq)
}

func Colorize(line string, rules []Rule) string {
	index := make(Index)
	for _, rule := range rules {
		re := rule.Regexp
		if re == nil {
			continue
		}

		colors := strings.Split(rule.Colors, ",")

		matches := re.FindAllStringSubmatchIndex(line, -1)

		if len(matches) == 0 {
			continue
		}

		if rule.Overwrite {
			slog.Debug("Overwriting other rules for current line")
			index.Reset()
			index.Extent(line, matches, colors)
			break
		}

		index.Extent(line, matches, colors)
	}

	var buf strings.Builder
	size := len(line)
	buf.Grow(2 * size)

	for i := range size {
		if v, ok := index[i]; ok {
			buf.WriteString("\x1b[" + join(v) + "m")
		}
		buf.WriteByte(line[i])
	}

	if v, ok := index[size]; ok {
		buf.WriteString("\x1b[" + strings.Join(v, ";") + "m")
	}

	buf.WriteString("\x1b[" + termenv.ResetSeq + "m")

	return buf.String()
}

func join(s []string) string {
	f := slices.DeleteFunc(s, func(str string) bool { return str == "" })
	return strings.Join(f, ";")
}
