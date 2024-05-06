package generator

import (
	. "github.com/dave/jennifer/jen"
	"github.com/stoewer/go-strcase"
	"strings"
)

// GenerateRootCommand generates the RootCommand function, which will add the
// basic usage information to the root command and give the subcommand a
// anchor point.
func GenerateRootCommand(metadata SubcommandDefinition, targetFilename string) error {
	f := NewFile(escapePackage(metadata.Name))
	warningComment(f)

	f.ImportName("github.com/spf13/cobra", "cobra")
	f.ImportName("github.com/G-PORTAL/gpcore-cli/cmd/help", "help")

	commandName := "Root" + strcase.UpperCamelCase(metadata.Name) + "Command"
	command := strings.ReplaceAll(metadata.Name, "_", "-")
	f.Var().Add(Id(commandName).Op("=").Op("&").Qual("github.com/spf13/cobra", "Command").Values(Dict{
		Id("Use"):              Lit(command),
		Id("Short"):            Lit(metadata.Description),
		Id("Long"):             Lit(metadata.Description),
		Id("SilenceUsage"):     True(),
		Id("SilenceErrors"):    True(),
		Id("TraverseChildren"): True(),
		Id("Args"):             Qual("github.com/spf13/cobra", "OnlyValidArgs"),
		Id("RunE"):             Qual("github.com/G-PORTAL/gpcore-cli/cmd/help", "UnknownSubcommandAction"),
	}))

	return f.Save(targetFilename)
}
