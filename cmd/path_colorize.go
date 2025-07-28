package cmd

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gobwas/glob"
)

//go:embed LS_COLORS.txt
var DefaultLsColors string

type LsColor struct {
	Glob glob.Glob
	Code string
}

var LsColorsMap []LsColor

func GetLsColor(line string) (string, error) {
	lsColors := os.Getenv("LS_COLORS")

	if len(lsColors) <= 0 {
		lsColors = DefaultLsColors
	}

	if len(LsColorsMap) == 0 {
		entries := strings.SplitSeq(lsColors, ":")
		for entry := range entries {
			parts := strings.Split(entry, "=")
			if len(parts) != 2 {
				continue
			}
			pattern := parts[0]
			colorCode := parts[1]

			g, err := glob.Compile(pattern)
			if err != nil {
				slog.Debug("Failed compiling glob", "pattern", pattern, "error", err)
				continue
			}
			LsColorsMap = append(LsColorsMap, LsColor{Glob: g, Code: colorCode})
		}
	}

	for _, lsColor := range LsColorsMap {
		fileName := filepath.Base(line)
		if lsColor.Glob.Match(fileName) {
			return fmt.Sprintf("\033[%sm", lsColor.Code), nil
		}
	}

	return "", fmt.Errorf("File color doesn't exists")
}

func GetColorForMode(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}

	stat := info.Sys().(*syscall.Stat_t)
	perms := stat.Mode & 0777

	mode := info.Mode()

	if mode.IsDir() {
		return Ansi.Blue, nil // Directory
	}
	if mode&0111 != 0 {
		return "*" + Ansi.Red, nil // Other executable files
	}
	if mode&os.ModeSymlink != 0 {
		return Ansi.Cyan, nil // Symlink
	}
	if perms == 0777 {
		return Ansi.Bold + Ansi.Green, nil
	}
	if mode.IsRegular() {
		return Ansi.White, nil // Regular file
	}
	return "", fmt.Errorf("Failed to find color from mode")
}
