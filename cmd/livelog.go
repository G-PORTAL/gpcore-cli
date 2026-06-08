package cmd

import (
	"buf.build/gen/go/gportal/gpcore/grpc/go/gpcore/api/admin/v1/adminv1grpc"
	adminv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/admin/v1"
	cloudv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/cloud/v1"
	"github.com/G-PORTAL/gpcore-cli/pkg/api"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func LiveLogCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(&cobra.Command{
		Use:                   "livelog",
		Short:                 "Live log stream",
		Long:                  "Live log stream",
		DisableFlagsInUseLine: true,
		Args:                  cobra.OnlyValidArgs,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			conn := cobraCmd.Context().Value("conn").(*grpc.ClientConn)
			admin := adminv1grpc.NewAdminServiceClient(conn)
			stream, err := admin.SubscribeServerLogs(cobraCmd.Context(), &adminv1.SubscribeServerLogsRequest{})
			if err != nil {
				return err
			}

			// We wait for (and print out) relevant notifications until the break
			// request is received.
			return api.StreamMessages(cobraCmd, stream, func(msg *adminv1.SubscribeServerLogsResponse) {
				// TODO: Filter for source/server/datacenter
				// TODO: Only above level ...

				if msg.GetNotification().GetServerLog() != nil {
					m := msg.GetNotification().GetServerLog()

					color := text.Colors{text.FgWhite}
					switch m.GetLevel() {
					case cloudv1.ServerLogLevelType_SERVER_LOG_LEVEL_TYPE_WARNING:
						color = text.Colors{text.FgYellow}
					case cloudv1.ServerLogLevelType_SERVER_LOG_LEVEL_TYPE_ERROR:
						color = text.Colors{text.FgRed}
					}

					source := ""
					switch m.GetSource() {
					case cloudv1.ServerLogSourceType_SERVER_LOG_SOURCE_TYPE_METADATA:
						source = text.Colors{text.FgGreen, text.BgBlack}.Sprint("Metadata")
					case cloudv1.ServerLogSourceType_SERVER_LOG_SOURCE_TYPE_IRONIC:
						source = text.Colors{text.FgYellow, text.BgBlack}.Sprint("Ironic")
					case cloudv1.ServerLogSourceType_SERVER_LOG_SOURCE_TYPE_INTERNAL:
						source = text.Colors{text.FgMagenta, text.BgBlack}.Sprint("Internal")
					}
					time := m.GetUpdatedAt().AsTime().Format("15:04:05")
					server := text.Colors{text.FgCyan, text.BgBlack}.Sprint(m.GetServer().Name)
					datacenter := text.Colors{text.FgBlue, text.BgBlack}.Sprint(m.GetServer().GetDatacenter().Name)

					cobraCmd.Printf("%s: [%s] [%s] [%s] ->  %s\n", color.Sprint(time), datacenter, server, source, color.Sprint(m.GetMessage()))
				}
			})
		},
	})
}
