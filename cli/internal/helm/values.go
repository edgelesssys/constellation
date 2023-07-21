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
		"encryption": map[string]any{
			"enabled": true,
			"type":    "wireguard",
		},
		"l7Proxy": false,
		"ipam": map[string]any{
			"operator": map[string]any{
				"clusterPoolIPv4PodCIDRList": []string{
					"10.244.0.0/16",
				},
			},
		},
		"strictModeCIDR": "10.244.0.0/16",
		"image": map[string]any{
			"repository": "ghcr.io/3u13r/cilium",
			"suffix":     "",
			"tag":        "v1.12.1-edg",
			"digest":     "sha256:fdac430143fe719331698b76fbe66410631a21afd3405407d56db260d2d6999b",
			"useDigest":  true,
		},
		"operator": map[string]any{
			"image": map[string]any{
				"repository":    "ghcr.io/3u13r/operator",
				"tag":           "v1.12.1-edg",
				"suffix":        "",
				"genericDigest": "sha256:a225d8d3976fd2a05cfa0c929cd32e60283abedf6bae51db4709df19b2fb70cb",
				"useDigest":     true,
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
		"encryption": map[string]any{
			"enabled": true,
			"type":    "wireguard",
		},
		"l7Proxy": false,
		"ipam": map[string]any{
			"operator": map[string]any{
				"clusterPoolIPv4PodCIDRList": []string{
					"10.244.0.0/16",
				},
			},
		},
		"strictModeCIDR": "10.244.0.0/16",
		"image": map[string]any{
			"repository": "ghcr.io/3u13r/cilium",
			"suffix":     "",
			"tag":        "v1.12.1-edg",
			"digest":     "sha256:fdac430143fe719331698b76fbe66410631a21afd3405407d56db260d2d6999b",
			"useDigest":  true,
		},
		"operator": map[string]any{
			"image": map[string]any{
				"repository":    "ghcr.io/3u13r/operator",
				"tag":           "v1.12.1-edg",
				"suffix":        "",
				"genericDigest": "sha256:a225d8d3976fd2a05cfa0c929cd32e60283abedf6bae51db4709df19b2fb70cb",
				"useDigest":     true,
			},
		},
		"egressMasqueradeInterfaces":          "eth0",
		"enableIPv4Masquerade":                true,
		"kubeProxyReplacement":                "strict",
		"enableCiliumEndpointSlice":           true,
		"kubeProxyReplacementHealthzBindAddr": "0.0.0.0:10256",
	},
	cloudprovider.GCP.String(): {
		"endpointRoutes": map[string]any{
			"enabled": true,
		},
		"tunnel": "disabled",
		"encryption": map[string]any{
			"enabled": true,
			"type":    "wireguard",
		},
		"image": map[string]any{
			"repository": "ghcr.io/3u13r/cilium",
			"suffix":     "",
			"tag":        "v1.12.1-edg",
			"digest":     "sha256:fdac430143fe719331698b76fbe66410631a21afd3405407d56db260d2d6999b",
			"useDigest":  true,
		},
		"operator": map[string]any{
			"image": map[string]any{
				"repository":    "ghcr.io/3u13r/operator",
				"suffix":        "",
				"tag":           "v1.12.1-edg",
				"genericDigest": "sha256:a225d8d3976fd2a05cfa0c929cd32e60283abedf6bae51db4709df19b2fb70cb",
				"useDigest":     true,
			},
		},
		"l7Proxy": false,
		"ipam": map[string]any{
			"mode": "kubernetes",
		},
		"kubeProxyReplacement":                "strict",
		"enableCiliumEndpointSlice":           true,
		"kubeProxyReplacementHealthzBindAddr": "0.0.0.0:10256",
	},
	cloudprovider.OpenStack.String(): {
		"endpointRoutes": map[string]any{
			"enabled": true,
		},
		"encryption": map[string]any{
			"enabled": true,
			"type":    "wireguard",
		},
		"l7Proxy": false,
		"ipam": map[string]any{
			"operator": map[string]any{
				"clusterPoolIPv4PodCIDRList": []string{
					"10.244.0.0/16",
				},
			},
		},
		"strictModeCIDR": "10.244.0.0/16",
		"image": map[string]any{
			"repository": "ghcr.io/3u13r/cilium",
			"suffix":     "",
			"tag":        "v1.12.1-edg",
			"digest":     "sha256:fdac430143fe719331698b76fbe66410631a21afd3405407d56db260d2d6999b",
			"useDigest":  true,
		},
		"operator": map[string]any{
			"image": map[string]any{
				"repository":    "ghcr.io/3u13r/operator",
				"tag":           "v1.12.1-edg",
				"suffix":        "",
				"genericDigest": "sha256:a225d8d3976fd2a05cfa0c929cd32e60283abedf6bae51db4709df19b2fb70cb",
				"useDigest":     true,
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
			"enabled": true,
			"type":    "wireguard",
		},
		"image": map[string]any{
			"repository": "ghcr.io/3u13r/cilium",
			"suffix":     "",
			"tag":        "v1.12.1-edg",
			"digest":     "sha256:fdac430143fe719331698b76fbe66410631a21afd3405407d56db260d2d6999b",
			"useDigest":  true,
		},
		"operator": map[string]any{
			"image": map[string]any{
				"repository":    "ghcr.io/3u13r/operator",
				"suffix":        "",
				"tag":           "v1.12.1-edg",
				"genericDigest": "sha256:a225d8d3976fd2a05cfa0c929cd32e60283abedf6bae51db4709df19b2fb70cb",
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
