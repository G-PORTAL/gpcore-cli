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
	Optional    bool        `yaml:"optional"` // Proto field is a pointer type (oneof/optional)
	// Source, when set, makes the flag optional and falls back to a value from
	// the session when the flag is left empty. The only supported value is
	// "session.CurrentProject", which uses the project selected via
	// "project use". If neither the flag nor the session value is set, the
	// command errors out. This lets project-scoped commands omit --project-id
	// once a project has been selected.
	Source string `yaml:"source"`
}

// APICall maps a CLI action to a gRPC endpoint via the "api-call" field in the
// YAML definitions (e.g. "admin.ListServers", "cloudv2.ListNodes").
//
// Intentional API coverage gaps (do NOT add commands for these):
//   - payment.* credit-card RPCs (AddCreditCard, RemoveCreditCard,
//     ListCreditCards, ChangeDefaultCreditCard): credit cards are no longer
//     used, so these calls are not supported.
//   - payment.* plan-code RPCs (ListPlanCodes, GetDefaultPlanCode,
//     ChangeDefaultPlanCode): Lago is no longer used, so plan codes are not
//     supported.
//   - auth.* RPCs (CreateClient, ListClients, GetClient, UpdateClient,
//     DeleteClient, ResetClientSecret, Register, ResendConfirmEMail, GetUser):
//     OAuth client management is not needed in the CLI.
//   - admin.GetDashboard: not needed in the CLI.
//   - admin.CreateProjectNetwork: dropped. It requires resolving subnet IDs via
//     a ListSubnets endpoint that is not available on the gRPC API, and the
//     request carries nested struct fields the generator cannot express. The
//     previously disabled cmd/project/_network_create.go stub was removed.
//
// Internal / agent-plane services are also intentionally not surfaced in the
// CLI: network.v1.*, metadata.v1.*, gateway.v1.*, and cloud.v2.ReadinessCheck.
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
	Group         string            `yaml:"group"`
}

type SubcommandMetadata struct {
	Name       string
	Action     Action
	Definition SubcommandDefinition
}
