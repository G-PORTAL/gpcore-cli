package cmd

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

// SelfupdateCommand adds the selfupdate command to the root command. This command
// will download the latest version of the CLI from Github and replace the current
// binary with the new one. To access the gitlab API a secret token is required.
// This token is stored in cmd/secrets_gen.go and is not part of the repository.
func SelfupdateCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(&cobra.Command{
		Use:                   "selfupdate",
		Short:                 "Update the CLI to the latest version",
		Long:                  "Update the CLI to the latest version",
		DisableFlagsInUseLine: true,
		Args:                  cobra.OnlyValidArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			latest, found, err := selfupdate.DetectLatest(cmd.Context(), selfupdate.ParseSlug("G-PORTAL/gpcore-cli"))
			if err != nil {
				return fmt.Errorf("error occurred while detecting version: %w", err)
			}
			if !found {
				return fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
			}

			if latest.LessOrEqual(Version) {
				log.Printf("Current version (%s) is the latest", Version)
				return nil
			}

			exe, err := os.Executable()
			if err != nil {
				return errors.New("could not locate executable path")
			}

			if err := selfupdate.UpdateTo(cmd.Context(), latest.AssetURL, latest.AssetName, exe); err != nil {
				return fmt.Errorf("error occurred while updating binary: %w", err)
			}

			log.Printf("Successfully updated to version %s", latest.Version())
			return nil
		},
	})
}
