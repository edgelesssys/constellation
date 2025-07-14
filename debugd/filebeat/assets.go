/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package filebeat

import "embed"

// Assets are the exported Filebeat template files.
//
//go:embed templates/*
var Assets embed.FS
