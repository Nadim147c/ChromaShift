package main

import (
	"colorize/cmd"
	"embed"
)

//go:embed rules/*
var StaticRules embed.FS

//go:embed config.toml
var StaticConfig string

func main() {
	cmd.StaticRulesDirectory = StaticRules
	cmd.StaticConfig = StaticConfig
	cmd.Execute()
}
