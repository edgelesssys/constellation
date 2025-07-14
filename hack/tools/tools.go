//go:build tools

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

// The tools module is used to keep tool dependencies separate from the main dependencies of the repo
// For more details see: https://github.com/golang/go/issues/25922#issuecomment-1038394599
package main

import (
	_ "github.com/google/go-licenses"
	_ "github.com/google/keep-sorted"
	_ "github.com/katexochen/sh/v3/cmd/shfmt"
	_ "golang.org/x/tools/cmd/stringer"
	_ "golang.org/x/vuln/cmd/govulncheck"
)
