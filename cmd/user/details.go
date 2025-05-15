package user

import (
	"github.com/G-PORTAL/gpcore-cli/pkg/client"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/G-PORTAL/gpcore-cli/pkg/protobuf"
	"github.com/charmbracelet/ssh"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var detailsCmd = &cobra.Command{
	Use:                   "details",
	Short:                 "Get details about the current user",
	Long:                  "Get details about the current user",
	DisableFlagsInUseLine: true,
	Args:                  cobra.NoArgs,
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		ctx := client.ExtractContext(cobraCmd)
		user := client.GetUser(ctx)
		sshSession := ctx.Value("ssh").(*ssh.Session)

		tbl := table.NewWriter()
		tbl.SetStyle(table.StyleRounded)
		tbl.SetOutputMirror(*sshSession)
		tbl.AppendRow([]interface{}{"Type", user.GetType()})
		tbl.AppendRow([]interface{}{"Has admin credentials?", config.HasAdminConfig()})
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
			jsonData, err := protobuf.MarshalIndent(user)
			if err != nil {
				return err
			}
			cobraCmd.Println(string(jsonData))
		}

		return nil
	},
}

func init() {
	RootUserCommand.AddCommand(detailsCmd)
}
