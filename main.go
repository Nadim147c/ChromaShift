package main

import (
	"colorize/cmd"
	"embed"
)

//go:embed rules/*
var StaticRules embed.FS

func main() {
	cmd.StaticRulesDirectory = StaticRules
	cmd.Execute()
}
