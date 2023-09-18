package cmd

import (
	"archive/zip"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// SelfupdateCommand adds the selfupdate command to the root command. This command
// will download the latest version of the CLI from gitlab and replace the current
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
			// Download latest artifact from gitlab
			url := "https://gitlab.g-portal.se/api/v4/projects/gpcloud%2Fgpcloud-cli/jobs/artifacts/master/download?job=build"

			client := &http.Client{}
			req, err := http.NewRequest("GET", url, nil)
			req.Header.Set("PRIVATE-TOKEN", GitlabSecretToken)
			cmd.Println("Downloading latest version ...")
			resp, err := client.Do(req)
			defer resp.Body.Close()
			if err != nil {
				return err
			}

			// Save the zip file to disk
			targetFilepath := filepath.Join(os.TempDir(), "gpcloud-cli.zip")
			defer os.Remove(targetFilepath)
			f, err := os.Create(targetFilepath)
			defer f.Close()
			if err != nil {
				return err
			}
			_, err = io.Copy(f, resp.Body)
			if err != nil {
				return err
			}

			// Extract zip archive
			cmd.Println("Extracting content ...")
			archive, err := zip.OpenReader(targetFilepath)
			if err != nil {
				return err
			}
			defer archive.Close()

			binaryToReplace, err := os.Executable()
			cleanedFilepath, err := filepath.EvalSymlinks(binaryToReplace)
			if err != nil {
				return err
			}
			tempFilepath := cleanedFilepath + ".tmp"

			for _, file := range archive.File {
				if file.Name == "gpc" {
					cmd.Printf("Replacing binary %s ...\n", cleanedFilepath)
					dstFile, err := os.OpenFile(tempFilepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
					defer dstFile.Close()
					if err != nil {
						return err
					}
					// Copy new file to disk
					fh, err := file.Open()
					_, err = io.Copy(dstFile, fh)
					if err != nil {
						return err
					}

					// Swap files
					_ = os.Remove(binaryToReplace)
					err = os.Rename(tempFilepath, binaryToReplace)
					if err != nil {
						return err
					}

					cmd.Println("Done")
					break
				}
			}

			return nil
		},
	})
}
