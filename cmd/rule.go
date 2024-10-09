package cmd

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/BurntSushi/toml"
)

var StaticRulesDirectory embed.FS

type (
	CommandRules struct {
		SkipColor SkipColor `toml:"skip-color"`
		Rules     []Rule    `toml:"rules"`
		Stderr    bool      `toml:"stderr"`
		PTY       bool      `toml:"pty"`
	}

	SkipColor struct {
		Argument  string `toml:"argument"`
		Arguments string `toml:"arguments"`
	}

	Rule struct {
		Regexp    *regexp.Regexp `toml:"regexp"`
		Colors    string         `toml:"colors"`
		Overwrite bool           `toml:"overwrite"`
		Priority  int            `toml:"priority"`
		Type      string         `toml:"type"`
	}
)

func SortRules(cmdRules *CommandRules) {
	sort.Slice(cmdRules.Rules, func(i int, j int) bool {
		if cmdRules.Rules[i].Overwrite != cmdRules.Rules[j].Overwrite {
			return cmdRules.Rules[i].Overwrite
		}
		return cmdRules.Rules[i].Priority < cmdRules.Rules[j].Priority
	})
}

func LoadRules(ruleFile string) (CommandRules, error) {
	var cmdRules CommandRules

	if len(RulesDirectory) > 0 {
		ruleFilePath := filepath.Join(RulesDirectory, ruleFile)
		Debug("Loading rules file:", ruleFilePath)

		_, err := toml.DecodeFile(ruleFilePath, &cmdRules)
		if err == nil {
			SortRules(&cmdRules)
			return cmdRules, err
		} else {
			Debug("Failed decoding toml file:", err)
		}

	}

	rulesPaths := []string{}
	envRulesDir := os.Getenv("CHROMASHIFT_RULES")

	if len(envRulesDir) > 0 {
		rulesPaths = append(rulesPaths, envRulesDir)
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		rulesPaths = append(rulesPaths, filepath.Join(homeDir, ".config/chromashift/rules"))
	} else {
		Debug("Error getting home directory:", err)
	}

	for _, rulesDir := range rulesPaths {
		ruleFilePath := path.Join(rulesDir, ruleFile)
		file, err := os.Open(ruleFilePath)
		if err != nil {
			Debug("Failed to load rules file:", ruleFilePath)
			continue
		}
		defer file.Close()

		Debug("Loading rules file:", ruleFilePath)

		content, err := io.ReadAll(file)
		if err != nil {
			Debug(err)
			continue

		}
		_, err = toml.Decode(string(content), &cmdRules)
		if err != nil {
			Debug("Error decoding toml", err)
			continue
		}

		SortRules(&cmdRules)

		return cmdRules, nil

	}

	ruleFilePath := filepath.Join("rules", ruleFile)

	Debug("Loading rules from embed rules:", ruleFilePath)

	fileContentBytes, err := StaticRulesDirectory.ReadFile(ruleFilePath)
	if err == nil {
		_, err := toml.Decode(string(fileContentBytes), &cmdRules)
		SortRules(&cmdRules)
		return cmdRules, err
	}

	return cmdRules, fmt.Errorf("No rules found.")
}
