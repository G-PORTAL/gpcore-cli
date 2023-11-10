package agent

import (
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/spf13/cobra"
	"os"
)

var stopCmd = &cobra.Command{
	Use:                   "stop",
	Short:                 "Stop the agent",
	Long:                  "Stop the agent",
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		sshSession := cobraCmd.Context().Value("ssh").(*ssh.Session)
		log.SetOutput(*sshSession)

		// There is a separate IsAgentRunning() function, but this is just for
		// external use. Internally (like this code), we just can check the
		// IsRunning variable.
		if IsRunning {
			log.Infof("Stopping agent ...")
			DoneChan <- os.Interrupt
		}
		return nil
	},
}

func init() {
	AgentCommand.AddCommand(stopCmd)
}
