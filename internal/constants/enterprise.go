//go:build enterprise

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constants

// CosignPublicKey signs all our releases.
const CosignPublicKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEf8F1hpmwE+YCFXzjGtaQcrL6XZVT
JmEe5iSLvG1SyQSAew7WdMKF6o9t8e2TFuCkzlOhhlws2OHWbiFZnFWCFw==
-----END PUBLIC KEY-----
`

// VersionBuild is the category of the current build.
const VersionBuild = "Enterprise build; see documentation for license agreement"

// Edition is the edition of the CLI.
const Edition = "Enterprise"
