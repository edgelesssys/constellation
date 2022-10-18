/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

// KMSConfig is the configuration needed to set up Constellation's key management service.
type KMSConfig struct {
	MasterSecret []byte
	Salt         []byte
}
