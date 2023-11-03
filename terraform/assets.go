/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import "embed"

// Assets are the exported Terraform template files.
//
//go:embed infrastructure/*
//go:embed infrastructure/*/.terraform.lock.hcl
//go:embed infrastructure/iam/*/.terraform.lock.hcl
var Assets embed.FS
