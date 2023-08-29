package main

import (
	"os"

	command "gpcloud-cli/cmd"
)

func main() {
	cmd := command.New()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
