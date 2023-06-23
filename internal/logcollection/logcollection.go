/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package logcollection uses logstash and filebeat to collect logs centrally for debugging purposes.
package logcollection

import (
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/role"
)

const (
	openSearchHost = "https://search-e2e-logs-y46renozy42lcojbvrt3qq7csm.eu-central-1.es.amazonaws.com:443"
)

type LogMetadata struct {
	Provider string
	Name     string
	Role     role.Role
	VPCIP    string
	UID      string
}

func prepareInfoMap(m map[string]string, metadata LogMetadata) map[string]string {
	m = filterInfoMap(m)

	m["provider"] = metadata.Provider

	m["name"] = "unknown"
	if metadata.Name != "" {
		m["name"] = metadata.Name
	}

	m["role"] = metadata.Role.String()

	m["vpcip"] = "unknown"
	if metadata.VPCIP != "" {
		m["vpcip"] = metadata.VPCIP
	}

	m["uid"] = "unknown"
	if metadata.UID != "" {
		m["uid"] = metadata.UID
	}

	return m
}

func filterInfoMap(in map[string]string) map[string]string {
	out := make(map[string]string)

	for k, v := range in {
		if strings.HasPrefix(k, "logcollect.") {
			out[strings.TrimPrefix(k, "logcollect.")] = v
		}
	}

	delete(out, "logcollect")

	return out
}
