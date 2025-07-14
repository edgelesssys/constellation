/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

// Package testdata contains testing data for an attestation process.
package testdata

import _ "embed"

// HCLReport is an example HCL report from an Azure TDX VM.
//
//go:embed hclreport.bin
var HCLReport []byte
