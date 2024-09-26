package cmd_test

import (
	"colorize/cmd"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
)

func mockStatForConfig(path string) (os.FileInfo, error) {
	if path == "/fake/home/.config/colorize/config.toml" {
		return nil, nil
	}
	return nil, os.ErrNotExist
}

func mockDecodeFileForConfig(file string, v interface{}) (toml.MetaData, error) {
	configMap, ok := v.(*map[string]cmd.Config)
	if !ok {
		return toml.MetaData{}, fmt.Errorf("invalid type for config map")
	}
	if *configMap == nil {
		*configMap = make(map[string]cmd.Config)
	}
	(*configMap)["test"] = cmd.Config{Regexp: "test-regexp", File: "test-file"}
	return toml.MetaData{}, nil
}

func TestLoadConfig(t *testing.T) {
	verbose := cmd.Verbose
	cmd.Verbose = true
	defer func() { cmd.Verbose = verbose }()

	// Test case 1: Config file specified explicitly
	t.Run("Config file specified explicitly", func(t *testing.T) {
		cmd.ConfigFile = "/fake/path/to/config.toml"

		result, err := cmd.LoadConfig(mockStatForConfig, mockDecodeFileForConfig)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		expectedConfig := map[string]cmd.Config{
			"test": {Regexp: "test-regexp", File: "test-file"},
		}

		if !reflect.DeepEqual(result, expectedConfig) {
			t.Fatalf("Expected config %+v, but got %+v", expectedConfig, result)
		}
	})

	// Test case 2: Config file found in default paths
	t.Run("Config file found in default paths", func(t *testing.T) {
		cmd.ConfigFile = ""

		homeDir := "/fake/home"
		os.Setenv("HOME", homeDir)

		result, err := cmd.LoadConfig(mockStatForConfig, mockDecodeFileForConfig)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		expectedConfig := map[string]cmd.Config{
			"test": {Regexp: "test-regexp", File: "test-file"},
		}

		if !reflect.DeepEqual(result, expectedConfig) {
			t.Fatalf("Expected config %+v, but got %+v", expectedConfig, result)
		}
	})

	// Test case 3: No config file found
	t.Run("No config file found", func(t *testing.T) {
		cmd.ConfigFile = "" // No explicit config file

		// Mock os.Stat to simulate no config files found
		noConfigMockStat := func(path string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}

		_, err := cmd.LoadConfig(noConfigMockStat, mockDecodeFileForConfig)
		if err == nil || err.Error() != "No config found." {
			t.Fatalf("Expected 'No config found.' error, but got: %v", err)
		}
	})
}
