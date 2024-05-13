//go:build linux

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/edgelesssys/constellation/v2/internal/role"
	"k8s.io/kubernetes/pkg/util/iptables"
	"k8s.io/utils/exec"
)

// PrepareControlPlaneNode sets up iptables for the control plane node only
// if an internal load balancer is used.
//
// This is needed since during `kubeadm init` the API server must talk to the
// kubeAPIEndpoint, which is the load balancer IP address. During that time, the
// only healthy VM is the VM itself. Therefore, traffic is sent to the load balancer
// and the 5-tuple is (VM IP, <some port>, LB IP, 6443, TCP).
// Now the load balancer does not re-write the source IP address only the destination (DNAT).
// Therefore the 5-tuple is (VM IP, <some port>, VM IP, 6443, TCP).
// Now the VM responds to the SYN packet with a SYN-ACK packet, but the outgoing
// connection waits on a response from the load balancer and not the VM therefore
// dropping the packet.
//
// OpenShift also uses the same mechanism to redirect traffic to the API server:
// https://github.com/openshift/machine-config-operator/blob/e453bd20bac0e48afa74e9a27665abaf454d93cd/templates/master/00-master/azure/files/opt-libexec-openshift-azure-routes-sh.yaml
func (c *Cloud) PrepareControlPlaneNode(ctx context.Context, log *slog.Logger) error {
	selfMetadata, err := c.Self(ctx)
	if err != nil {
		return fmt.Errorf("failed to get self metadata: %w", err)
	}

	// skipping iptables setup for worker nodes
	if selfMetadata.Role != role.ControlPlane {
		log.Info("not a control plane node, skipping iptables setup")
		return nil
	}

	// skipping iptables setup if no internal LB exists e.g.
	// for public LB architectures
	loadbalancerIP, err := c.getLoadBalancerPrivateIP(ctx)
	if err != nil {
		log.With(slog.Any("error", err)).Warn("skipping iptables setup, failed to get load balancer private IP")
		return nil
	}

	log.Info(fmt.Sprintf("Setting up iptables for control plane node with load balancer IP %s", loadbalancerIP))
	iptablesExec := iptables.New(exec.New(), iptables.ProtocolIPv4)

	const chainName = "azure-lb-nat"
	if _, err := iptablesExec.EnsureChain(iptables.TableNAT, chainName); err != nil {
		return fmt.Errorf("failed to create iptables chain: %w", err)
	}

	if _, err := iptablesExec.EnsureRule(iptables.Append, iptables.TableNAT, "PREROUTING", "-j", chainName); err != nil {
		return fmt.Errorf("failed to add rule to iptables chain: %w", err)
	}

	if _, err := iptablesExec.EnsureRule(iptables.Append, iptables.TableNAT, "OUTPUT", "-j", chainName); err != nil {
		return fmt.Errorf("failed to add rule to iptables chain: %w", err)
	}

	if _, err := iptablesExec.EnsureRule(iptables.Append, iptables.TableNAT, chainName, "--dst", loadbalancerIP, "-p", "tcp", "--dport", "6443", "-j", "REDIRECT"); err != nil {
		return fmt.Errorf("failed to add rule to iptables chain: %w", err)
	}

	return nil
}
