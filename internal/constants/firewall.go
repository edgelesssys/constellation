/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package constants

import (
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudtypes"
)

var (
	// IngressRulesNoDebug is the default set of ingress rules for a Constellation cluster without debug mode.
	IngressRulesNoDebug = cloudtypes.Firewall{
		{
			Name:        "bootstrapper",
			Description: "bootstrapper default port",
			Protocol:    "tcp",
			IPRange:     "0.0.0.0/0",
			FromPort:    BootstrapperPort,
		},
		{
			Name:        "ssh",
			Description: "SSH",
			Protocol:    "tcp",
			IPRange:     "0.0.0.0/0",
			FromPort:    SSHPort,
		},
		{
			Name:        "nodeport",
			Description: "NodePort",
			Protocol:    "tcp",
			IPRange:     "0.0.0.0/0",
			FromPort:    NodePortFrom,
			ToPort:      NodePortTo,
		},
		{
			Name:        "kubernetes",
			Description: "Kubernetes",
			Protocol:    "tcp",
			IPRange:     "0.0.0.0/0",
			FromPort:    KubernetesPort,
		},
		{
			Name:        "konnectivity",
			Description: "konnectivity",
			Protocol:    "tcp",
			IPRange:     "0.0.0.0/0",
			FromPort:    KonnectivityPort,
		},
		{
			Name:        "recovery",
			Description: "control-plane recovery",
			Protocol:    "tcp",
			IPRange:     "0.0.0.0/0",
			FromPort:    RecoveryPort,
		},
	}

	// IngressRulesDebug is the default set of ingress rules for a Constellation cluster with debug mode.
	IngressRulesDebug = append(IngressRulesNoDebug, cloudtypes.Firewall{
		{
			Name:        "debugd",
			Description: "debugd",
			Protocol:    "tcp",
			IPRange:     "0.0.0.0/0",
			FromPort:    DebugdPort,
		},
	}...)

	// EgressRules is the default set of egress rules for a Constellation cluster.
	EgressRules = cloudtypes.Firewall{}
)
