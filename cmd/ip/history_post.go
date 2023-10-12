package ip

import (
	adminv1 "buf.build/gen/go/gportal/gportal-cloud/protocolbuffers/go/gpcloud/api/admin/v1"
	"github.com/spf13/cobra"
)

func HistoryHookPost(resp *adminv1.ListIPHistoriesResponse, cobraCmd *cobra.Command) ([]map[string]string, error) {
	res := make([]map[string]string, 0)
	for i := range resp.IpHistories {
		res = append(res, map[string]string{
			"User": resp.IpHistories[i].GetUser().GetUsername(),
		})
	}

	return res, nil
}
