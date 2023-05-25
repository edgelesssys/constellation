/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versionsapi

var (
	// APIV1 is the v1 API version.
	APIV1 = apiVersion{slug: "v1"}
	// APIV2 is the v2 API version.
	APIV2 = apiVersion{slug: "v2"}
)

type apiVersion struct {
	slug string
}

func (v apiVersion) String() string {
	return v.slug
}
