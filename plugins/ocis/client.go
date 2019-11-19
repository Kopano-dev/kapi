package main

import (
	gateway "github.com/cs3org/go-cs3apis/cs3/gateway/v0alpha"

	"google.golang.org/grpc"
)

func NewConn(endpoint string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func GetGatewayServiceClient(endpoint string) (gateway.GatewayServiceClient, error) {
	conn, err := NewConn(endpoint)
	if err != nil {
		return nil, err
	}

	return gateway.NewGatewayServiceClient(conn), nil
}

func (p *OcisPlugin) getClient() (gateway.GatewayServiceClient, error) {
	return GetGatewayServiceClient("ocis:9999")
}
