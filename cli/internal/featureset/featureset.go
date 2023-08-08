/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package featureset provides a way to check whether a feature is enabled in the current build.
// This package should not implement any logic itself, but only define constants that are set at build time.
package featureset

// Edition is the edition of a build.
type Edition int

const (
	// EditionOSS is the open-source software edition.
	EditionOSS Edition = iota
	// EditionEnterprise is the enterprise edition.
	EditionEnterprise
)

// CanFetchMeasurements returns whether the current build can fetch measurements.
const CanFetchMeasurements = canFetchMeasurements

// CanUseEmbeddedMeasurmentsAndImage returns whether the current build can use embedded measurements and can provide a node image.
const CanUseEmbeddedMeasurmentsAndImage = canUseEmbeddedMeasurmentsAndImage

// CanUpgradeCheck returns whether the current build can check for upgrades.
// This also includes fetching new measurements.
const CanUpgradeCheck = canUpgradeCheck

// CurrentEdition is the edition of the current build.
const CurrentEdition = edition
