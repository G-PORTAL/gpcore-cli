package generator

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"regexp"
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
	// Fallback, when set on an admin.* action, declares a semantically
	// equivalent user-facing (cloud.*) endpoint that is called at runtime
	// instead of the admin endpoint when the session has no admin credentials
	// (config.HasAdminConfig() == false). Only true equivalents may be wired
	// up here: the fallback must accept the same parameters (params marked
	// admin-only are rejected in user sessions) and return the same response
	// item type.
	Fallback *Fallback `yaml:"fallback"`
	// FallbackHint customizes the error shown when an admin.* action without
	// a Fallback is invoked in a user session (e.g. pointing to an existing
	// user-facing command like "flavour list-project").
	FallbackHint string `yaml:"fallback-hint"`
}

// Fallback describes the user-facing endpoint used instead of an admin.*
// api-call when the session has no admin credentials. RootKey and Fields
// default to the action's values when left empty.
type Fallback struct {
	APICall APICall  `yaml:"api-call"`
	RootKey string   `yaml:"root-key"`
	Fields  []string `yaml:"fields"`
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
	// AdminOnly marks a param that only the admin endpoint understands. On
	// actions with a Fallback, setting such a flag in a user session is a
	// runtime error (the fallback request has no matching field).
	AdminOnly bool `yaml:"admin-only"`
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
