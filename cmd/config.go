package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Regexp string `toml:"regexp"`
	File   string `toml:"file"`
}

func LoadConfig() (map[string]Config, error) {
	var config map[string]Config

	if len(ConfigFile) > 0 {
		if Verbose {
			fmt.Println("Using config file:", ConfigFile)
		}
		_, err := DecodeTomlFile(ConfigFile, &config)
		return config, err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		if Verbose {
			fmt.Println("Error getting home directory:", err)
		}
		homeDir = ""
	}

	configPaths := []string{
		os.Getenv("COLORIZE_CONFIG"),
		filepath.Join(homeDir, ".config/colorize/config.toml"),
		"/usr/local/etc/colorize/config.toml",
		"/etc/colorize/config.toml",
	}

	for _, path := range configPaths {
		if path == "" {
			continue
		}

		if _, err := Stat(path); err == nil {
			ConfigFile = path

			if Verbose {
				fmt.Println("Using config file:", ConfigFile)
			}

			_, err = DecodeTomlFile(ConfigFile, &config)
			if err == nil {
				return config, nil
			}

			if Verbose {
				fmt.Fprintf(os.Stderr, "Can't load config from path: %s", err)
			}

			break
		}
	}

	return nil, fmt.Errorf("No config found.")
}
