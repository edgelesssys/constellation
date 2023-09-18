/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package metricbeat

import "embed"

// Assets are the exported Metricbeat template files.
//
//go:embed templates/*
var Assets embed.FS
