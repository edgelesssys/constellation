/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package k8sapi

import (
	"flag"
	"fmt"

	"github.com/google/shlex"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func ParseJoinCommand(joinCommand string) (*kubeadm.BootstrapTokenDiscovery, error) {
	// Format:
	// kubeadm join [API_SERVER_ENDPOINT] --token [TOKEN] --discovery-token-ca-cert-hash [DISCOVERY_TOKEN_CA_CERT_HASH] --control-plane

	// split and verify that this is a kubeadm join command
	argv, err := shlex.Split(joinCommand)
	if err != nil {
		return nil, fmt.Errorf("kubadm join command could not be tokenized: %v", joinCommand)
	}
	if len(argv) < 3 {
		return nil, fmt.Errorf("kubadm join command is too short: %v", argv)
	}
	if argv[0] != "kubeadm" || argv[1] != "join" {
		return nil, fmt.Errorf("not a kubeadm join command: %v", argv)
	}

	result := kubeadm.BootstrapTokenDiscovery{APIServerEndpoint: argv[2]}

	var caCertHash string
	// parse flags
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.StringVar(&result.Token, "token", "", "")
	flags.StringVar(&caCertHash, "discovery-token-ca-cert-hash", "", "")
	flags.Bool("control-plane", false, "")
	flags.String("certificate-key", "", "")
	if err := flags.Parse(argv[3:]); err != nil {
		return nil, fmt.Errorf("parsing flag arguments: %v %w", argv, err)
	}

	if result.Token == "" {
		return nil, fmt.Errorf("missing flag argument token: %v", argv)
	}
	if caCertHash == "" {
		return nil, fmt.Errorf("missing flag argument discovery-token-ca-cert-hash: %v", argv)
	}
	result.CACertHashes = []string{caCertHash}

	return &result, nil
}
