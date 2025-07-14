/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package metricbeat

import "embed"

// Assets are the exported Metricbeat template files.
//
//go:embed templates/*
var Assets embed.FS
