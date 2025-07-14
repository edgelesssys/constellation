//go:build enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package featureset

const (
	edition                           = EditionEnterprise
	canFetchMeasurements              = true
	canUpgradeCheck                   = true
	canUseEmbeddedMeasurmentsAndImage = true
)
