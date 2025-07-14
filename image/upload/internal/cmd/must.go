/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package cmd

func must(err error) {
	if err != nil {
		panic(err)
	}
}
