package generator

import (
	. "github.com/dave/jennifer/jen"
	"github.com/stoewer/go-strcase"
)

// GenerateAddCommands generates the AddGeneratedCommands function, which will
// add all generated commands to the root command.
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
			for _, command := range commands {
				pkg := escapePackage(command)
				g.Id("cmd").Dot("AddCommand").Call(
					Qual("github.com/G-PORTAL/gpcore-cli/cmd/"+pkg, "Root"+strcase.UpperCamelCase(pkg)+"Command"))
			}
		})

	return f.Save(targetFilename)
}
