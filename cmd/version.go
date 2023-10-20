package cmd

import (
	"fmt"
)

// ProductName is the name of the product
const ProductName = "gpcloud"

var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

// GetVersionDisplay composes the parts of the version in a way that's suitable
// for displaying to humans.
func GetVersionDisplay() string {
	return fmt.Sprintf("%s - %s\n", ProductName, getHumanVersion())
}

func getHumanVersion() string {
	info := fmt.Sprintf("Version %s", Version)
	if Commit != "" {
		info += fmt.Sprintf(" (%s)", Commit)
	}
	if Date != "" {
		info += fmt.Sprintf(" - build on %s", Date)
	}

	// Strip off any single quotes added by the git information.
	return info
}
