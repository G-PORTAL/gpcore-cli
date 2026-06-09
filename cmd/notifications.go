package cmd

import (
	"buf.build/gen/go/gportal/gpcore/grpc/go/gpcore/api/cloud/v1/cloudv1grpc"
	cloudv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/cloud/v1"
	"github.com/G-PORTAL/gpcore-cli/pkg/api"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// NotificationsCommand registers the user-facing "notifications" command, which
// streams notifications via cloud.SubscribeNotifications. It shares the
// break/recv stream loop with "livelog" through api.StreamMessages; only the
// per-message rendering differs.
func NotificationsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(&cobra.Command{
		Use:                   "notifications",
		Short:                 "Live notification stream",
		Long:                  "Subscribe to and print live notifications for your account",
		DisableFlagsInUseLine: true,
		Args:                  cobra.OnlyValidArgs,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			conn := cobraCmd.Context().Value("conn").(*grpc.ClientConn)
			cloudClient := cloudv1grpc.NewCloudServiceClient(conn)
			stream, err := cloudClient.SubscribeNotifications(cobraCmd.Context(), &cloudv1.SubscribeNotificationsRequest{})
			if err != nil {
				return err
			}

			return api.StreamMessages(cobraCmd, stream, func(msg *cloudv1.SubscribeNotificationsResponse) {
				notification := msg.GetNotification()
				if notification == nil {
					return
				}

				switch {
				case notification.GetNode() != nil:
					node := notification.GetNode()
					label := text.Colors{text.FgCyan, text.BgBlack}.Sprint("Node")
					cobraCmd.Printf("[%s] %s (%s)\n", label, node.GetFqdn(), node.GetId())
				case notification.GetProject() != nil:
					project := notification.GetProject()
					label := text.Colors{text.FgGreen, text.BgBlack}.Sprint("Project")
					cobraCmd.Printf("[%s] %s (%s)\n", label, project.GetName(), project.GetId())
				case notification.GetUser() != nil:
					user := notification.GetUser()
					label := text.Colors{text.FgMagenta, text.BgBlack}.Sprint("User")
					cobraCmd.Printf("[%s] %s (%s)\n", label, user.GetFullName(), user.GetId())
				case notification.GetServerLog() != nil:
					m := notification.GetServerLog()
					label := text.Colors{text.FgYellow, text.BgBlack}.Sprint("ServerLog")
					t := m.GetUpdatedAt().AsTime().Format("15:04:05")
					cobraCmd.Printf("[%s] %s -> %s\n", label, t, m.GetMessage())
				case notification.GetHeartbeat() != nil:
					// Heartbeats keep the stream alive; ignore them in output.
				}
			})
		},
	})
}
