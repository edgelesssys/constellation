//go:build !enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package featureset

const (
	edition                           = EditionOSS
	canFetchMeasurements              = false
	canUpgradeCheck                   = false
	canUseEmbeddedMeasurmentsAndImage = false
)
