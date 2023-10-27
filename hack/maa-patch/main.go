/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/maa"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <attestation-url>\n", os.Args[0])
		os.Exit(1)
	}

	attestationURL := os.Args[1]
	if _, err := url.Parse(attestationURL); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid attestation URL: %s\n", err)
		os.Exit(1)
	}

	p := maa.NewAzurePolicyPatcher()
	if err := p.Patch(context.Background(), attestationURL); err != nil {
		panic(err)
	}
}
