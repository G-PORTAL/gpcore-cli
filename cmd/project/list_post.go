package project

import (
	cloudv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/cloud/v1"
	"github.com/spf13/cobra"
	"gpcloud-cli/pkg/config"
)

func ListHookPost(resp *cloudv1.ListProjectsResponse, cmd *cobra.Command) {
	session := cmd.Context().Value("session").(*config.SessionConfig)
	user := cmd.Context().Value("user").(*cloudv1.User)

	for i := range resp.Projects {
		// Mark active project
		if resp.Projects[i].Id == *session.CurrentProject {
			name := "*" + resp.Projects[i].Name
			resp.Projects[i].Name = name
		}

		// Mark default project
		for _, member := range resp.Projects[i].GetMembers() {
			if member.GetUser().GetId() == user.GetId() && member.GetDefault() {
				name := resp.Projects[i].Name + " (default)"
				resp.Projects[i].Name = name
			}
		}
	}
}
