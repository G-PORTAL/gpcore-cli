package generator

import (
	"fmt"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"gopkg.in/yaml.v3"
	"regexp"
	"strings"
)

type Action struct {
	Name          string
	Client        string
	APICall       APICall  `yaml:"api-call"`
	Params        []Param  `yaml:"params"`
	Description   string   `yaml:"description"`
	RootKey       string   `yaml:"root-key"`
	Identifier    string   `yaml:"identifier"`
	IdentifierKey string   `yaml:"identifier-key"`
	Fields        []string `yaml:"fields"`
	NoPagination  bool     `yaml:"no-pagination"`
}

func (action *Action) CanCall() bool {
	adminCall := strings.HasPrefix(action.APICall.Client, "admin")
	return !adminCall || config.HasAdminConfig()
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

	versionRegex := regexp.MustCompile(`(.*)(v[1-9]+)`)
	versionMatches := versionRegex.FindStringSubmatch(api.Client)
	if len(versionMatches) == 3 {
		api.Client = versionMatches[1]
		api.Version = versionMatches[2]
	}

	return nil
}

type SubcommandDefinition struct {
	Name          string
	Actions       map[string]Action `yaml:"actions"`
	Identifier    string            `yaml:"identifier"`
	IdentifierKey string            `yaml:"identifier-key"`
	Description   string            `yaml:"description"`
}

type SubcommandMetadata struct {
	Name       string
	Action     Action
	Definition SubcommandDefinition
}
