/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

type helmLoader interface {
	Load(csp string, conformanceMode bool) ([]byte, error)
}

type stubHelmLoader struct {
	loadErr error
}

func (d *stubHelmLoader) Load(csp string, conformanceMode bool) ([]byte, error) {
	return nil, d.loadErr
}
