/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package crds

import _ "embed"

var (
	// OLMCRDs contains olmCRDs.yaml from [OLM Release].
	//
	// [OLM Release]: https://github.com/operator-framework/operator-lifecycle-manager/releases
	//
	//go:embed olmCRDs.yaml
	OLMCRDs []byte
	// OLM contains olm.yaml from [OLM Release].
	//
	// [OLM Release]: https://github.com/operator-framework/operator-lifecycle-manager/releases
	//
	//go:embed olmDeployment.yaml
	OLM []byte
)
