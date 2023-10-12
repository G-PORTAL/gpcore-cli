package project

import (
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"github.com/spf13/cobra"
	"gpcloud-cli/pkg/client"
	"gpcloud-cli/pkg/config"
)

func ListHookPost(resp *cloudv1.ListProjectsResponse, cobraCmd *cobra.Command) ([]map[string]string, error) {
	res := make([]map[string]string, 0)

	ctx := client.ExtractContext(cobraCmd)
	config := ctx.Value("config").(*config.SessionConfig)
	user := ctx.Value("user").(*cloudv1.User)

	for i := range resp.Projects {
		name := resp.Projects[i].Name
		// Mark active project
		if config.CurrentProject != nil && resp.Projects[i].Id == *config.CurrentProject {
			name = "*" + name
		}

		// Mark default project
		for _, member := range resp.Projects[i].GetMembers() {
			if member.GetUser().GetId() == user.GetId() && member.GetDefault() {
				name = name + " (default)"
			}
		}

		res = append(res, map[string]string{
			"Name": name,
		})
	}

	return res, nil
}
