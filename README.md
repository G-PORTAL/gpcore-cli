# gpcloud cli

Just run `go run gpc.go` to see the help menu.
On first run, it asks for credentials (client-id, client-secret) and saves them in a config file.
default config path: `~/.gpc.yaml`

# Basic usage
First you can list the projects you have access to:
```
$ go run gpc.go projects list
```

Then you can select the project you want to work with:
```
$ go run gpc.go projects use --id <project-id>
```

After a project was choosen, you can use the nodes commands:
```
$ go run gpc.go nodes list
```

Its more a proof of concept right now to have something to work with when generating the code automatically.
nodes and project folder should be generated from the proto files.