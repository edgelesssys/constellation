/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

var azureVals = map[string]interface{}{
	"endpointRoutes": map[string]interface{}{
		"enabled": true,
	},
	"encryption": map[string]interface{}{
		"enabled": true,
		"type":    "wireguard",
	},
	"l7Proxy": false,
	"ipam": map[string]interface{}{
		"operator": map[string]interface{}{
			"clusterPoolIPv4PodCIDRList": []string{
				"10.244.0.0/16",
			},
		},
	},
	"strictModeCIDR": "10.244.0.0/16",
	"image": map[string]interface{}{
		"repository": "ghcr.io/3u13r/cilium",
		"suffix":     "",
		"tag":        "v1.12.1-edg",
		"digest":     "sha256:fdac430143fe719331698b76fbe66410631a21afd3405407d56db260d2d6999b",
		"useDigest":  true,
	},
	"operator": map[string]interface{}{
		"image": map[string]interface{}{
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

var gcpVals = map[string]interface{}{
	"endpointRoutes": map[string]interface{}{
		"enabled": true,
	},
	"tunnel": "disabled",
	"encryption": map[string]interface{}{
		"enabled": true,
		"type":    "wireguard",
	},
	"image": map[string]interface{}{
		"repository": "ghcr.io/3u13r/cilium",
		"suffix":     "",
		"tag":        "v1.12.1-edg",
		"digest":     "sha256:fdac430143fe719331698b76fbe66410631a21afd3405407d56db260d2d6999b",
		"useDigest":  true,
	},
	"operator": map[string]interface{}{
		"image": map[string]interface{}{
			"repository":    "ghcr.io/3u13r/operator",
			"suffix":        "",
			"tag":           "v1.12.1-edg",
			"genericDigest": "sha256:a225d8d3976fd2a05cfa0c929cd32e60283abedf6bae51db4709df19b2fb70cb",
			"useDigest":     true,
		},
	},
	"l7Proxy": false,
	"ipam": map[string]interface{}{
		"mode": "kubernetes",
	},
	"kubeProxyReplacement":                "strict",
	"enableCiliumEndpointSlice":           true,
	"kubeProxyReplacementHealthzBindAddr": "0.0.0.0:10256",
}

var qemuVals = map[string]interface{}{
	"endpointRoutes": map[string]interface{}{
		"enabled": true,
	},
	"encryption": map[string]interface{}{
		"enabled": true,
		"type":    "wireguard",
	},
	"l7Proxy": false,
}
