# GPCORE CLI

Commandline tool to access the GPCORE API.

![header](logo.png)


## Install and update

Download the latest binary from the release page, make it executable and run it.
To update the CLI, just run the ```gpcore selfupdate``` command. It will download the
latest release from GitHub and replace the current binary.

## Overview

The commandline tool works in two different modes: As a server and as a client.
The server is a standard ssh server, which does the heavy lifting. The server
logs into the GPCORE API once and wait for commands from the client.

The server acts as a proxy between the client and the API. The client connects
to the server and sends commands to it. The server executes the commands and
returns the result to the client. With this architecture, the client does not
open up a new connection for every command, which saves time and resources.

On first run, it asks for credentials (client-id, client-secret, username and
password). Non-critical information will be stored in a config file
( ~/.config/gportal/config.json ). Sensitive information will be stored in the
keyring (if supported by the OS). The keyring is encrypted with the user's
password. To secure the connection between client and server, a SSH public/private
key pair will be generated and secured with a passphrase. The passphrase is
the same as the password for the GPCORE account. This way, the connection
between client and server is secured and no other ssh client can connect to it.

If you messed up your config, the sensitive data in the keyring or the public/private
key, you can reset everything with the ```gpcore agent setup``` command. Use the
```--admin``` flag to setup the admin credentials as well.

The agent (SSH server) will start automatically and place it in the background,
until the user actively stops it with ```gpcore agent stop```. So the first command will
take a little longer (because the agent has to start), but all following commands
will be executed immediately.

The client itself is a simple SSH client. It connects to the agent and sends
commands to it. The result is printed to stdout. You can use the standard
SSH command (ssh) to connect through it, but it is not that convenient.

## Usage

The commandline tool is separated into subcommands. To get a list of all
available subcommands, run ```gpcore help```. To get help for a specific
subcommand, run ```gpcore help <subcommand>```.

As and example, to list all projects, run ```gpcore project list```. If you run
just the subcommand without any arguments, you will get a list of all available
actions for that subcommand.

Some commands need flags or specific parameters. To get a list of all flags and
parameters, run ```gpcore help <subcommand> <action>```. For example, to change
the active project, run ```gpcore project set-active --id <project-uuid>```.

By default, the output is formatted as a ASCII table. If you want to pass the
output to other tools for processing, you can append the flag ```--csv``` or
```--json```. For example: ```gpcore project list --json | jq 'length```

## Development

All files which ends with ```_gen.go``` are autogenerated and should not be
edited manually. To regenerate them, run ```go generate```. If you want to
add a new subcommand, see ```pkg/generator/definition/``` for some examples.
Generated code files are NOT checked in, because the user can always generate
them himself. This is done to avoid merge conflicts.

You can always add custom subcommands without generating it. Just add a new
file to ```cmd/```. The file name will be the name of the subcommand.

With hooks, you can inject code at some points in auto generated code. For
example, if you want to remove some colums, format certain colums or validate
input, you can do this with hooks. Create a file in the same package with the
same name of the action (auto generated file), but with the prefix ```_pre.go```
to execute code before the action and ```_post.go``` to execute code after the
action but before the output is printed. As an example, see ```project/list_post.go```.


### Ongoing tasks

* catch unauthorized error and ask for user/pass
* Pagination support for long lists (pending because of Jennifer migration)
* Complete the API endpoints, so everything is accessible through the CLI tool
