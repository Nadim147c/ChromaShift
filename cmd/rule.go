package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/BurntSushi/toml"
)

type (
	CommandRules struct {
		SkipColor SkipColor `toml:"skip-color"`
		Rules     []Rule    `toml:"rules"`
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
	}
)

func loadRules(ruleFile string) (CommandRules, error) {
	var cmdRules CommandRules

	homeDir, err := os.UserHomeDir()
	if err != nil {
		if verbose {
			fmt.Println("Error getting home directory:", err)
		}
		homeDir = ""
	}

	rulesPaths := []string{
		os.Getenv("COLORIZE_RULES"),
		filepath.Join(homeDir, ".config/colorize/rules"),
		"/etc/colorize/rules",
	}

	for _, rulesDir := range rulesPaths {
		if rulesDir == "" {
			continue
		}

		ruleFilePath := path.Join(rulesDir, ruleFile)

		if _, err := os.Stat(ruleFilePath); err == nil {
			if verbose {
				fmt.Println("Using rules file:", ruleFilePath)
			}

			_, err := toml.DecodeFile(ruleFilePath, &cmdRules)
			if err == nil {
				sort.Slice(cmdRules.Rules, func(i int, j int) bool {
					if cmdRules.Rules[i].Overwrite != cmdRules.Rules[j].Overwrite {
						return cmdRules.Rules[i].Overwrite
					}
					return cmdRules.Rules[i].Priority < cmdRules.Rules[j].Priority
				})

				if verbose {
					fmt.Printf("SkipColor: %+v\n", cmdRules.SkipColor)

					for _, v := range cmdRules.Rules {
						fmt.Printf("rule: %+v\n", v)
					}
				}

				return cmdRules, nil
			} else {
				fmt.Fprintln(os.Stderr, "Can't load rules from path:", err)
			}

		}
	}

	return cmdRules, fmt.Errorf("No rules found.")
}
