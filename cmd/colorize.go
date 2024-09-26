package cmd

import (
	"fmt"
	"strings"
)

func ExtentColorMapFromMatches(colorMap map[int]string, matches [][]int, colors []string) {
	for _, match := range matches {
		for i := range (len(match) - 2) / 2 {
			start := match[i*2+2]
			end := match[i*2+3]
			cfgStyle := strings.TrimSpace(colors[i%len(colors)])
			colorMap[start] = cfgStyle

			if len(colorMap[end]) > 0 {
				colorMap[end] = "reset " + colorMap[end]
			} else {
				colorMap[end] = "reset"
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

		if rule.Overwrite && len(matches) > 0 {
			if Verbose {
				fmt.Println("Overwriting other rules for current line")
			}
			colorMap = make(map[int]string)
			ExtentColorMapFromMatches(colorMap, matches, colors)
			break
		} else {
			ExtentColorMapFromMatches(colorMap, matches, colors)
		}

	}

	for i, char := range line {
		if len(colorMap[i]) > 0 {
			color := ""
			for _, style := range strings.Split(colorMap[i], " ") {
				color += Ansi.GetColor(style)
			}
			coloredLine = coloredLine + color + string(char)
		} else {
			coloredLine += string(char)
		}
	}

	return coloredLine + Ansi.Reset
}
