/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package testdata contains testing data for an attestation process.
package testdata

import _ "embed"

// HCLReport is an example HCL report from an Azure TDX VM.
//
//go:embed hclreport.bin
var HCLReport []byte
