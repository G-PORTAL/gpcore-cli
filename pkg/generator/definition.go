package generator

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"regexp"
)

type Action struct {
	Name        string
	Client      string
	APICall     APICall           `yaml:"api-call"`
	Params      map[string]string `yaml:"params"`
	Description string            `yaml:"description"`
}

type APICall struct {
	Client   string
	Endpoint string
}

func (api *APICall) UnmarshalYAML(value *yaml.Node) error {
	regex := regexp.MustCompile(`([^.]+).(.+)`)

	matches := regex.FindStringSubmatch(value.Value)
	if len(matches) != 3 {
		return errors.New(fmt.Sprintf("invalid api call definition: %s", value.Value))
	}

	api.Client = matches[1]
	api.Endpoint = matches[2]

	return nil
}

type SubcommandDefinition struct {
	Name       string
	Actions    map[string]Action `yaml:"actions"`
	Identifier string            `yaml:"identifier"`
}

type SubcommandMetadata struct {
	Name       string
	Action     Action
	Definition SubcommandDefinition
}
