package cmd

import (
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/xdg"
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
			parts := strings.SplitN(entry, "=", 2)
			if len(parts) != 2 {
				continue
			}
			pattern := parts[0]
			colorCode := parts[1]

			g, err := glob.Compile(pattern)
			if err != nil {
				slog.Debug(
					"Failed compiling glob",
					"pattern",
					pattern,
					"error",
					err,
				)
				continue
			}
			LsColorsMap = append(LsColorsMap, LsColor{Glob: g, Code: colorCode})
		}
	}

	for _, lsColor := range LsColorsMap {
		fileName := filepath.Base(line)
		if lsColor.Glob.Match(fileName) {
			return lsColor.Code, nil
		}
	}

	return "", fmt.Errorf("File color doesn't exists")
}

type PathKind int

const (
	PathSymlink = iota
	PathDirectory
	PathRegular
	PathSpecial
)

type FileMetadata struct {
	IsDirectory  bool
	IsSymlink    bool
	IsExecutable bool
	IsEveyone    bool
	Kind         PathKind
}

func GetFileMetadata(path string) (*FileMetadata, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	path, err = FindPath(cwd, path)
	if err != nil {
		return nil, err
	}

	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	mode := info.Mode()
	metadata := &FileMetadata{}

	if mode&os.ModeSymlink != 0 {
		metadata.IsSymlink = true
		metadata.IsExecutable = mode&0o111 != 0
		metadata.Kind = PathSymlink
	} else if mode.IsDir() {
		metadata.IsDirectory = true
		metadata.Kind = PathDirectory
	} else if mode.IsRegular() {
		metadata.Kind = PathDirectory
		metadata.IsExecutable = mode&0o111 != 0
		metadata.IsEveyone = mode.Perm()&0o777 != 0
	} else {
		metadata.Kind = PathSpecial
	}

	return metadata, nil
}

var separator = regexp.MustCompile(`[/\\]+`)

// FindPath expands environment-aware variables and returns an absolute path
// from a given base path (not CWD).
// Path Prefix Expands:
//   - $HOME or ~      : xdg.Home
//   - $XDG_CONFIG_HOME: xdg.ConfigHome
//   - $XDG_CACHE_HOME : xdg.CacheHome
//   - $XDG_DATA_HOME  : xdg.DataHome
func FindPath(basePath, inputPath string) (string, error) {
	if inputPath == "" {
		return "", errors.New("empty path")
	}

	if filepath.IsAbs(inputPath) {
		return filepath.Clean(inputPath), nil
	}

	split := separator.Split(inputPath, 2)
	if len(split) != 2 {
		joined := filepath.Join(basePath, inputPath)
		return filepath.Clean(joined), nil
	}

	path := filepath.Join(basePath, inputPath)

	base, rest := split[0], split[1]
	switch base {
	case "~", "$HOME":
		path = filepath.Join(xdg.Home, rest)
	case "$XDG_CONFIG_HOME":
		path = filepath.Join(xdg.ConfigHome, rest)
	case "$XDG_CACHE_HOME":
		path = filepath.Join(xdg.CacheHome, rest)
	case "$XDG_DATA_HOME":
		path = filepath.Join(xdg.DataHome, rest)
	}

	return filepath.Clean(path), nil
}
