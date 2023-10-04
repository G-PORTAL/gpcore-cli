package main

import (
	"github.com/google/martian/log"
	"github.com/shirou/gopsutil/v3/process"
	"gpcloud-cli/pkg/agent"
	"gpcloud-cli/pkg/client"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

//go:generate go run ./pkg/generator/generator.go
//go:generate gofmt -s -w ./cmd/

func main() {
	log.SetLevel(log.Info)

	// TODO: Put these in cobra commands
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
				if name == "gpc" {
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
	conn, err := net.DialTimeout("tcp", "localhost:9001", 3*time.Second)
	if err != nil {
		log.Infof("Starting agent in background ...")
		err := exec.Command(os.Args[0], "agent", "start").Start()
		if err != nil {
			panic(err)
		}
		// Give the agent some time to start
		time.Sleep(2 * time.Second)
	} else {
		err := conn.Close()
		if err != nil {
			return
		}
	}

	// Start the client
	c, err := client.NewClient()
	if err != nil {
		log.Errorf("Failed to create client: %s", err.Error())
	}
	client.Execute(c, strings.Join(os.Args[1:], " "))
}
