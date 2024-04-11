package generator

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"regexp"
)

type Action struct {
	Name        string
	Client      string
	APICall     APICall  `yaml:"api-call"`
	Params      []Param  `yaml:"params"`
	Description string   `yaml:"description"`
	RootKey     string   `yaml:"root-key"`
	Identifier  string   `yaml:"identifier"`
	Fields      []string `yaml:"fields"`
}

type Param struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"`
	Description string      `yaml:"description"`
	Required    bool        `yaml:"required"`
	Default     interface{} `yaml:"default"`
}

type APICall struct {
	Client   string
	Endpoint string
	Version  string
}

func (api *APICall) UnmarshalYAML(value *yaml.Node) error {
	regex := regexp.MustCompile(`([^.]+).(.+)`)

	matches := regex.FindStringSubmatch(value.Value)
	if len(matches) != 3 {
		return fmt.Errorf("invalid api call definition: %s", value.Value)
	}

	api.Client = matches[1]
	api.Endpoint = matches[2]
	api.Version = "v1"
	if len(matches) == 4 {
		versionRegex := regexp.MustCompile(`.*\/(v[0-9])`)

		versionMatches := versionRegex.FindStringSubmatch(api.Endpoint)
		if len(versionMatches) == 2 {
			api.Version = versionMatches[1]
			api.Endpoint = api.Endpoint[:len(api.Endpoint)-len(versionMatches[1])-1]
		}
	}

	return nil
}

type SubcommandDefinition struct {
	Name        string
	Actions     map[string]Action `yaml:"actions"`
	Identifier  string            `yaml:"identifier"`
	Description string            `yaml:"description"`
}

type SubcommandMetadata struct {
	Name       string
	Action     Action
	Definition SubcommandDefinition
}