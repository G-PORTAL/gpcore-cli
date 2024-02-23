package agent

import (
	"fmt"
	"github.com/G-PORTAL/gpcore-cli/cmd"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/G-PORTAL/gpcore-cli/pkg/consts"
	"github.com/G-PORTAL/gpcore-go/pkg/gpcore/client"
	"github.com/spf13/cobra"
	"net"
	"strconv"
	"time"
)

var printVersion = false

func New() *cobra.Command {
	rootCmd := cobra.Command{
		Use:   consts.BinaryName,
		Short: fmt.Sprintf("%s is the command line tool for interacting with the GP Cloud API", consts.BinaryName),
		Long:  fmt.Sprintf("%s is the command line tool for interacting with the GPCore API", consts.BinaryName),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if printVersion {
				cobraCmd.Print(cmd.GetVersionDisplay())
				return nil
			}
			cobraCmd.Println(cobraCmd.UsageString())
			return nil
		},
		SilenceUsage: true,
	}

	// Application information
	rootCmd.Flags().BoolVarP(&printVersion, "version", "V", false, "print version information and quit")

	// GPCORE API
	// TODO: Will set on first run (when agent starts),the following client calls will ignore these, so, move this to the agent only or reconnect the API on every change
	rootCmd.PersistentFlags().StringVarP(&config.Endpoint, "endpoint", "e", client.DefaultEndpoint, "set API endpoint")

	// Output formats and verbosity
	rootCmd.PersistentFlags().BoolVarP(&config.Verbose, "verbose", "v", false, "verbose mode")
	rootCmd.PersistentFlags().BoolVarP(&config.JSONOutput, "json", "j", false, "output as JSON")
	rootCmd.PersistentFlags().BoolVarP(&config.CSVOutput, "csv", "x", false, "output as CSV")

	// Special client commands
	cmd.SelfupdateCommand(&rootCmd)
	cmd.LiveLogCommand(&rootCmd)
	//InteractiveCLICommand(&rootCmd)

	// Autogenerated commands
	cmd.AddGeneratedCommands(&rootCmd)

	// Special commands
	rootCmd.AddCommand(agentCommand)
	rootCmd.AddCommand(setupCommand)

	return &rootCmd
}

func IsAgentRunning() bool {
	// It makes a difference if we execute this as the agent or from "outside"
	// (the client). So we need to check if the agent is running by checking
	// if the port is open or not.
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(consts.AgentHost, strconv.Itoa(consts.AgentPort)), 200*time.Millisecond)
	if err != nil {
		return false
	}
	if conn != nil {
		defer conn.Close()
		return true
	}
	return false
}

var agentCommand = &cobra.Command{
	Use:                   "agent",
	Short:                 "Agent related actions",
	Long:                  "Agent related actions",
	DisableFlagsInUseLine: true,
	Args:                  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		return cobraCmd.Usage()
	},
}
