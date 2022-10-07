/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package constants contains the constants used by Constellation.
Constants should never be overwritable by command line flags or configuration files.
*/
package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCosignPublicKeyBase64(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFZjhGMWhwbXdFK1lDRlh6akd0YVFjckw2WFpWVApKbUVlNWlTTHZHMVN5UVNBZXc3V2RNS0Y2bzl0OGUyVEZ1Q2t6bE9oaGx3czJPSFdiaUZabkZXQ0Z3PT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg==", CosignPublicKeyBase64())
}
