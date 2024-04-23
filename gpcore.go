package main

import (
	"github.com/G-PORTAL/gpcore-cli/cmd/agent"
	"github.com/G-PORTAL/gpcore-cli/pkg/client"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/G-PORTAL/gpcore-cli/pkg/consts"
	"github.com/charmbracelet/log"
	"gopkg.in/op/go-logging.v1"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//go:generate go run ./pkg/generator/generator.go
//go:generate gofmt -s -w ./cmd/

// GPCORE CLI (gpc in short) is a command line interface for the GPCORE
// API. It is written in Go and uses the Cobra framework. The commands are auto
// generated from definition files in the pkg/generator/definition directory
// (the generated files get the _gen postfix). Custom commands can be added in
// the cmd/ directory and will not be overwritten. If there is a <command>_pre.go
// or a <command>_post.go file in the cmd/ directory, it will be executed before
// or after the command. This can be used to modify the response or to add some
// additional logic.
//
// The client and the agent communicate via SSH. The agent is a SSH server that
// executes commands on the GPCORE API. The client is a SSH client that
// connects to the agent and executes commands on the agent. The agent is only
// listening on localhost and the SSH keypair is stored in the users home
// directory. The private key is secured with a password. This client/server
// architecture is used to leave the connection open and to avoid the need to
// authenticate on every request. The agent is started in the background if it
// is not already running.
func main() {
	// Initialize logger
	var format = logging.MustStringFormatter(`%{color}%{time:15:04:05} %{shortfunc} [%{level:.4s}]%{color:reset} %{message}`)
	var backend = logging.AddModuleLevel(logging.NewBackendFormatter(logging.NewLogBackend(os.Stderr, "", 0), format))
	backend.SetLevel(logging.ERROR, "")
	logging.SetBackend(backend)

	command := strings.Join(os.Args[1:], " ")
	if command == "" {
		command = "help"
	}

	// Some commands which can run without a valid config
	if command == "--version" ||
		command == "--help" ||
		command == "help" ||
		command == "agent" ||
		strings.HasPrefix(command, "agent setup") {
		RunCommandWithoutClient()
		return
	}

	// If we reach this point, we need a valid config, otherwise stuff will
	// break. So we ensure we have at least "some" config.
	if !config.HasConfig() {
		log.Info("No config found, creating new one with `gpcore agent setup`")
		return
	}

	// agent start is a special command, because we do not have a server at
	// that moment, so we can not connect to anything. For that, we need to handle
	// that special case and bypass the normal command execution.
	if command == "agent start" {
		RunCommandWithoutClient()
		return
	}

	// Same goes for agent stop. We only stop the agent if there is a running agent.
	if command == "agent stop" && !agent.IsAgentRunning() {
		log.Errorf("Agent is not running")
		return
	}

	// Launch agent in the background if not already running. To do that, we
	// try to connect to it.
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(consts.AgentHost, strconv.Itoa(consts.AgentPort)), 30*time.Second)
	if err != nil {
		log.Infof("Starting agent in background ...")
		err = exec.Command(os.Args[0], "agent", "start").Start()
		if err != nil {
			panic(err)
		}
		// Give the agent some time to start
		// TODO: Optimize this
		time.Sleep(1 * time.Second)
	} else {
		err = conn.Close()
		if err != nil {
			return
		}
	}

	// Start the client
	c, err := client.NewClient()
	if err != nil {
		log.Errorf("Failed to create client: %s", err.Error())
		log.Fatal("Check your config file and/or reset it with \"gpcore agent setup\"")
	}
	defer c.Close()

	client.Execute(c, command)
}

func RunCommandWithoutClient() {
	rootCmd := agent.New()
	rootCmd.SetArgs(os.Args[1:])
	err := rootCmd.Execute()
	if err != nil {
		log.Errorf("Command failed: %s", err)
	}
}
