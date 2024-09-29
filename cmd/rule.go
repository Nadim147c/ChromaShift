package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
)

type (
	CommandRules struct {
		SkipColor SkipColor `toml:"skip-color"`
		Rules     []Rule    `toml:"rules"`
		Stderr    bool      `toml:"stderr"`
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

func LoadRules(ruleFile string, stat StatFunc, decodeFile DecodeFileFunc) (CommandRules, error) {
	var cmdRules CommandRules

	if len(RulesDirectory) > 0 {
		ruleFilePath := filepath.Join(RulesDirectory, ruleFile)
		if Verbose {
			fmt.Println("Using rules file:", ruleFilePath)
		}
		_, err := decodeFile(ruleFilePath, &cmdRules)
		if err != nil {
			if Verbose {
				fmt.Fprintf(os.Stderr, "Can't load rules from path: %s", err)
			}
		}
		return cmdRules, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		if Verbose {
			fmt.Println("Error getting home directory:", err)
		}
		homeDir = ""
	}

	rulesPaths := []string{
		os.Getenv("COLORIZE_RULES"),
		filepath.Join(homeDir, ".config/colorize/rules"),
		"/usr/local/etc/colorize/rules",
		"/etc/colorize/rules",
	}

	for _, rulesDir := range rulesPaths {
		if rulesDir == "" {
			continue
		}

		ruleFilePath := path.Join(rulesDir, ruleFile)

		if _, err := stat(ruleFilePath); err == nil {
			if Verbose {
				fmt.Println("Using rules file:", ruleFilePath)
			}

			_, err := decodeFile(ruleFilePath, &cmdRules)
			if err == nil {
				sort.Slice(cmdRules.Rules, func(i int, j int) bool {
					if cmdRules.Rules[i].Overwrite != cmdRules.Rules[j].Overwrite {
						return cmdRules.Rules[i].Overwrite
					}
					return cmdRules.Rules[i].Priority < cmdRules.Rules[j].Priority
				})

				if Verbose {
					fmt.Printf("stderr: %+v\n", cmdRules.Stderr)
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
