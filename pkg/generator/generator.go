//go:build ignore

package main

import (
	"github.com/gertd/go-pluralize"
	"github.com/stoewer/go-strcase"
	"gopkg.in/yaml.v3"
	"gpcloud-cli/pkg/generator"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	generatedFileSuffix := "_gen"

	log.Println("Generate subcommands ...")

	pl := pluralize.NewClient()
	templateFuncMap := template.FuncMap{
		"Title":     strings.Title,
		"ToLower":   strings.ToLower,
		"ToUpper":   strings.ToUpper,
		"ToKebab":   strcase.KebabCase,
		"ToSnake":   strcase.SnakeCase,
		"ToCamel":   strcase.LowerCamelCase,
		"Pluralize": pl.Plural,
		"HasPrefix": strings.HasPrefix,
		"HasHook": func(command string, subcommand string, hookType string) bool {
			if _, err := os.Stat("./cmd/" + command + "/" + subcommand + "_" + hookType + ".go"); !os.IsNotExist(err) {
				log.Printf("  Include hook %s/%s for type %s", command, subcommand, hookType)
				return true
			}
			return false
		},
		"IsEnumType": func(s string) bool {
			return strings.Contains(s, ".")
		},
		"StripPackage": func(s string) string {
			parts := strings.Split(s, ".")
			return parts[1]
		},
		"EnumToProto": func(enumType string, value string) string {
			// cloudv1.ProjectEnvironment -> ProjectEnvironment_PROJECT_ENVIRONMENT_[VALUE]
			parts := strings.Split(enumType, ".")
			return parts[0] + "." + parts[1] + "_" + strcase.UpperSnakeCase(parts[1]) + "_" + strings.ToUpper(value)
		},
		"EnumToValue": func(enumType string) string {
			// PROJECT_ENVIRONMENT_[VALUE] -> VALUE
			parts := strings.Split(enumType, "_")
			return parts[len(parts)-1]
		},
	}

	// Read in the template files
	subcommandTemplate, err := template.
		New("subcommand.tmpl").
		Funcs(templateFuncMap).
		ParseFiles("./pkg/generator/template/subcommand.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	rootcommandTemplate, err := template.
		New("rootcommand.tmpl").
		Funcs(templateFuncMap).
		ParseFiles("./pkg/generator/template/rootcommand.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	addcommandsTemplate, err := template.
		New("addcommands.tmpl").
		Funcs(templateFuncMap).
		ParseFiles("./pkg/generator/template/addcommands.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	var targetFile *os.File

	// Get all definitions
	definitionFiles, err := os.ReadDir("./pkg/generator/definition")
	if err != nil {
		log.Fatal(err)
	}

	for _, definitionFile := range definitionFiles {
		log.Printf("Generate subcommand as definied in %s ...\n", definitionFile.Name())
		definition, err := os.ReadFile("./pkg/generator/definition/" + definitionFile.Name())
		if err != nil {
			log.Fatal(err)
		}
		metadata := generator.SubcommandDefinition{}
		err = yaml.Unmarshal(definition, &metadata)
		if err != nil {
			log.Fatal(err)
		}

		subcommandName := strings.TrimSuffix(definitionFile.Name(), filepath.Ext(definitionFile.Name()))
		metadata.Name = subcommandName

		// Create directory if not exist
		if _, err := os.Stat("./cmd/" + subcommandName); os.IsNotExist(err) {
			log.Printf("  Create directory ./cmd/%s ...\n", subcommandName)
			err = os.Mkdir("./cmd/"+subcommandName, 0755)
			if err != nil {
				log.Fatal(err)
			}
		}
		// Create root command if not exist
		if _, err := os.Stat("./cmd/" + subcommandName + "/root.go"); os.IsNotExist(err) {
			log.Printf("  Create root command ./cmd/%s/root"+generatedFileSuffix+".go ...\n", subcommandName)
			targetFile, err = os.Create("./cmd/" + subcommandName + "/root" + generatedFileSuffix + ".go")
			if err != nil {
				log.Fatal(err)
			}
			err = rootcommandTemplate.Funcs(templateFuncMap).Execute(targetFile, metadata)
			if err != nil {
				log.Fatal(err)
			}
			err = targetFile.Close()
			if err != nil {
				log.Fatal(err)
			}
		}

		// Generate all subcommands
		for action, meta := range metadata.Actions {
			// Check if the subcommand is overwritten by the user
			if _, err := os.Stat("./cmd/" + subcommandName + "/" + strcase.SnakeCase(action) + ".go"); !os.IsNotExist(err) {
				log.Printf("  Subcommand %s for %s already exists, skipping ...\n", action, subcommandName)
				continue
			}
			log.Printf("  Generate subcommand %s ...\n", action)
			targetFile, err = os.Create("./cmd/" + subcommandName + "/" + strcase.SnakeCase(action) + generatedFileSuffix + ".go")
			if err != nil {
				log.Fatal(err)
			}
			err = subcommandTemplate.Funcs(templateFuncMap).Execute(targetFile, generator.SubcommandMetadata{
				Definition: metadata,
				Action:     meta,
				Name:       action,
			})
			if err != nil {
				log.Fatal(err)
			}
			err = targetFile.Close()
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Generate the AddCommands func, so the commands get added to the root command
	targetFile, err = os.Create("./cmd/addcommands" + generatedFileSuffix + ".go")
	if err != nil {
		log.Fatal(err)
	}
	commandList := []string{}
	for _, definitionFile := range definitionFiles {
		subcommandName := strings.TrimSuffix(definitionFile.Name(), filepath.Ext(definitionFile.Name()))
		commandList = append(commandList, subcommandName)
	}
	err = addcommandsTemplate.Funcs(templateFuncMap).Execute(targetFile, commandList)
	if err != nil {
		log.Fatal(err)
	}
	err = targetFile.Close()
	if err != nil {
		log.Fatal(err)
	}
}
