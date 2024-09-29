package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

//go:embed LS_COLORS.txt
var defaultLsColors string

type LsColor struct {
	Glob glob.Glob
	Code string
}

var lsColorsMap []LsColor

func GetLsColor(line string) string {
	lsColors := os.Getenv("LS_COLORS")

	if len(lsColors) <= 0 {
		lsColors = defaultLsColors
	}

	if len(lsColorsMap) == 0 {
		entries := strings.Split(lsColors, ":")
		for _, entry := range entries {
			parts := strings.Split(entry, "=")
			if len(parts) != 2 {
				continue
			}
			pattern := parts[0]
			colorCode := parts[1]

			g, err := glob.Compile(pattern)
			if err != nil {
				if Verbose {
					fmt.Fprintln(os.Stderr, "Failed compiling glob", pattern)
				}
				continue
			}
			lsColorsMap = append(lsColorsMap, LsColor{Glob: g, Code: colorCode})
		}
	}

	for _, lsColor := range lsColorsMap {
		fileName := filepath.Base(line)
		if lsColor.Glob.Match(fileName) {
			return fmt.Sprintf("\033[%sm", lsColor.Code)
		}
	}
	return Ansi.Blue
}
