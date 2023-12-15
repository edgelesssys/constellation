/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import "github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"

// Values for the Cilium Helm releases for AWS.
var ciliumVals = map[string]map[string]any{
	cloudprovider.AWS.String(): {
		"endpointRoutes": map[string]any{
			"enabled": true,
		},
		"extraArgs": []string{"--node-encryption-opt-out-labels=invalid.label"},
		"encryption": map[string]any{
			"enabled":        true,
			"type":           "wireguard",
			"nodeEncryption": true,
			"strictMode": map[string]any{
				"enabled":                   true,
				"allowRemoteNodeIdentities": false,
				"podCIDRList":               []string{"10.244.0.0/16"},
			},
		},
		"l7Proxy": false,
		"ipam": map[string]any{
			"operator": map[string]any{
				"clusterPoolIPv4PodCIDRList": []string{
					"10.244.0.0/16",
				},
			},
		},
		"image": map[string]any{
			"repository": "ghcr.io/3u13r/cilium",
			"suffix":     "",
			"tag":        "v1.15.0-pre.2-edg.1",
			"digest":     "sha256:eebf631fd0f27e1f28f1fdeb2e049f2c83b887381466245c4b3e26440daefa27",
			"useDigest":  true,
		},
		"operator": map[string]any{
			"image": map[string]any{
				"repository":    "ghcr.io/3u13r/operator",
				"tag":           "v1.15.0-pre.2-edg.1",
				"suffix":        "",
				"genericDigest": "sha256:bfaeac2e05e8c38f439b0fbc36558fd8d11602997f2641423e8d86bd7ac6a88c",
				"useDigest":     true,
			},
		},
		"bpf": map[string]any{
			"masquerade": true,
		},
		"ipMasqAgent": map[string]any{
			"enabled": true,
			"config": map[string]any{
				"masqLinkLocal": true,
			},
		},
		"kubeProxyReplacement":                "strict",
		"enableCiliumEndpointSlice":           true,
		"kubeProxyReplacementHealthzBindAddr": "0.0.0.0:10256",
	},
	cloudprovider.Azure.String(): {
		"endpointRoutes": map[string]any{
			"enabled": true,
		},
		"extraArgs": []string{"--node-encryption-opt-out-labels=invalid.label"},
		"encryption": map[string]any{
			"enabled":        true,
			"type":           "wireguard",
			"nodeEncryption": true,
			"strictMode": map[string]any{
				"enabled":                   true,
				"allowRemoteNodeIdentities": false,
				"podCIDRList":               []string{"10.244.0.0/16"},
			},
		},
		"l7Proxy": false,
		"ipam": map[string]any{
			"operator": map[string]any{
				"clusterPoolIPv4PodCIDRList": []string{
					"10.244.0.0/16",
				},
			},
		},
		"image": map[string]any{
			"repository": "ghcr.io/3u13r/cilium",
			"suffix":     "",
			"tag":        "v1.15.0-pre.2-edg.1",
			"digest":     "sha256:eebf631fd0f27e1f28f1fdeb2e049f2c83b887381466245c4b3e26440daefa27",
			"useDigest":  true,
		},
		"operator": map[string]any{
			"image": map[string]any{
				"repository":    "ghcr.io/3u13r/operator",
				"tag":           "v1.15.0-pre.2-edg.1",
				"suffix":        "",
				"genericDigest": "sha256:bfaeac2e05e8c38f439b0fbc36558fd8d11602997f2641423e8d86bd7ac6a88c",
				"useDigest":     true,
			},
		},
		"bpf": map[string]any{
			"masquerade": true,
		},
		"ipMasqAgent": map[string]any{
			"enabled": true,
			"config": map[string]any{
				"masqLinkLocal": true,
			},
		},
		"kubeProxyReplacement":                "strict",
		"enableCiliumEndpointSlice":           true,
		"kubeProxyReplacementHealthzBindAddr": "0.0.0.0:10256",
	},
	cloudprovider.GCP.String(): {
		"endpointRoutes": map[string]any{
			"enabled": true,
		},
		"extraArgs": []string{"--node-encryption-opt-out-labels=invalid.label"},
		"tunnel":    "disabled",
		"encryption": map[string]any{
			"enabled":        true,
			"type":           "wireguard",
			"nodeEncryption": true,
			"strictMode": map[string]any{
				"enabled":                   true,
				"allowRemoteNodeIdentities": false,
			},
		},
		"image": map[string]any{
			"repository": "ghcr.io/3u13r/cilium",
			"suffix":     "",
			"tag":        "v1.15.0-pre.2-edg.1",
			"digest":     "sha256:eebf631fd0f27e1f28f1fdeb2e049f2c83b887381466245c4b3e26440daefa27",
			"useDigest":  true,
		},
		"operator": map[string]any{
			"image": map[string]any{
				"repository":    "ghcr.io/3u13r/operator",
				"suffix":        "",
				"tag":           "v1.15.0-pre.2-edg.1",
				"genericDigest": "sha256:bfaeac2e05e8c38f439b0fbc36558fd8d11602997f2641423e8d86bd7ac6a88c",
				"useDigest":     true,
			},
		},
		"l7Proxy": false,
		"ipam": map[string]any{
			"mode": "kubernetes",
		},
		"bpf": map[string]any{
			"masquerade": true,
		},
		"ipMasqAgent": map[string]any{
			"enabled": true,
			"config": map[string]any{
				"masqLinkLocal": true,
			},
		},
		"kubeProxyReplacement":                "strict",
		"enableCiliumEndpointSlice":           true,
		"kubeProxyReplacementHealthzBindAddr": "0.0.0.0:10256",
	},
	cloudprovider.OpenStack.String(): {
		"endpointRoutes": map[string]any{
			"enabled": true,
		},
		"extraArgs": []string{"--node-encryption-opt-out-labels=invalid.label"},
		"encryption": map[string]any{
			"enabled":        true,
			"type":           "wireguard",
			"nodeEncryption": true,
			"strictMode": map[string]any{
				"enabled":     true,
				"podCIDRList": []string{"10.244.0.0/16"},
			},
		},
		"l7Proxy": false,
		"ipam": map[string]any{
			"operator": map[string]any{
				"clusterPoolIPv4PodCIDRList": []string{
					"10.244.0.0/16",
				},
			},
		},
		"image": map[string]any{
			"repository": "ghcr.io/3u13r/cilium",
			"suffix":     "",
			"tag":        "v1.15.0-pre.2-edg.1",
			"digest":     "sha256:eebf631fd0f27e1f28f1fdeb2e049f2c83b887381466245c4b3e26440daefa27",
			"useDigest":  true,
		},
		"operator": map[string]any{
			"image": map[string]any{
				"repository":    "ghcr.io/3u13r/operator",
				"tag":           "v1.15.0-pre.2-edg.1",
				"suffix":        "",
				"genericDigest": "sha256:bfaeac2e05e8c38f439b0fbc36558fd8d11602997f2641423e8d86bd7ac6a88c",
				"useDigest":     true,
			},
		},
		"bpf": map[string]any{
			"masquerade": true,
		},
		"ipMasqAgent": map[string]any{
			"enabled": true,
			"config": map[string]any{
				"masqLinkLocal": true,
			},
		},
		"kubeProxyReplacement":                "strict",
		"enableCiliumEndpointSlice":           true,
		"kubeProxyReplacementHealthzBindAddr": "0.0.0.0:10256",
	},
	cloudprovider.QEMU.String(): {
		"endpointRoutes": map[string]any{
			"enabled": true,
		},
		"encryption": map[string]any{
			"enabled":        true,
			"type":           "wireguard",
			"nodeEncryption": true,
			"strictMode": map[string]any{
				"enabled":     true,
				"podCIDRList": []string{"10.244.0.0/16"},
			},
		},
		"image": map[string]any{
			"repository": "ghcr.io/3u13r/cilium",
			"suffix":     "",
			"tag":        "v1.15.0-pre.2-edg.1",
			"digest":     "sha256:eebf631fd0f27e1f28f1fdeb2e049f2c83b887381466245c4b3e26440daefa27",
			"useDigest":  true,
		},
		"operator": map[string]any{
			"image": map[string]any{
				"repository":    "ghcr.io/3u13r/operator",
				"suffix":        "",
				"tag":           "v1.15.0-pre.2-edg.1",
				"genericDigest": "sha256:bfaeac2e05e8c38f439b0fbc36558fd8d11602997f2641423e8d86bd7ac6a88c",
				"useDigest":     true,
			},
		},
		"ipam": map[string]any{
			"operator": map[string]any{
				"clusterPoolIPv4PodCIDRList": []string{
					"10.244.0.0/16",
				},
			},
		},
		"bpf": map[string]any{
			"masquerade": true,
		},
		"ipMasqAgent": map[string]any{
			"enabled": true,
			"config": map[string]any{
				"masqLinkLocal": true,
			},
		},
		"kubeProxyReplacement":                "strict",
		"enableCiliumEndpointSlice":           true,
		"kubeProxyReplacementHealthzBindAddr": "0.0.0.0:10256",
		"l7Proxy":                             false,
	},
}

var controlPlaneNodeSelector = map[string]any{"node-role.kubernetes.io/control-plane": ""}

var controlPlaneTolerations = []map[string]any{
	{
		"key":      "node-role.kubernetes.io/control-plane",
		"effect":   "NoSchedule",
		"operator": "Exists",
	},
	{
		"key":      "node-role.kubernetes.io/master",
		"effect":   "NoSchedule",
		"operator": "Exists",
	},
}
