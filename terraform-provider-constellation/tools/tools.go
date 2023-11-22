//go:build tools
// +build tools

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package tools

import (
	// Documentation generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
