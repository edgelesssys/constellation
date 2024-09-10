package main

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/utils/v2/openstack/clientconfig"
)

func main() {
	ctx := context.Background()
	clientOpts := &clientconfig.ClientOpts{
		AuthType: clientconfig.AuthV3Password,
		AuthInfo: &clientconfig.AuthInfo{
			AuthURL:        authURL,
			UserDomainName: userDomainName,
			Username:       username,
			Password:       password,
		},
	}

	serversClient, err := clientconfig.NewServiceClient(ctx, "compute", clientOpts)
	if err != nil {
		panic(err)
	}
	// serversClient.Microversion = microversion

	listServersOpts := servers.ListOpts{}
	pager := servers.List(serversClient, listServersOpts)
	fmt.Printf("Servers: %+v\n", pager)
}
