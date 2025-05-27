package cmd

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

var StaticConfig string

type (
	Config map[string]Command

	Command struct {
		Regexp string       `toml:"regexp"`
		File   string       `toml:"file"`
		Sub    *SubCommands `toml:"sub"`
	}

	SubCommands map[string]SubCommand

	SubCommand struct {
		Regexp string `toml:"regexp"`
		File   string `toml:"file"`
	}
)

func GetRuleFileNameForSubcommand(subCommands *SubCommands, args []string) (string, error) {
	subCommandName := args[1]

	if (*subCommands)[subCommandName].File != "" {
		return (*subCommands)[subCommandName].File, nil
	}

	for _, values := range *subCommands {
		commandStr := strings.Join(args, " ")
		if values.Regexp == "" {
			continue
		}
		if matched, _ := regexp.Match(values.Regexp, []byte(commandStr)); matched {
			return values.File, nil
		}
	}
	return "", fmt.Errorf("No matching subcommand")
}

func GetRuleFileName(config *Config, args []string) (string, error) {
	cmdName := args[0]
	cmdBaseName := filepath.Base(cmdName)
	if commandConfig, found := (*config)[cmdBaseName]; found {
		if commandConfig.Sub == nil {
			return commandConfig.File, nil
		}

		Debug("Loading sub commands for", cmdBaseName)
		ruleFileName, err := GetRuleFileNameForSubcommand(commandConfig.Sub, args)
		if err == nil {
			return ruleFileName, nil
		} else {
			Debug(err)
		}
	}

	for name, values := range *config {
		if cmdName == name || cmdBaseName == name {
			if values.Sub == nil {
				return values.File, nil
			}

			Debug("Loading sub commands for", name)
			ruleFileName, err := GetRuleFileNameForSubcommand(values.Sub, args)
			if err == nil {
				return ruleFileName, nil
			} else {
				Debug(err)
			}
		}

		Debug("Regex", values.Regexp)

		if values.Regexp == "" {
			continue
		}

		commandStr := strings.Join(args, " ")
		if matched, _ := regexp.Match(values.Regexp, []byte(commandStr)); matched {
			if values.Sub == nil {
				return values.File, nil
			}

			Debug("Loading sub commands for", name)
			ruleFileName, err := GetRuleFileNameForSubcommand(values.Sub, args)
			if err == nil {
				return ruleFileName, nil
			} else {
				Debug(err)
			}
		}
	}

	return "", fmt.Errorf("No matching command")
}

func LoadConfig() (*Config, error) {
	var config Config

	Debug("Loading embedded config")

	_, err := toml.Decode(StaticConfig, &config)
	if err != nil {
		Debug("Err loading embedded config", err)
	}

	if len(ConfigFile) > 0 {
		Debug("Loading config file:", ConfigFile)
		_, err := toml.DecodeFile(ConfigFile, &config)
		if err == nil {
			return nil, err
		} else {
			Debug("Failed Loading config file:", err)
		}
	}

	configPaths := []string{}
	if path := os.Getenv("CHROMASHIFT_CONFIG"); path != "" {
		configPaths = append(configPaths, path)
	}

	cfgDir, err := os.UserConfigDir()
	if err != nil {
		Debug("Error getting config directory:", err)
	} else {
		path := filepath.Join(cfgDir, "ChromaShift", "config.toml")
		configPaths = append(configPaths, path)
	}

	for path := range slices.Values(configPaths) {
		loadConfigFile(path, config)
	}

	if len(config) == 0 {
		return nil, fmt.Errorf("no config found")
	}

	return &config, nil
}

func loadConfigFile(path string, config Config) {
	file, err := os.Open(path)
	if err != nil {
		Debug("Failed to loading config file:", path)
		Debug(err)
		return
	}
	defer file.Close()

	Debug("Loading config file:", path)

	var additionalConfig Config
	_, err = toml.NewDecoder(file).Decode(&additionalConfig)
	if err != nil {
		Debug("Can't load config from path:", path)
		Debug(err)
		return
	}

	maps.Copy(config, additionalConfig)
}
