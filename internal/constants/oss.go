//go:build !enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constants

// CosignPublicKey signs all our development builds.
const CosignPublicKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAELcPl4Ik+qZuH4K049wksoXK/Os3Z
b92PDCpM7FZAINQF88s1TZS/HmRXYk62UJ4eqPduvUnJmXhNikhLbMi6fw==
-----END PUBLIC KEY-----
`

// VersionBuild is the category of the current build.
const VersionBuild = "Open-source software build; AGPL-3.0-only applies"
