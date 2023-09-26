/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package filebeat

import "embed"

// Assets are the exported Filebeat template files.
//
//go:embed templates/*
var Assets embed.FS
