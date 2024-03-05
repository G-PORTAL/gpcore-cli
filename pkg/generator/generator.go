//go:build ignore

// WARNING: This is not ideal. Generating code from go templates worked well for
// simple stuff, but implementing complex things with conditional imports is a
// pain. I started to implement the code generation with the code generator lib
// Jennifer, which worked well, but takes some more time. So I decided to go with
// this solution for now. It's not ideal, but it works, and we can move on with
// this project. Adding more features to the code generator is not a good idea
// and the migration to Jennifer should be considered.

package main

import (
	"github.com/G-PORTAL/gpcore-cli/pkg/generator"
	"github.com/stoewer/go-strcase"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	generatedFileSuffix := "_gen"

	log.Println("Generate subcommands ...")

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

		subcommandName := strings.Replace(strings.TrimSuffix(definitionFile.Name(), filepath.Ext(definitionFile.Name())), "-", "_", -1)
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
			targetFilename := "./cmd/" + subcommandName + "/root" + generatedFileSuffix + ".go"
			err = generator.GenerateRootCommand(metadata, targetFilename)
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
			targetFilename := "./cmd/" + subcommandName + "/" + strcase.SnakeCase(action) + generatedFileSuffix + ".go"
			err = generator.GenerateSubCommand(generator.SubcommandMetadata{
				Definition: metadata,
				Action:     meta,
				Name:       action,
			}, targetFilename)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Generate the Helper functions file
	if _, err := os.Stat("./pkg/protobuf"); os.IsNotExist(err) {
		err = os.Mkdir("./pkg/protobuf", 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	targetFilenameHelpers := "./pkg/protobuf/helpers" + generatedFileSuffix + ".go"
	err = generator.GenerateHelpersFile(targetFilenameHelpers)
	if err != nil {
		log.Fatal(err)
	}

	// Generate the AddCommands func, so the commands get added to the root command
	targetFilename := "./cmd/addcommands" + generatedFileSuffix + ".go"
	commandList := []string{}
	for _, definitionFile := range definitionFiles {
		subcommandName := strings.TrimSuffix(definitionFile.Name(), filepath.Ext(definitionFile.Name()))
		commandList = append(commandList, subcommandName)
	}
	err = generator.GenerateAddCommands(commandList, targetFilename)
	if err != nil {
		log.Fatal(err)
	}
}
