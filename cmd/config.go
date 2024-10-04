package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

var StaticConfig string

type Config struct {
	Regexp string `toml:"regexp"`
	File   string `toml:"file"`
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
	envConfigPath := os.Getenv("COLORIZE_CONFIG")
	if len(envConfigPath) > 0 {
		configPaths = append(configPaths, envConfigPath)
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPaths = append(configPaths, filepath.Join(homeDir, ".config/colorize/config.toml"))
	} else {
		Debug("Error getting home directory:", err)
	}

	for _, configPath := range configPaths {
		if _, err := Stat(configPath); err != nil {
			Debug("Failed to loading config file:", configPath)
			Debug(err)
			continue
		}

		Debug("Loading config file:", configPath)

		var additionalConfig map[string]Config
		_, err = DecodeTomlFile(configPath, &additionalConfig)
		if err == nil {
			for key, value := range additionalConfig {
				config[key] = value
			}
			continue
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
