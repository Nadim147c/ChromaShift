package main

import (
	"embed"

	"cshift/cmd"
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
