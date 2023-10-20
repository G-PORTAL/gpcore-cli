# GPORTAL Cloud CLI

Just run `go run gpc.go` to see the help menu.
On first run, it asks for credentials (client-id, client-secret) and saves them
in a config file. To create a new oauth client for the credentials, go to
https://panel.g-portal.cloud/user/settings/clients. The default config path is
located at `~/.gpc.yaml`.

## Install and update

Use the install.sh script to install the latest version of the tool. It will
download the latest release from github and install it to ~/bin. Specify a
GitLab token to download the latest build from the GitLab CI.

```
$ TOKEN=glpat-xxxxxxxxxxxxxxxxxxxxxxxxx ./install.sh
```

To update the tool itself, use the selfupdate command:

```
$ gpc selfupdate
```

## Basic usage
First you can list the projects you have access to:
```
$ gpc project list
```

Then you can select the project you want to work with:
```
$ gpc project use --id <project-id>
```

After a project was chosen, you can use the nodes commands:
```
$ gpc node list
```

Enums will be specified without the prefix:
```
$ gpc project create --environment staging --name "New Project"
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

* Pagination support for long lists
* Event log -> return when node ready
* Project Log watcher
* Separate admin and non admin commands (or check if the user is an admin?)