package cmd

import "strings"

func ExtentColorMapFromMatches(colorMap map[int]string, matches [][]int, colors []string) {
	for _, match := range matches {
		for i := range (len(match) - 2) / 2 {
			start := match[i*2+2]
			end := match[i*2+3]

			cfgStyle := strings.TrimSpace(colors[i%len(colors)])
			ansiStyles := ""
			for _, style := range strings.Split(cfgStyle, " ") {
				ansiStyles += Ansi.GetColor(style)
			}
			colorMap[start] = ansiStyles

			if len(colorMap[end]) > 0 {
				colorMap[end] = Ansi.Reset + colorMap[end]
			} else {
				colorMap[end] = Ansi.Reset
			}

		}
	}
}

func ExtentColorMapWithLsColors(colorMap map[int]string, matches [][]int, currentLine string) {
	for _, match := range matches {
		for i := range (len(match) - 2) / 2 {
			start := match[i*2+2]
			end := match[i*2+3]
			path := currentLine[start:end]

			basePathIndex := start
			for i := len(path) - 1; i >= 0; i-- {
				if path[i] == '/' || path[i] == '\\' {
					basePathIndex = start + i + 1
					break
				}
			}

			colorMap[start] = Ansi.Blue

			cfgStyle := GetLsColor(currentLine[basePathIndex:end])
			colorMap[basePathIndex] = cfgStyle

			if len(colorMap[end]) > 0 {
				colorMap[end] = Ansi.Reset + colorMap[end]
			} else {
				colorMap[end] = Ansi.Reset
			}
		}
	}
}

func ColorizeLine(line string, rules []Rule) string {
	coloredLine := ""

	colorMap := make(map[int]string)

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
			Debug("Overwriting other rules for current line")
			colorMap = make(map[int]string)
			ExtentColorMapFromMatches(colorMap, matches, colors)
			break
		}

		if rule.Type == "path" {
			Debug("Use LS_COLORS parser")
			ExtentColorMapWithLsColors(colorMap, matches, line)
			continue
		}

		ExtentColorMapFromMatches(colorMap, matches, colors)
	}

	for i, char := range line {
		if len(colorMap[i]) > 0 {
			color := colorMap[i]
			coloredLine = coloredLine + color + string(char)
		} else {
			coloredLine += string(char)
		}
	}

	return coloredLine + Ansi.Reset
}
