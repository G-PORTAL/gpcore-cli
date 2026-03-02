package project

// This command is disabled because the ListSubnets endpoint is missing on
// gRPC (which is needed to resolve subnet IDs). It will be re-enabled once
// the endpoint is available.
//
// TODO: Re-enable once admin.ListSubnets is available in the gRPC API.
// The command should:
//   - Accept --subnet-ids and resolve them via ListSubnets
//   - Register itself with RootProjectCommand.AddCommand(networkCreateCmd)
//   - Fix MarkFlagRequired to use "subnet-ids" (not "subnets")
