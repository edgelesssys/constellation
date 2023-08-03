/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
/*
Overrides contains helm values that are dynamically injected into the helm charts.
*/
package helm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/openstack"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// TODO(malt3): switch over to DNS name on AWS and Azure
// soon as every apiserver certificate of every control-plane node
// has the dns endpoint in its SAN list.
// extraCiliumValues extends the given values map by some values depending on user input.
// This extra step of separating the application of user input is necessary since service upgrades should
// reuse user input from the init step. However, we can't rely on reuse-values, because
// during upgrades we all values need to be set locally as they might have changed.
// Also, the charts are not rendered correctly without all of these values.
func extraCiliumValues(provider cloudprovider.Provider, conformanceMode bool, output terraform.ApplyOutput) map[string]any {
	extraVals := map[string]any{}
	if conformanceMode {
		extraVals["kubeProxyReplacementHealthzBindAddr"] = ""
		extraVals["kubeProxyReplacement"] = "partial"
		extraVals["sessionAffinity"] = true
		extraVals["cni"] = map[string]any{
			"chainingMode": "portmap",
		}
	}

	extraVals["k8sServiceHost"] = output.IP
	extraVals["k8sServicePort"] = constants.KubernetesPort
	if provider == cloudprovider.GCP {
		extraVals["ipv4NativeRoutingCIDR"] = output.GCP.IPCidrPod
		extraVals["strictModeCIDR"] = output.GCP.IPCidrPod
	}
	return extraVals
}

// extraConstellationServicesValues extends the given values map by some values depending on user input.
// Values set inside this function are only applied during init, not during upgrade.
func extraConstellationServicesValues(cfg *config.Config, masterSecret, salt []byte, uid, serviceAccURI string, output terraform.ApplyOutput,
) (map[string]any, error) {
	attestationConfigJSON, err := json.Marshal(cfg.GetAttestationConfig())
	if err != nil {
		return nil, fmt.Errorf("marshalling measurements: %w", err)
	}
	extraVals := map[string]any{}
	extraVals["join-service"] = map[string]any{
		"measurementSalt":    base64.StdEncoding.EncodeToString(salt),
		"attestationVariant": cfg.GetAttestationConfig().GetVariant().String(),
		"attestationConfig":  string(attestationConfigJSON),
	}
	extraVals["verification-service"] = map[string]any{
		"attestationVariant": cfg.GetAttestationConfig().GetVariant().String(),
		"loadBalancerIP":     output.IP,
	}
	extraVals["konnectivity"] = map[string]any{
		"loadBalancerIP": output.IP,
	}

	extraVals["key-service"] = map[string]any{
		"masterSecret": base64.StdEncoding.EncodeToString(masterSecret),
		"salt":         base64.StdEncoding.EncodeToString(salt),
	}
	switch cfg.GetProvider() {
	case cloudprovider.OpenStack:
		extraVals["openstack"] = map[string]any{
			"deployYawolLoadBalancer": cfg.DeployYawolLoadBalancer(),
		}
		if cfg.DeployYawolLoadBalancer() {
			extraVals["yawol-controller"] = map[string]any{
				"yawolOSSecretName": "yawolkey",
				// has to be larger than ~30s to account for slow OpenStack API calls.
				"openstackTimeout": "1m",
				"yawolFloatingID":  cfg.Provider.OpenStack.FloatingIPPoolID,
				"yawolFlavorID":    cfg.Provider.OpenStack.YawolFlavorID,
				"yawolImageID":     cfg.Provider.OpenStack.YawolImageID,
			}
		}
	case cloudprovider.GCP:
		serviceAccountKey, err := gcpshared.ServiceAccountKeyFromURI(serviceAccURI)
		if err != nil {
			return nil, fmt.Errorf("getting service account key: %w", err)
		}
		rawKey, err := json.Marshal(serviceAccountKey)
		if err != nil {
			return nil, fmt.Errorf("marshaling service account key: %w", err)
		}
		if output.GCP == nil {
			return nil, fmt.Errorf("no GCP output from Terraform")
		}
		extraVals["ccm"] = map[string]any{
			"GCP": map[string]any{
				"projectID":         output.GCP.ProjectID,
				"uid":               uid,
				"secretData":        string(rawKey),
				"subnetworkPodCIDR": output.GCP.IPCidrPod,
			},
		}
	case cloudprovider.Azure:
		if output.Azure == nil {
			return nil, fmt.Errorf("no Azure output from Terraform")
		}
		ccmConfig, err := getCCMConfig(*output.Azure, serviceAccURI)
		if err != nil {
			return nil, fmt.Errorf("getting Azure CCM config: %w", err)
		}
		extraVals["ccm"] = map[string]any{
			"Azure": map[string]any{
				"azureConfig": string(ccmConfig),
			},
		}
	}

	return extraVals, nil
}

// getCCMConfig returns the configuration needed for the Kubernetes Cloud Controller Manager on Azure.
func getCCMConfig(tfOutput terraform.AzureApplyOutput, serviceAccURI string) ([]byte, error) {
	creds, err := azureshared.ApplicationCredentialsFromURI(serviceAccURI)
	if err != nil {
		return nil, fmt.Errorf("getting service account key: %w", err)
	}
	useManagedIdentityExtension := creds.PreferredAuthMethod == azureshared.AuthMethodUserAssignedIdentity
	config := cloudConfig{
		Cloud:                       "AzurePublicCloud",
		TenantID:                    creds.TenantID,
		SubscriptionID:              tfOutput.SubscriptionID,
		ResourceGroup:               tfOutput.ResourceGroup,
		LoadBalancerSku:             "standard",
		SecurityGroupName:           tfOutput.NetworkSecurityGroupName,
		LoadBalancerName:            tfOutput.LoadBalancerName,
		UseInstanceMetadata:         true,
		VMType:                      "vmss",
		Location:                    creds.Location,
		UseManagedIdentityExtension: useManagedIdentityExtension,
		UserAssignedIdentityID:      tfOutput.UserAssignedIdentity,
	}

	return json.Marshal(config)
}

// extraOperatorValues returns the values for the constellation-operator chart.
func extraOperatorValues(uid string) map[string]any {
	return map[string]any{
		"constellation-operator": map[string]any{
			"constellationUID": uid,
		},
	}
}

// extraCSIValues returns the values for the csi chart.
func extraCSIValues(provider cloudprovider.Provider, serviceAccURI string) (map[string]any, error) {
	var csiVals map[string]any
	if provider == cloudprovider.OpenStack {
		creds, err := openstack.AccountKeyFromURI(serviceAccURI)
		if err != nil {
			return nil, err
		}
		cinderIni := creds.CloudINI().CinderCSIConfiguration()
		csiVals = map[string]any{
			"cinder-config": map[string]any{
				"secretData": cinderIni,
			},
		}
	}
	return csiVals, nil
}
