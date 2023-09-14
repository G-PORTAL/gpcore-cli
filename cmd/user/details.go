package user

import (
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"encoding/json"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"gpcloud-cli/pkg/config"
	"os"
)

var detailsCmd = &cobra.Command{
	Use:                   "details",
	Short:                 "Get details about the current user",
	Long:                  "Get details about the current user",
	DisableFlagsInUseLine: true,
	Args:                  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		user := cmd.Context().Value("user").(*cloudv1.User)

		tbl := table.NewWriter()
		tbl.SetStyle(table.StyleRounded)
		tbl.SetOutputMirror(os.Stdout)
		tbl.AppendRow([]interface{}{"Type", user.GetType()})
		tbl.AppendRow([]interface{}{"Id", user.GetId()})
		tbl.AppendRow([]interface{}{"Keycloak ID", user.GetKeycloakId()})
		tbl.AppendRow([]interface{}{"Username", user.GetUsername()})
		tbl.AppendRow([]interface{}{"Full Name", user.GetFullName()})
		tbl.AppendRow([]interface{}{"Email", user.GetEmail()})

		tbl.AppendRow([]interface{}{"Is confirmed?", user.GetConfirmed()})
		tbl.AppendRow([]interface{}{"Is locked?", user.GetLocked()})

		if !config.JSONOutput {
			tbl.Render()
		}

		if config.JSONOutput {
			jsonData, err := json.MarshalIndent(user, "", "  ")
			if err != nil {
				return err
			}
			cmd.Println(string(jsonData))
		}

		return nil
	},
}

func init() {
	RootUserCommand.AddCommand(detailsCmd)
}
