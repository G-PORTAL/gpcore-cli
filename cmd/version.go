package cmd

import (
	"fmt"
	"github.com/G-PORTAL/gpcore-cli/pkg/consts"
)

var (
	version string
	commit  string
	date    string
)

var (
	Version = valueOrFallback(version, func() string { return "dev" })
	Commit  = valueOrFallback(commit, func() string { return "" })
	Date    = valueOrFallback(date, func() string { return "" })
)

// GetVersionDisplay composes the parts of the version in a way that's suitable
// for displaying to humans.
func GetVersionDisplay() string {
	return fmt.Sprintf("%s - %s\n", consts.BinaryName, GetHumanVersion())
}

func GetHumanVersion() string {
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

func valueOrFallback(val string, fn func() string) string {
	if val != "" {
		return val
	}
	return fn()
}
