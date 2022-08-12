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
	"strictModeCIDRs": []string{
		"10.244.0.0/16",
	},
	"image": map[string]interface{}{
		"repository": "ghcr.io/3u13r/cilium",
		"suffix":     "v1.12.0-edg2",
		"tag":        "latest",
		"digest":     "sha256:8dee8839bdf4cfdc28a61c4586f23f2dbfabe03f94dee787c4d749cfcc02c6bf",
		"useDigest":  false,
	},
	"operator": map[string]interface{}{
		"image": map[string]interface{}{
			"repository":    "ghcr.io/3u13r/operator",
			"tag":           "v1.12.0-edg2",
			"suffix":        "",
			"genericDigest": "sha256:adbdeb0199aa1d870940c3363bfa5b69a5c8b4f533fc9f67463f8d447077464a",
			"useDigest":     true,
		},
	},
	"egressMasqueradeInterfaces": "eth0",
	"enableIPv4Masquerade":       true,
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
		"tag":        "v1.12.0-edg2",
		"digest":     "sha256:8dee8839bdf4cfdc28a61c4586f23f2dbfabe03f94dee787c4d749cfcc02c6bf",
		"useDigest":  true,
	},
	"operator": map[string]interface{}{
		"image": map[string]interface{}{
			"repository":    "ghcr.io/3u13r/operator",
			"suffix":        "",
			"tag":           "v1.12.0-edg2",
			"genericDigest": "sha256:adbdeb0199aa1d870940c3363bfa5b69a5c8b4f533fc9f67463f8d447077464a",
			"useDigest":     true,
		},
	},
	"l7Proxy": false,
	"ipam": map[string]interface{}{
		"mode": "kubernetes",
	},
}

var qemuVals = map[string]interface{}{
	"endpointRoutes": map[string]interface{}{
		"enabled": true,
	},
	"encryption": map[string]interface{}{
		"enabled": true,
		"type":    "wireguard",
	},
}
