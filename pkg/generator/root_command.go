package generator

import (
	. "github.com/dave/jennifer/jen"
	"github.com/stoewer/go-strcase"
)

// GenerateRootCommand generates the RootCommand function, which will add the
// basic usage information to the root command and give the subcommand a
// anchor point.
func GenerateRootCommand(metadata SubcommandDefinition, targetFilename string) error {
	f := NewFile(escapePackage(metadata.Name))
	warningComment(f)

	f.ImportName("github.com/spf13/cobra", "cobra")

	commandName := "Root" + strcase.UpperCamelCase(metadata.Name) + "Command"
	f.Var().Add(Id(commandName).Op("=").Op("&").Qual("github.com/spf13/cobra", "Command").Values(Dict{
		Id("Use"):                   Lit(metadata.Name),
		Id("Short"):                 Lit(metadata.Description),
		Id("Long"):                  Lit(metadata.Description),
		Id("DisableFlagsInUseLine"): True(),
		Id("Args"):                  Qual("github.com/spf13/cobra", "MatchAll").Call(Qual("github.com/spf13/cobra", "ExactArgs").Call(Lit(0)), Qual("github.com/spf13/cobra", "OnlyValidArgs")),
		Id("RunE"): Func().Params(Id("cmd").Op("*").Qual("github.com/spf13/cobra", "Command"), Id("args").Index().String()).Error().Block(
			Return(Id("cmd").Dot("Usage").Call()),
		),
	}))

	return f.Save(targetFilename)
}
