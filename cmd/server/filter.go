package server

import (
	"buf.build/gen/go/gportal/gpcore/grpc/go/gpcore/api/admin/v1/adminv1grpc"
	adminv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/admin/v1"
	typesv1 "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/type/v1"
	"fmt"
	"strconv"
	"strings"

	"github.com/G-PORTAL/gpcore-cli/pkg/api"
	"github.com/G-PORTAL/gpcore-cli/pkg/client"
	"github.com/G-PORTAL/gpcore-cli/pkg/config"
	"github.com/G-PORTAL/gpcore-cli/pkg/protobuf"
	"github.com/charmbracelet/ssh"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// filter is implemented manually (not generated) because it calls ListServers
// with an ExtendedSearch payload, which contains a repeated Filter message that
// the YAML generator cannot express as command flags. Use "server
// search-options" to discover the valid filter names and values.

var filterSearch string
var filterFilters []string

// parseFilter converts a "name=value" string into a typesv1.Filter. Values are
// treated as strings by default. Prefix the value with "int:" or "bool:" to
// send an integer or boolean filter value respectively.
func parseFilter(raw string) (*typesv1.Filter, error) {
	parts := strings.SplitN(raw, "=", 2)
	if len(parts) != 2 || parts[0] == "" {
		return nil, fmt.Errorf("invalid filter %q, expected format name=value", raw)
	}
	name, value := parts[0], parts[1]

	builder := typesv1.Filter_builder{Name: name}
	switch {
	case strings.HasPrefix(value, "int:"):
		n, err := strconv.ParseInt(strings.TrimPrefix(value, "int:"), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer filter value for %q: %w", name, err)
		}
		builder.IntegerValue = &n
	case strings.HasPrefix(value, "bool:"):
		b, err := strconv.ParseBool(strings.TrimPrefix(value, "bool:"))
		if err != nil {
			return nil, fmt.Errorf("invalid boolean filter value for %q: %w", name, err)
		}
		builder.BooleanValue = &b
	default:
		builder.StringValue = &value
	}
	return builder.Build(), nil
}

var filterCmd = &cobra.Command{
	Args:                  cobra.OnlyValidArgs,
	DisableFlagsInUseLine: true,
	Long:                  "List servers using extended search filters. Use 'server search-options' to discover valid filter names and values.",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		ctx := client.ExtractContext(cobraCmd)
		grpcConn := ctx.Value("conn").(*grpc.ClientConn)
		grpcClient := adminv1grpc.NewAdminServiceClient(grpcConn)

		var extendedSearch *typesv1.SearchRequest
		if len(filterFilters) > 0 {
			filters := make([]*typesv1.Filter, 0, len(filterFilters))
			for _, raw := range filterFilters {
				f, err := parseFilter(raw)
				if err != nil {
					return err
				}
				filters = append(filters, f)
			}
			extendedSearch = typesv1.SearchRequest_builder{Filters: filters}.Build()
		}

		sshSession := ctx.Value("ssh").(*ssh.Session)
		tbl := table.NewWriter()
		tbl.SetStyle(table.StyleRounded)
		tbl.SetOutputMirror(*sshSession)
		cobraCmd.SetOut(*sshSession)
		defer cobraCmd.SetOut(nil)
		tbl.AppendHeader(table.Row{"Id", "Name", "InPool", "PowerState", "ProvisionState", "CreatedAt"})

		var combinedData []proto.Message
		var totalPages int32
		pagination := &typesv1.PaginationRequest{Page: 1}
		for {
			req := &adminv1.ListServersRequest{Pagination: pagination}
			if filterSearch != "" {
				req.Search = &filterSearch
			}
			if extendedSearch != nil {
				req.ExtendedSearch = extendedSearch
			}

			resp, err := grpcClient.ListServers(cobraCmd.Context(), req)
			if err != nil {
				return err
			}

			for _, entry := range resp.Servers {
				tbl.AppendRow(table.Row{
					fmt.Sprintf("%v", entry.Id),
					fmt.Sprintf("%v", entry.Name),
					api.FormatBoolean(entry.InPool),
					api.FormatServerPowerState(entry.PowerState),
					api.FormatServerProvisioningState(entry.ProvisionState),
					api.FormatDate(entry.CreatedAt),
				})
				combinedData = append(combinedData, entry)
			}

			if resp.Pagination == nil {
				break
			}
			totalPages = resp.GetPagination().GetTotal()
			pagination.Page++
			if resp.Pagination.Page >= totalPages {
				break
			}
		}

		if config.CSVOutput {
			tbl.RenderCSV()
			return nil
		}
		if !config.JSONOutput {
			tbl.Render()
		}
		if config.JSONOutput {
			jsonData, err := protobuf.MarshalIndent(combinedData)
			if err != nil {
				return err
			}
			cobraCmd.Println(string(jsonData))
		}
		return nil
	},
	Short:     "List servers using extended search filters",
	Use:       "filter",
	ValidArgs: []string{"search", "filter"},
}

func init() {
	filterCmd.Flags().StringVar(&filterSearch, "search", "", "Optional free-text search term")
	filterCmd.Flags().StringArrayVar(&filterFilters, "filter", nil, "Extended search filter in the form name=value (prefix value with int: or bool: to type it). Repeatable.")

	RootServerCommand.AddCommand(filterCmd)
}
