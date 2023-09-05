# GPCloud CLI

Just run `go run gpc.go` to see the help menu.
On first run, it asks for credentials (client-id, client-secret) and saves them
in a config file. To create a new oauth client for the credentials, go to
https://panel.g-portal.cloud/user/settings/clients. The default config path is
located at `~/.gpc.yaml`.

Before you can use the tool, you need to generate to code with:

```
go generate
```

# Basic usage
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

# TODOs

* Format output for console usage (table formatter?)
* Add hooks to subcommands (pre/post)
* Add mandatory params (in addition to optional ones)
* Gitlab build pipeline
* Auto-Update command to update to tool itself
