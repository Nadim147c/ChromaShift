package cmd_test

import (
	"colorize/cmd"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/BurntSushi/toml"
)

func mockStatForRules(path string) (os.FileInfo, error) {
	if path == "/etc/colorize/rules/test.rules" {
		return nil, nil
	}
	return nil, os.ErrNotExist
}

func mockDecodeFileForRules(file string, v interface{}) (toml.MetaData, error) {
	if file == "/etc/colorize/rules/test.rules" {
		cmdRules := v.(*cmd.CommandRules)
		cmdRules.SkipColor = cmd.SkipColor{
			Argument:  "skip-arg",
			Arguments: "skip-args",
		}
		cmdRules.Rules = []cmd.Rule{
			{Regexp: regexp.MustCompile("test1"), Colors: "red", Overwrite: false, Priority: 0},
			{Regexp: regexp.MustCompile("test2"), Colors: "green", Overwrite: false, Priority: 0},
		}
		return toml.MetaData{}, nil
	}
	return toml.MetaData{}, fmt.Errorf("error decoding file")
}

func TestLoadRules(t *testing.T) {
	verbose := cmd.Verbose
	cmd.Verbose = true
	defer func() { cmd.Verbose = verbose }()

	t.Run("Rules file found and loaded from specified directory", func(t *testing.T) {
		cmd.RulesDirectory = "/fake/rules/directory"
		defer func() { cmd.RulesDirectory = "" }()
		ruleFile := "test.rules"

		mockStatForRulesWithSpecificDir := func(path string) (os.FileInfo, error) {
			if path == "/fake/rules/directory/test.rules" {
				return nil, nil
			}
			return nil, os.ErrNotExist
		}

		mockDecodeFileForRulesWithSpecificDir := func(file string, v interface{}) (toml.MetaData, error) {
			if file == "/fake/rules/directory/test.rules" {
				cmdRules := v.(*cmd.CommandRules)
				cmdRules.SkipColor = cmd.SkipColor{
					Argument:  "skip-arg",
					Arguments: "skip-args",
				}
				cmdRules.Rules = []cmd.Rule{
					{Regexp: regexp.MustCompile("test1"), Colors: "red", Overwrite: false, Priority: 0},
					{Regexp: regexp.MustCompile("test2"), Colors: "green", Overwrite: false, Priority: 0},
				}
				return toml.MetaData{}, nil
			}
			return toml.MetaData{}, fmt.Errorf("error decoding file")
		}

		result, err := cmd.LoadRules(ruleFile, mockStatForRulesWithSpecificDir, mockDecodeFileForRulesWithSpecificDir)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		expectedSkipColor := cmd.SkipColor{
			Argument:  "skip-arg",
			Arguments: "skip-args",
		}
		expectedRules := []cmd.Rule{
			{Regexp: regexp.MustCompile("test1"), Colors: "red", Overwrite: false, Priority: 0},
			{Regexp: regexp.MustCompile("test2"), Colors: "green", Overwrite: false, Priority: 0},
		}

		if !reflect.DeepEqual(result.SkipColor, expectedSkipColor) {
			t.Fatalf("Expected skip color %+v, but got %+v", expectedSkipColor, result.SkipColor)
		}

		if !reflect.DeepEqual(result.Rules, expectedRules) {
			t.Fatalf("Expected rules %+v, but got %+v", expectedRules, result.Rules)
		}
	})

	t.Run("Rules file found and loaded", func(t *testing.T) {
		ruleFile := "test.rules"
		result, err := cmd.LoadRules(ruleFile, mockStatForRules, mockDecodeFileForRules)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		expectedSkipColor := cmd.SkipColor{
			Argument:  "skip-arg",
			Arguments: "skip-args",
		}
		expectedRules := []cmd.Rule{
			{Regexp: regexp.MustCompile("test1"), Colors: "red", Overwrite: false, Priority: 0},
			{Regexp: regexp.MustCompile("test2"), Colors: "green", Overwrite: false, Priority: 0},
		}

		if !reflect.DeepEqual(result.SkipColor, expectedSkipColor) {
			t.Fatalf("Expected skip color %+v, but got %+v", expectedSkipColor, result.SkipColor)
		}

		if !reflect.DeepEqual(result.Rules, expectedRules) {
			t.Fatalf("Expected rules %+v, but got %+v", expectedRules, result.Rules)
		}
	})

	t.Run("No rules file found", func(t *testing.T) {
		ruleFile := "nonexistent.rules"
		_, err := cmd.LoadRules(ruleFile, mockStatForRules, mockDecodeFileForRules)
		if err == nil || err.Error() != "No rules found." {
			t.Fatalf("Expected 'No rules found.' error, but got: %v", err)
		}
	})
}
