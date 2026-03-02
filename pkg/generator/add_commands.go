package generator

import (
	. "github.com/dave/jennifer/jen"
	"github.com/stoewer/go-strcase"
	"sort"
)

// CommandGroup defines a group of commands for help output organization.
type CommandGroup struct {
	ID    string
	Title string
}

// CommandGroups defines the available command groups in display order.
var CommandGroups = []CommandGroup{
	{ID: "resources", Title: "Cloud Resources:"},
	{ID: "networking", Title: "Networking:"},
	{ID: "billing", Title: "Billing & Reporting:"},
	{ID: "admin", Title: "Administration:"},
}

// GenerateAddCommands generates the AddGeneratedCommands function, which will
// add all generated commands to the root command, including command group
// registration for organized help output.
func GenerateAddCommands(commands []string, targetFilename string) error {
	f := NewFile("cmd")
	warningComment(f)

	f.ImportName("github.com/spf13/cobra", "cobra")
	for _, command := range commands {
		pkg := escapePackage(command)
		f.ImportName("github.com/G-PORTAL/gpcore-cli/cmd/"+pkg, pkg)
	}

	f.Func().Id("AddGeneratedCommands").
		Params(Id("cmd").Op("*").Qual("github.com/spf13/cobra", "Command")).
		BlockFunc(func(g *Group) {
			// Register command groups
			for _, group := range CommandGroups {
				g.Id("cmd").Dot("AddGroup").Call(
					Op("&").Qual("github.com/spf13/cobra", "Group").Values(Dict{
						Id("ID"):    Lit(group.ID),
						Id("Title"): Lit(group.Title),
					}))
			}
			g.Line()

			// Sort commands for consistent output
			sorted := make([]string, len(commands))
			copy(sorted, commands)
			sort.Strings(sorted)

			// Add commands
			for _, command := range sorted {
				pkg := escapePackage(command)
				g.Id("cmd").Dot("AddCommand").Call(
					Qual("github.com/G-PORTAL/gpcore-cli/cmd/"+pkg, "Root"+strcase.UpperCamelCase(pkg)+"Command"))
			}
		})

	return f.Save(targetFilename)
}
