/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"flag"
	"log"

	"github.com/edgelesssys/constellation/v2/terraform-provider-constellation/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// TODO(msanft): Set this accordingly in the release CI.
var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		// TODO(msanft): Verify that this will be the published name.
		Address: "registry.terraform.io/edgelesssys/constellation",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
