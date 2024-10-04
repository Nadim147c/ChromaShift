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

	if Verbose {
		fmt.Println("Loading embeded config")
	}
	_, err := DecodeToml(StaticConfig, &config)
	if err != nil {
		if Verbose {
			fmt.Println("Err loading embeded config", err)
		}
	}

	if len(ConfigFile) > 0 {
		if Verbose {
			fmt.Println("Loading config file:", ConfigFile)
		}
		_, err := DecodeTomlFile(ConfigFile, &config)
		if err == nil {
			return config, err
		} else {
			if Verbose {
				fmt.Println("Failed Loading config file:", err)
			}
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
		if Verbose {
			fmt.Println("Error getting home directory:", err)
		}
	}

	for _, configPath := range configPaths {
		if _, err := Stat(configPath); err != nil {
			if Verbose {
				fmt.Println("Failed to loading config file:", configPath)
				fmt.Println(err)
			}
			continue
		}

		if Verbose {
			fmt.Println("Loading config file:", configPath)
		}

		var additionalConfig map[string]Config
		_, err = DecodeTomlFile(configPath, &additionalConfig)
		if err == nil {
			for key, value := range additionalConfig {
				config[key] = value
			}
			continue
		} else {
			if Verbose {
				fmt.Println("Can't load config from path:", configPath)
				fmt.Println(err)
			}
		}

	}

	if len(config) > 0 {
		return config, nil
	} else {
		return config, fmt.Errorf("no config found")
	}
}
