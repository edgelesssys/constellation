/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

func must(err error) {
	if err != nil {
		panic(err)
	}
}
