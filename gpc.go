package main

import (
	"fmt"
	"github.com/G-PORTAL/gpcloud-cli/pkg/agent"
	"github.com/G-PORTAL/gpcloud-cli/pkg/client"
	"github.com/G-PORTAL/gpcloud-cli/pkg/consts"
	"github.com/charmbracelet/log"
	"github.com/shirou/gopsutil/v3/process"
	"gopkg.in/op/go-logging.v1"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

//go:generate go run ./pkg/generator/generator.go
//go:generate gofmt -s -w ./cmd/

// GPCloud CLI (gpc in short) is a command line interface for the G-Portal Cloud
// API. It is written in Go and uses the Cobra framework. The commands are auto
// generated from definition files in the pkg/generator/definition directory
// (the generated files get the _gen postfix). Custom commands can be added in
// the cmd/ directory and will not be overwritten. If there is a <command>_pre.go
// or a <command>_post.go file in the cmd/ directory, it will be executed before
// or after the command. This can be used to modify the response or to add some
// additional logic.
//
// The client and the agent communicate via SSH. The agent is a SSH server that
// executes commands on the G-Portal Cloud API. The client is a SSH client that
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

	// TODO: Put these in cobra commands?

	// If we are in agent mode, start the agent
	if len(os.Args) > 2 && os.Args[1] == "agent" {
		// Stop running agent(s)
		if os.Args[2] == "stop" {
			processes, err := process.Processes()
			if err != nil {
				panic(err)
				return
			}
			for _, p := range processes {
				name, err := p.Name()
				if err != nil {
					panic(err)
					return
				}
				if name == consts.BinaryName {
					err := p.Kill()
					if err != nil {
						panic(err)
						return
					}
				}
			}
		}

		// Start the agent
		if os.Args[2] == "start" {
			agent.StartServer()
		}

		return
	}

	// Otherwise start the client
	client.Setup()

	// Launch agent in the background if not already running. To do that, we
	// try to connect to it.
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%v", consts.AgentHost, consts.AgentPort), 3*time.Second)
	if err != nil {
		log.Infof("Starting agent in background ...")
		err = exec.Command(os.Args[0], "agent", "start").Start()
		if err != nil {
			panic(err)
		}
		// Give the agent some time to start
		time.Sleep(2 * time.Second)
	} else {
		err = conn.Close()
		if err != nil {
			return
		}
	}

	// Start the client
	c, err := client.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %s", err.Error())
	}

	defer c.Close()
	client.Execute(c, strings.Join(os.Args[1:], " "))
}
