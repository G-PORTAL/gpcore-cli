package main

import (
	"os"

	command "gpcloud-cli/cmd"
)

//go:generate go run ./pkg/generator/generator.go

func main() {
	cmd := command.New()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
