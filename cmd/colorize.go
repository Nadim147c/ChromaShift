package cmd

import (
	"iter"
	"log/slog"
	"strings"
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

type Index map[int]string

func (i *Index) Reset() {
	*i = make(Index)
}

func (i Index) AddStyle(idx int, style string) {
	i[idx] = i[idx] + style
}

func (i Index) ResetStyle(idx int) {
	i.AddStyle(idx, Ansi.Reset)
}

func (i Index) Extent(line string, matches [][]int, colors []string) {
loop:
	for match := range RegexMatches(matches) {
		idx, start, end := match.Values()

		cfgStyle := strings.TrimSpace(colors[idx%len(colors)])

		var ansiStyles strings.Builder
		for style := range strings.SplitSeq(cfgStyle, " ") {
			style = strings.ToLower(strings.TrimSpace(style))
			if style == "path" {
				i.ExtentPath(line, start, end)
				continue loop
			}

			ansiStyles.WriteString(Ansi.GetColor(style))
		}

		i.AddStyle(start, ansiStyles.String())
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

	i.AddStyle(start, Ansi.Blue)

	meta, metaErr := GetFileMetadata(path)

	if metaErr == nil && meta.IsEveyone {
		i.AddStyle(basePathIndex, Ansi.Green+Ansi.Bold)
		i.AddStyle(end, Ansi.Reset+"!")
		return
	}

	if metaErr == nil && meta.IsExecutable {
		i.AddStyle(basePathIndex, Ansi.Red+Ansi.Bold)
		i.AddStyle(end, Ansi.Reset+"*")
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
		i.AddStyle(basePathIndex, Ansi.Gray)
		return
	}
	if meta.IsSymlink {
		i.AddStyle(basePathIndex, Ansi.Magenta)
		return
	}
	if meta.IsDirectory {
		i.AddStyle(basePathIndex, Ansi.Bold+Ansi.Blue)
		return
	}

	if line[basePathIndex] == '.' {
		i.AddStyle(basePathIndex, Ansi.Gray)
		return
	}
	i.AddStyle(basePathIndex, Ansi.Reset)
}

func Colorize(line string, rules []Rule) string {
	var buf strings.Builder

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

	var last int
	for i, char := range line {
		if v, ok := index[i]; ok {
			buf.WriteString(v)
		}
		buf.WriteRune(char)
		last++
	}

	if v, ok := index[last]; ok {
		buf.WriteString(v)
	}

	buf.WriteString(Ansi.Reset)

	return buf.String()
}
