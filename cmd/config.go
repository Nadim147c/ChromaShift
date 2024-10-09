package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var StaticConfig string

type (
	Config struct {
		Regexp string         `toml:"regexp"`
		File   string         `toml:"file"`
		Sub    map[string]Sub `toml:"sub"`
	}

	Sub struct {
		Regexp string `toml:"regexp"`
		File   string `toml:"file"`
	}
)

func GetRuleFileNameForSubcommand(subCommands map[string]Sub, args []string) (string, error) {
	subCommandName := args[1]
	if len(subCommands[subCommandName].File) > 0 {
		return subCommands[subCommandName].File, nil
	}
	for _, values := range subCommands {
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

func GetRuleFileName(config map[string]Config, args []string) (string, error) {
	cmdName := args[0]
	cmdBaseName := filepath.Base(cmdName)
	if commandConfig, found := config[cmdBaseName]; found {
		if len(commandConfig.Sub) == 0 {
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

	for name, values := range config {
		if cmdName == name || cmdBaseName == name {
			if len(values.Sub) == 0 {
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
			if len(values.Sub) == 0 {
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

func LoadConfig() (map[string]Config, error) {
	var config map[string]Config

	Debug("Loading embeded config")

	_, err := DecodeToml(StaticConfig, &config)
	if err != nil {
		Debug("Err loading embeded config", err)
	}

	if len(ConfigFile) > 0 {
		Debug("Loading config file:", ConfigFile)
		_, err := DecodeTomlFile(ConfigFile, &config)
		if err == nil {
			return config, err
		} else {
			Debug("Failed Loading config file:", err)
		}
	}

	configPaths := []string{}
	envConfigPath := os.Getenv("CHROMASHIFT_CONFIG")
	if len(envConfigPath) > 0 {
		configPaths = append(configPaths, envConfigPath)
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPaths = append(configPaths, filepath.Join(homeDir, ".config/chromashift/config.toml"))
	} else {
		Debug("Error getting home directory:", err)
	}

	for _, configPath := range configPaths {
		file, err := os.Open(configPath)
		if err != nil {
			Debug("Failed to loading config file:", configPath)
			Debug(err)
			continue
		}
		defer file.Close()

		Debug("Loading config file:", configPath)

		content, err := io.ReadAll(file)
		if err != nil {
			Debug(err)
			continue
		}

		var additionalConfig map[string]Config
		_, err = DecodeToml(string(content), &additionalConfig)
		if err == nil {
			for key, value := range additionalConfig {
				config[key] = value
			}
		} else {
			Debug("Can't load config from path:", configPath)
			Debug(err)
		}

	}

	if len(config) > 0 {
		return config, nil
	} else {
		return config, fmt.Errorf("no config found")
	}
}
