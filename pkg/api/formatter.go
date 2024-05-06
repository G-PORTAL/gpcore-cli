package api

import (
	adminv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/admin/v1"
	cloudv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/cloud/v1"
	typesv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/type/v1"
	"github.com/jedib0t/go-pretty/v6/text"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
)

var SpecialFormatters = []string{
	"Price",
	"Date",
	"DateTime",
	"Boolean",
	"Company",
	"BasicUser",
	"BasicProject",
	"Datacenter",
	"Flavour",
	"Image",
	"NetboxId",
	"Currency",
	"ProjectEnvironment",
	"SshKey",
	"SubnetList",
	"OperatingSystemFamily",
	"BillStatus",
	"DatacenterStatus",
	"FlavourAvailability",
	"ServerLogLevel",
	"ServerLogSource",
	"NetworkType",
	"AgentType",
	"SwitchType",
	"ServerProvisioningState",
	"ServerPowerState",
	"BillingPeriod",
}

func FormatPrice(price *typesv1.Price) string {
	return text.FgBlue.Sprint(price.GetFormatted())
}

func FormatDate(date *timestamppb.Timestamp) string {
	return text.FgCyan.Sprint(date.AsTime().Format("2006-01-02"))
}

func FormatDateTime(date *timestamppb.Timestamp) string {
	return text.FgCyan.Sprint(date.AsTime().Format(" 2006-01-02 15:04:05"))
}

func FormatBoolean(b bool) string {
	if b {
		return text.FgGreen.Sprintf("✔")
	}
	return text.FgRed.Sprintf("✖")
}

func FormatCompany(company *cloudv1.BillingProfileCompany) string {
	return company.GetName()
}

func FormatBasicUser(user *cloudv1.BasicUser) string {
	return user.GetFullName()
}

func FormatBasicProject(project *cloudv1.BasicProject) string {
	return project.GetName()
}

func FormatDatacenter(datacenter *cloudv1.Datacenter) string {
	return datacenter.GetName()
}

func FormatFlavour(flavour *cloudv1.Flavour) string {
	return flavour.GetName()
}

func FormatImage(image *cloudv1.Image) string {
	return image.GetName()
}

func FormatNetboxId(netboxId *int64) string {
	return text.FgCyan.Sprintf("%d", *netboxId)
}

func FormatSubnetList(subnets []*cloudv1.Subnet) string {
	var cidrs []string

	for _, subnet := range subnets {
		cidrs = append(cidrs, subnet.GetCidr().GetCidr())
	}

	return strings.Join(cidrs, ", ")
}

func FormatSshKey(sshKey *typesv1.SSHKey) string {
	return sshKey.GetPublicKey()
}

func FormatOperatingSystemFamily(family cloudv1.OperatingSystemFamily) string {
	return text.FgYellow.Sprintf(strings.TrimPrefix(family.String(), "OPERATING_SYSTEM_FAMILY_"))
}

func FormatBillStatus(status cloudv1.BillStatus) string {
	s := strings.TrimPrefix(status.String(), "BILL_STATUS_")
	switch s {
	case "FAILED":
		return text.FgRed.Sprintf(s)
	case "SUCCEEDED":
		return text.FgGreen.Sprintf(s)
	case "PENDING":
		return text.FgYellow.Sprintf(s)
	}

	return s
}
func FormatDatacenterStatus(status cloudv1.DatacenterStatus) string {
	s := strings.TrimPrefix(status.String(), "DATACENTER_STATUS_")
	switch s {
	case "AVAILABLE":
		return text.FgGreen.Sprintf(s)
	case "UNAVAILABLE":
		return text.FgRed.Sprintf(s)
	case "COMMING_SOON":
		return text.FgYellow.Sprintf(s)
	default:
		return s
	}
}

func FormatFlavourAvailability(flavour cloudv1.FlavourAvailability) string {
	f := strings.TrimPrefix(flavour.String(), "FLAVOUR_AVAILABILITY_")
	switch f {
	case "HIGH":
		return text.FgGreen.Sprintf(f)
	case "MIDDLE":
		return text.FgYellow.Sprintf(f)
	case "LOW":
		return text.FgRed.Sprintf(f)
	case "PREORDER":
		return text.FgCyan.Sprintf(f)
	case "OUT_OF_STOCK":
		return text.FgRed.Sprintf(f)
	default:
		return "?"
	}
}

func FormatServerLogLevel(level cloudv1.ServerLogLevelType) string {
	l := strings.TrimPrefix(level.String(), "SERVER_LOG_LEVEL_TYPE_")
	switch l {
	case "DEBUG":
		return text.FgCyan.Sprintf(l)
	case "INFO":
		return text.FgGreen.Sprintf(l)
	case "WARNING":
		return text.FgYellow.Sprintf(l)
	case "ERROR":
		return text.FgRed.Sprintf(l)
	default:
		return l
	}
}

func FormatServerLogSource(source cloudv1.ServerLogSourceType) string {
	s := strings.TrimPrefix(source.String(), "SERVER_LOG_SOURCE_TYPE_")
	switch s {
	case "INTERNAL":
		return text.FgCyan.Sprintf(s)
	case "IRONIC":
		return text.FgYellow.Sprintf(s)
	case "NETBOX":
		return text.FgBlue.Sprintf(s)
	default:
		return s
	}
}

func FormatNetworkType(network cloudv1.NetworkType) string {
	n := strings.TrimPrefix(network.String(), "NETWORK_TYPE_")
	switch n {
	case "PUBLIC":
		return text.FgCyan.Sprintf(n)
	case "MANAGEMENT":
		return text.FgGreen.Sprintf(n)
	default:
		return n
	}
}

func FormatSwitchType(switchType cloudv1.SwitchType) string {
	return strings.TrimPrefix(switchType.String(), "SWITCH_TYPE_")
}

func FormatAgentType(agentType adminv1.AgentType) string {
	return strings.TrimPrefix(agentType.String(), "AGENT_TYPE_")
}

func FormatCurrency(currency cloudv1.Currency) string {
	return strings.TrimPrefix(currency.String(), "CURRENCY_")
}

func FormatProjectEnvironment(env cloudv1.ProjectEnvironment) string {
	s := strings.TrimPrefix(env.String(), "PROJECT_ENVIRONMENT_")
	switch s {
	case "PRODUCTION":
		return text.FgRed.Sprintf(s)
	case "STAGING":
		return text.FgYellow.Sprintf(s)
	case "DEVELOPMENT":
		return text.FgGreen.Sprintf(s)
	default:
		return s
	}
}

func FormatServerProvisioningState(state cloudv1.ServerProvisioningState) string {
	s := strings.TrimPrefix(state.String(), "SERVER_PROVISIONING_STATE_")
	switch s {
	case "ACTIVE":
		return text.FgGreen.Sprintf(s)
	default:
		return s
	}
}

func FormatBillingPeriod(period cloudv1.BillingPeriod) string {
	return strings.TrimPrefix(period.String(), "BILLING_PERIOD_")
}

func FormatServerPowerState(state cloudv1.ServerPowerState) string {
	return strings.TrimPrefix(state.String(), "SERVER_POWER_STATE_")
	// TODO: More colors here
}
