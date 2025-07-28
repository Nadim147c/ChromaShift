package cmd

import (
	"embed"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/BurntSushi/toml"
)

var StaticRulesDirectory embed.FS

type (
	CommandRules struct {
		Rules  []Rule `toml:"rules"`
		Stderr bool   `toml:"stderr"`
		PTY    bool   `toml:"pty"`
	}

	Rule struct {
		Regexp    *regexp.Regexp `toml:"regexp"`
		Colors    string         `toml:"colors"`
		Overwrite bool           `toml:"overwrite"`
		Priority  int            `toml:"priority"`
		Type      string         `toml:"type"`
	}
)

func SortRules(rules []Rule) {
	sort.Slice(rules, func(i int, j int) bool {
		if rules[i].Overwrite != rules[j].Overwrite {
			return rules[i].Overwrite
		}
		return rules[i].Priority < rules[j].Priority
	})
}

func LoadRules(ruleFile string) (*CommandRules, error) {
	var cmdRules CommandRules

	if len(RulesDirectory) > 0 {
		ruleFilePath := filepath.Join(RulesDirectory, ruleFile)
		slog.Debug("Loading rules file", "path", ruleFilePath)

		_, err := toml.DecodeFile(ruleFilePath, &cmdRules)
		if err == nil {
			SortRules(cmdRules.Rules)
			return &cmdRules, err
		} else {
			slog.Debug("Failed decoding toml file", "error", err)
		}
	}

	rulesPaths := []string{}
	envRulesDir := os.Getenv("CHROMASHIFT_RULES")

	if len(envRulesDir) > 0 {
		rulesPaths = append(rulesPaths, envRulesDir)
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		rulesPaths = append(rulesPaths, filepath.Join(homeDir, ".config/ChromaShift/rules"))
	} else {
		slog.Debug("Error getting home directory", "error", err)
	}

	for _, rulesDir := range rulesPaths {
		ruleFilePath := filepath.Join(rulesDir, ruleFile)

		file, err := os.Open(ruleFilePath)
		if err != nil {
			slog.Debug("Failed to load rules file", "path", ruleFilePath, "error", err)
			continue
		}
		defer file.Close()

		slog.Debug("Loading rules file", "path", ruleFilePath)

		content, err := io.ReadAll(file)
		if err != nil {
			slog.Debug("Error reading rules file", "error", err)
			continue
		}

		_, err = toml.Decode(string(content), &cmdRules)
		if err != nil {
			slog.Debug("Error decoding toml", "error", err)
			continue
		}

		SortRules(cmdRules.Rules)
		return &cmdRules, nil
	}

	ruleFilePath := filepath.Join("rules", ruleFile)

	slog.Debug("Loading rules from embedded rules", "path", ruleFilePath)

	fileContentBytes, err := StaticRulesDirectory.ReadFile(ruleFilePath)
	if err == nil {
		_, err := toml.Decode(string(fileContentBytes), &cmdRules)
		if err != nil {
			return nil, err
		}

		SortRules(cmdRules.Rules)
		return &cmdRules, err
	}

	return nil, fmt.Errorf("No rules found")
}
