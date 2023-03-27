/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

// Values for the Cilium Helm releases for AWS.
var awsVals = map[string]any{
	"endpointRoutes": map[string]any{
		"enabled": true,
	},
	"encryption": map[string]any{
		"enabled": true,
		"type":    "wireguard",
	},
	"hubble": map[string]any{
		"enabled": false,
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
}

// Values for the Cilium Helm releases for Azure.
var azureVals = map[string]any{
	"endpointRoutes": map[string]any{
		"enabled": true,
	},
	"encryption": map[string]any{
		"enabled": true,
		"type":    "wireguard",
	},
	"hubble": map[string]any{
		"enabled": false,
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
}

// Values for the Cilium Helm releases for GCP.
var gcpVals = map[string]any{
	"endpointRoutes": map[string]any{
		"enabled": true,
	},
	"tunnel": "disabled",
	"encryption": map[string]any{
		"enabled": true,
		"type":    "wireguard",
	},
	"hubble": map[string]any{
		"enabled": false,
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
}

// Values for the Cilium Helm releases for OpenStack.
var openStackVals = map[string]any{
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
}

var qemuVals = map[string]any{
	"endpointRoutes": map[string]any{
		"enabled": true,
	},
	"encryption": map[string]any{
		"enabled": true,
		"type":    "wireguard",
	},
	"hubble": map[string]any{
		"enabled": false,
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
}
