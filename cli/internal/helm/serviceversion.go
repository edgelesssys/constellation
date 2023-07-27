/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/semver"
)

// ServiceVersions bundles the versions of all services that are part of Constellation.
type ServiceVersions struct {
	cilium                 semver.Semver
	certManager            semver.Semver
	constellationOperators semver.Semver
	constellationServices  semver.Semver
	awsLBController        semver.Semver
	csiVersions            map[string]semver.Semver
}

// String returns a string representation of the ServiceVersions struct.
func (s ServiceVersions) String() string {
	builder := strings.Builder{}
	builder.WriteString("Service versions:\n")
	builder.WriteString(fmt.Sprintf("\tCilium: %s\n", s.cilium))
	builder.WriteString(fmt.Sprintf("\tcert-manager: %s\n", s.certManager))
	builder.WriteString(fmt.Sprintf("\tconstellation-operators: %s\n", s.constellationOperators))
	builder.WriteString(fmt.Sprintf("\tconstellation-services: %s\n", s.constellationServices))

	if s.awsLBController != (semver.Semver{}) {
		builder.WriteString(fmt.Sprintf("\taws-load-balancer-controller: %s\n", s.awsLBController))
	}

	builder.WriteString("\tCSI:")
	if len(s.csiVersions) != 0 {
		builder.WriteString("\n")
		for name, csiVersion := range s.csiVersions {
			builder.WriteString(fmt.Sprintf("\t\t%s: %s\n", name, csiVersion))
		}
	} else {
		builder.WriteString(" not installed\n")
	}

	return builder.String()
}

// ConstellationServices returns the version of the constellation-services release.
func (s ServiceVersions) ConstellationServices() semver.Semver {
	return s.constellationServices
}
