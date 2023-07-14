/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package logstash

import "embed"

// Assets are the exported Logstash template files.
//
//go:embed config/*
//go:embed templates/*
var Assets embed.FS
