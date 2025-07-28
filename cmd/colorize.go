package cmd

import (
	"log/slog"
	"strings"
)

type Index map[int]string

func (i Index) Add(idx int, style string) {
	i[idx] = i[idx] + style
}

func (i Index) Reset(idx int) {
	i.Add(idx, Ansi.Reset)
}

func ExtentIndexFromMatches(index Index, matches [][]int, colors []string) {
	for _, match := range matches {
		for i := range (len(match) - 2) / 2 {
			start := match[i*2+2]
			end := match[i*2+3]

			cfgStyle := strings.TrimSpace(colors[i%len(colors)])

			var ansiStyles strings.Builder
			for style := range strings.SplitSeq(cfgStyle, " ") {
				ansiStyles.WriteString(Ansi.GetColor(style))
			}
			index.Add(start, ansiStyles.String())
			index.Reset(end)
		}
	}
}

func ExtentIndexForPath(index Index, matches [][]int, line string) {
	for _, match := range matches {
		groups := match[2:]
		for i := range len(groups) / 2 {
			start := groups[i*2]
			end := groups[i*2+1]

			if start == -1 || end == -1 {
				continue
			}

			path := line[start:end]

			slog.Debug("Path", "value", path)

			basePathIndex := start
			for i := len(path) - 1; i >= 0; i-- {
				if path[i] == '/' || path[i] == '\\' {
					basePathIndex = start + i + 1
					break
				}
			}

			index.Add(start, Ansi.Blue)

			meta, metaErr := GetFileMetadata(path)

			if metaErr == nil && meta.IsEveyone {
				index.Add(basePathIndex, Ansi.Green+Ansi.Bold)
				index.Add(end, Ansi.Reset+"!")
				return
			}

			if metaErr == nil && meta.IsExecutable {
				index.Add(basePathIndex, Ansi.Red+Ansi.Bold)
				index.Add(end, Ansi.Reset+"*")
				return
			}

			defer index.Reset(end)

			style, err := GetLsColor(line[basePathIndex:end])
			if err == nil {
				slog.Debug("GetLsColor (LS_COLORS) failed", "error", err)
				index.Add(basePathIndex, style)
				return
			}

			if metaErr != nil {
				index.Add(basePathIndex, Ansi.Gray)
				return
			}
			if meta.IsSymlink {
				index.Add(basePathIndex, Ansi.Magenta)
				return
			}
			if meta.IsDirectory {
				index.Add(basePathIndex, Ansi.Bold+Ansi.Blue)
				return
			}

			if line[basePathIndex] == '.' {
				index.Add(basePathIndex, Ansi.Gray)
				return
			}
			index.Add(basePathIndex, Ansi.Reset)
		}
	}
}

func Colorize(line string, rules []Rule) string {
	var buf strings.Builder

	colorMap := make(Index)
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
			colorMap = make(Index)
			ExtentIndexFromMatches(colorMap, matches, colors)
			break
		}

		if rule.Type == "path" {
			slog.Debug("Using Path parser")
			ExtentIndexForPath(colorMap, matches, line)
			continue
		}

		ExtentIndexFromMatches(colorMap, matches, colors)
	}

	var last int
	for i, char := range line {
		if v, ok := colorMap[i]; ok {
			buf.WriteString(v)
		}
		buf.WriteRune(char)
		last++
	}

	if v, ok := colorMap[last]; ok {
		buf.WriteString(v)
	}

	buf.WriteString(Ansi.Reset)

	return buf.String()
}
