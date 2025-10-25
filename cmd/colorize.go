package cmd

import (
	"cshift/pkgs/ansi"
	"log/slog"
	"slices"
	"strings"
)

type Matches []ansi.Match

type Match struct {
	matches []int
	styles  []string
}

func (m *Match) Len() int {
	return len(m.matches) / 2
}

func (m *Match) Match(i int) (start, end int, sequence string) {
	idx := i * 2
	start, end = m.matches[idx], m.matches[idx+1]
	if len(m.styles) <= i {
		return
	}
	seqs := []string{}
	for style := range strings.FieldsSeq(m.styles[i]) {
		style = strings.ToLower(strings.TrimSpace(style))
		if style == "path" {
			// TODO: implement path coloring
			panic("path coloring has not been implement")
		}

		seqs = append(seqs, GetColorCode(style))
	}

	sequence = strings.Join(seqs, ";")
	return
}

// func (i Index) Extent(line string, matches [][]int, colors []string) {
// loop:
// 	for match := range RegexMatches(matches) {
// 		idx, start, end := match.Values()
//
// 		cfgStyle := strings.TrimSpace(colors[idx%len(colors)])
//
// 		for style := range strings.SplitSeq(cfgStyle, " ") {
// 			style = strings.ToLower(strings.TrimSpace(style))
// 			if style == "path" {
// 				i.ExtentPath(line, start, end)
// 				continue loop
// 			}
//
// 			seq := GetColorCode(style)
// 			i.AddStyle(start, seq)
// 		}
//
// 		i.ResetStyle(end)
// 	}
// }
//
// func (i Index) ExtentPath(line string, start, end int) {
// 	path := line[start:end]
//
// 	slog.Debug("Path", "value", path)
//
// 	basePathIndex := start
// 	for i := len(path) - 1; i >= 0; i-- {
// 		if path[i] == '/' || path[i] == '\\' {
// 			basePathIndex = start + i + 1
// 			break
// 		}
// 	}
//
// 	i.AddStyle(start, termenv.ANSIBlue.Sequence(false))
//
// 	meta, metaErr := GetFileMetadata(path)
//
// 	if metaErr == nil && meta.IsEveyone {
// 		i.AddStyle(
// 			basePathIndex,
// 			termenv.ANSIGreen.Sequence(false),
// 			termenv.BoldSeq,
// 		)
// 		i.AddStyle(end, termenv.ResetSeq)
// 		return
// 	}
//
// 	if metaErr == nil && meta.IsExecutable {
// 		i.AddStyle(
// 			basePathIndex,
// 			termenv.ANSIRed.Sequence(false),
// 			termenv.BoldSeq,
// 		)
// 		i.AddStyle(end, termenv.ResetSeq)
// 		return
// 	}
//
// 	defer i.ResetStyle(end)
//
// 	style, err := GetLsColor(line[basePathIndex:end])
// 	if err == nil {
// 		slog.Debug("GetLsColor (LS_COLORS) failed", "error", err)
// 		i.AddStyle(basePathIndex, style)
// 		return
// 	}
//
// 	if metaErr != nil {
// 		i.AddStyle(basePathIndex, termenv.ANSIBrightBlack.Sequence(false))
// 		return
// 	}
// 	if meta.IsSymlink {
// 		i.AddStyle(basePathIndex, termenv.ANSIMagenta.Sequence(false))
// 		return
// 	}
// 	if meta.IsDirectory {
// 		i.AddStyle(
// 			basePathIndex,
// 			termenv.BoldSeq,
// 			termenv.ANSIBlue.Sequence(false),
// 		)
// 		return
// 	}
//
// 	if line[basePathIndex] == '.' {
// 		i.AddStyle(basePathIndex, termenv.ANSIBrightBlack.Sequence(false))
// 		return
// 	}
// 	i.AddStyle(basePathIndex, termenv.ResetSeq)
// }

func Colorize(line string, rules []Rule) string {
	matchList := make(Matches, 0)
	for _, rule := range rules {
		re := rule.Regexp
		if re == nil {
			continue
		}

		styles := strings.Split(rule.Colors, ",")

		matches := re.FindAllStringSubmatchIndex(line, -1)

		if len(matches) == 0 {
			continue
		}

		if rule.Overwrite {
			slog.Debug("Overwriting other rules for current line")
			matchList = nil
			for m := range slices.Values(matches) {
				matchList = append(matchList, &Match{m, styles})
			}
			break
		}

		for m := range slices.Values(matches) {
			matchList = append(matchList, &Match{m, styles})
		}
	}

	return ansi.Colorize(line, matchList)
}
