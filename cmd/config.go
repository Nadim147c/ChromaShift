package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Regexp string `toml:"regexp"`
	File   string `toml:"file"`
}

func loadConfig() (map[string]Config, error) {
	var config map[string]Config

	homeDir, err := os.UserHomeDir()
	if err != nil {
		if verbose {
			fmt.Println("Error getting home directory:", err)
		}
		homeDir = ""
	}

	if len(cfgFile) != 0 {
		_, err = toml.DecodeFile(cfgFile, &config)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Can't load config from path: %s", err)
			}
		}
	}

	configPaths := []string{
		os.Getenv("COLORIZE_CONFIG"),
		filepath.Join(homeDir, ".config/colorize/config.toml"),
		"/etc/colorize/config.toml",
	}

	for _, path := range configPaths {
		if path == "" {
			continue
		}

		if _, err := os.Stat(path); err == nil {
			cfgFile = path

			if verbose {
				fmt.Println("Using config file:", cfgFile)
			}

			_, err = toml.DecodeFile(cfgFile, &config)
			if err == nil {
				return config, nil
			}

			if verbose {
				fmt.Fprintf(os.Stderr, "Can't load config from path: %s", err)
			}

			break
		}
	}

	return nil, fmt.Errorf("No config found.")
}
