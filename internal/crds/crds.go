package crds

import _ "embed"

var (
	//go:embed olmCRDs.yaml
	OLMCRDs []byte
	//go:embed olmDeployment.yaml
	OLM []byte
)
