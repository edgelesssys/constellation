/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package components

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalComponents(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	legacyFormat := `[{"URL":"https://example.com/foo.tar.gz","Hash":"1234567890","InstallPath":"/foo","Extract":true}]`
	newFormat := `[{"url":"https://example.com/foo.tar.gz","hash":"1234567890","install_path":"/foo","extract":true}]`

	var fromLegacy Components
	require.NoError(json.Unmarshal([]byte(legacyFormat), &fromLegacy))

	var fromNew Components
	require.NoError(json.Unmarshal([]byte(newFormat), &fromNew))

	assert.Equal(fromLegacy, fromNew)
}
