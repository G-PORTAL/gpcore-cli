package api

import (
	"crypto/tls"
	"fmt"
	"github.com/G-PORTAL/gpcore-go/pkg/gpcore/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
)

// NewGRPCConnection creates a new gRPC connection. We can not use the NewClient
// function from the client package, because we need the grpc.ClientConn object
// in order to open a new AdminClient. If we decide to expose the AdminClient
// in the client package, we can use the NewClient function and this function
// gets obsolete.
func NewGRPCConnection(extraOptions ...interface{}) (*grpc.ClientConn, error) {
	var options []grpc.DialOption
	// Certificate pinning
	options = append(options, grpc.WithTransportCredentials(
		credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
		})))

	// User Agent
	options = append(options, grpc.WithUserAgent(fmt.Sprintf("GPCORE CLI [%s]", grpc.Version)))
	endpoint := client.DefaultEndpoint

	for _, option := range extraOptions {
		if opt, ok := option.(grpc.DialOption); ok {
			options = append(options, opt)
			continue
		}
		if opt, ok := option.(client.EndpointOverrideOption); ok {
			endpoint = string(opt)
			continue
		}
		if opt, ok := option.(client.AuthProviderOption); ok {
			options = append(options, grpc.WithPerRPCCredentials(&client.AuthOption{
				Provider: &opt,
			}))
			continue
		}
		log.Printf("Unknown option type: %T", option)
	}

	clientConn, err := grpc.NewClient(endpoint, options...)
	if err != nil {
		return nil, err
	}

	return clientConn, nil
}
