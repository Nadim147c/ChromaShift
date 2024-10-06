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
var DefaultLsColors string

type LsColor struct {
	Glob glob.Glob
	Code string
}

var LsColorsMap []LsColor

func GetLsColor(line string) string {
	lsColors := os.Getenv("LS_COLORS")

	if len(lsColors) <= 0 {
		lsColors = DefaultLsColors
	}

	if len(LsColorsMap) == 0 {
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
				Debug("Failed compiling glob", pattern)
				continue
			}
			LsColorsMap = append(LsColorsMap, LsColor{Glob: g, Code: colorCode})
		}
	}

	for _, lsColor := range LsColorsMap {
		fileName := filepath.Base(line)
		if lsColor.Glob.Match(fileName) {
			return fmt.Sprintf("\033[%sm", lsColor.Code)
		}
	}

	return Ansi.Blue
}
