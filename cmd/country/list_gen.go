// Code generated DO NOT EDIT
// This code is AUTOGENERATED and will be overwritten by "go generate", so
// editing this file is a waste of time. To make changes, edit the template
// in pkg/generator/template/subcommand.tmpl. If you want to execute things
// before or after the command is executed, use a hook. See the usage_hook.go
// as an example.

package country

import (
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"encoding/json"
	"github.com/G-PORTAL/gpcloud-go/pkg/gpcloud/client"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:                   "list",
	Short:                 "List all available countries",
	Long:                  "List all available countries",
	DisableFlagsInUseLine: true,
	Args:                  cobra.OnlyValidArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		conn := cmd.Context().Value("conn").(*client.Client)
		resp, err := conn.CloudClient().ListCountries(cmd.Context(), &cloudv1.ListCountriesRequest{})
		if err != nil {
			return err
		}

		// TODO: Call hook if exist

		// TODO: Only Output response as json if requests with --json flag,
		// Otherwise, output a human readable table
		jsonData, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}
		cmd.Println(string(jsonData))

		return nil
	},
}

func init() {

	RootCountryCommand.AddCommand(listCmd)
}
