/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCertificateMarshalJSON(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	jsonARK := fmt.Sprintf("\"%s\"", arkPEM)
	var cert Certificate
	require.NoError(json.Unmarshal([]byte(jsonARK), &cert))

	out, err := json.Marshal(cert)
	require.NoError(err)
	assert.JSONEq(jsonARK, string(out))
}

func TestCertificateMarshalYAML(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	yamlARK := fmt.Sprintf("\"%s\"", arkPEM)
	var cert Certificate
	require.NoError(yaml.Unmarshal([]byte(yamlARK), &cert))

	out, err := yaml.Marshal(cert)
	require.NoError(err)
	assert.YAMLEq(yamlARK, string(out))
}
