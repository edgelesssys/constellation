/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package logstash

import "embed"

// Assets are the exported Logstash template files.
//
//go:embed config/*
//go:embed templates/*
var Assets embed.FS
