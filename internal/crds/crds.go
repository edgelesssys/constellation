/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package crds

import _ "embed"

var (
	//go:embed olmCRDs.yaml
	OLMCRDs []byte
	//go:embed olmDeployment.yaml
	OLM []byte
)
