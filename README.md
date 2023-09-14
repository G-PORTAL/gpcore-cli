# GPCloud CLI

Just run `go run gpc.go` to see the help menu.
On first run, it asks for credentials (client-id, client-secret) and saves them
in a config file. To create a new oauth client for the credentials, go to
https://panel.g-portal.cloud/user/settings/clients. The default config path is
located at `~/.gpc.yaml`.

## Basic usage
First you can list the projects you have access to:
```
$ go run gpc.go project list
```

Then you can select the project you want to work with:
```
$ go run gpc.go project use --id <project-id>
```

After a project was chosen, you can use the nodes commands:
```
$ go run gpc.go node list
```

## Development

All files which ends with ```_gen.go``` are autogenerated and should not be
edited manually. To regenerate them, run ```go generate```. If you want to
add a new subcommand, see ```pkg/generator/definition/``` for some examples.
Generated code files are NOT checked in, because the user can always generate
them himself. This is done to avoid merge conflicts.

You can always add custom subcommands without generating it. Just add a new
file to ```cmd/```. The file name will be the name of the subcommand.

# TODOs

* Add mandatory params (in addition to optional ones)
* Add more subcommands

* Gitlab build pipeline -> GoReleaser : 2023.9.[version]
* Auto-Update command to update to tool itself 

* Pagination support for long lists
* Custom description for subcommand flags
* No "usage" output on API error
* LiveLogs as a command
* Event log -> return when node ready
* Project Log watcher
