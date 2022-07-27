package azure

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/edgelesssys/constellation/internal/constants"
)

// LoadBalancer defines a Azure load balancer.
type LoadBalancer struct {
	Name          string
	Subscription  string
	ResourceGroup string
	Location      string
	PublicIPID    string
	UID           string
}

const (
	BackendAddressPoolWorkerName       = "backendAddressWorkerPool"
	BackendAddressPoolControlPlaneName = "backendAddressControlPlanePool"
)

// Azure returns a Azure representation of LoadBalancer.
func (l LoadBalancer) Azure() armnetwork.LoadBalancer {
	frontEndIPConfigName := "frontEndIPConfig"
	kubeHealthProbeName := "kubeHealthProbe"
	verifyHealthProbeName := "verifyHealthProbe"
	coordHealthProbeName := "coordHealthProbe"
	debugdHealthProbeName := "debugdHealthProbe"
	backEndAddressPoolNodeName := BackendAddressPoolWorkerName + "-" + l.UID
	backEndAddressPoolControlPlaneName := BackendAddressPoolControlPlaneName + "-" + l.UID

	return armnetwork.LoadBalancer{
		Name:     to.Ptr(l.Name),
		Location: to.Ptr(l.Location),
		SKU:      &armnetwork.LoadBalancerSKU{Name: to.Ptr(armnetwork.LoadBalancerSKUNameStandard)},
		Properties: &armnetwork.LoadBalancerPropertiesFormat{
			FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
				{
					Name: to.Ptr(frontEndIPConfigName),
					Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
						PublicIPAddress: &armnetwork.PublicIPAddress{
							ID: to.Ptr(l.PublicIPID),
						},
					},
				},
			},
			BackendAddressPools: []*armnetwork.BackendAddressPool{
				{
					Name: to.Ptr(backEndAddressPoolNodeName),
				},
				{
					Name: to.Ptr(backEndAddressPoolControlPlaneName),
				},
				{
					Name: to.Ptr("all"),
				},
			},
			Probes: []*armnetwork.Probe{
				{
					Name: to.Ptr(kubeHealthProbeName),
					Properties: &armnetwork.ProbePropertiesFormat{
						Protocol: to.Ptr(armnetwork.ProbeProtocolTCP),
						Port:     to.Ptr(int32(6443)),
					},
				},
				{
					Name: to.Ptr(verifyHealthProbeName),
					Properties: &armnetwork.ProbePropertiesFormat{
						Protocol: to.Ptr(armnetwork.ProbeProtocolTCP),
						Port:     to.Ptr[int32](constants.VerifyServiceNodePortGRPC),
					},
				},
				{
					Name: to.Ptr(coordHealthProbeName),
					Properties: &armnetwork.ProbePropertiesFormat{
						Protocol: to.Ptr(armnetwork.ProbeProtocolTCP),
						Port:     to.Ptr[int32](constants.BootstrapperPort),
					},
				},
				{
					Name: to.Ptr(debugdHealthProbeName),
					Properties: &armnetwork.ProbePropertiesFormat{
						Protocol: to.Ptr(armnetwork.ProbeProtocolTCP),
						Port:     to.Ptr[int32](4000),
					},
				},
			},
			LoadBalancingRules: []*armnetwork.LoadBalancingRule{
				{
					Name: to.Ptr("kubeLoadBalancerRule"),
					Properties: &armnetwork.LoadBalancingRulePropertiesFormat{
						FrontendIPConfiguration: &armnetwork.SubResource{
							ID: to.Ptr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/frontendIPConfigurations/" + frontEndIPConfigName),
						},
						FrontendPort: to.Ptr[int32](6443),
						BackendPort:  to.Ptr[int32](6443),
						Protocol:     to.Ptr(armnetwork.TransportProtocolTCP),
						Probe: &armnetwork.SubResource{
							ID: to.Ptr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/probes/" + kubeHealthProbeName),
						},
						DisableOutboundSnat: to.Ptr(true),
						BackendAddressPools: []*armnetwork.SubResource{
							{
								ID: to.Ptr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/backendAddressPools/" + backEndAddressPoolControlPlaneName),
							},
						},
					},
				},
				{
					Name: to.Ptr("verifyLoadBalancerRule"),
					Properties: &armnetwork.LoadBalancingRulePropertiesFormat{
						FrontendIPConfiguration: &armnetwork.SubResource{
							ID: to.Ptr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/frontendIPConfigurations/" + frontEndIPConfigName),
						},
						FrontendPort: to.Ptr[int32](constants.VerifyServiceNodePortGRPC),
						BackendPort:  to.Ptr[int32](constants.VerifyServiceNodePortGRPC),
						Protocol:     to.Ptr(armnetwork.TransportProtocolTCP),
						Probe: &armnetwork.SubResource{
							ID: to.Ptr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/probes/" + verifyHealthProbeName),
						},
						DisableOutboundSnat: to.Ptr(true),
						BackendAddressPools: []*armnetwork.SubResource{
							{
								ID: to.Ptr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/backendAddressPools/" + backEndAddressPoolControlPlaneName),
							},
						},
					},
				},
				{
					Name: to.Ptr("coordLoadBalancerRule"),
					Properties: &armnetwork.LoadBalancingRulePropertiesFormat{
						FrontendIPConfiguration: &armnetwork.SubResource{
							ID: to.Ptr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/frontendIPConfigurations/" + frontEndIPConfigName),
						},
						FrontendPort: to.Ptr[int32](constants.BootstrapperPort),
						BackendPort:  to.Ptr[int32](constants.BootstrapperPort),
						Protocol:     to.Ptr(armnetwork.TransportProtocolTCP),
						Probe: &armnetwork.SubResource{
							ID: to.Ptr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/probes/" + coordHealthProbeName),
						},
						DisableOutboundSnat: to.Ptr(true),
						BackendAddressPools: []*armnetwork.SubResource{
							{
								ID: to.Ptr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/backendAddressPools/" + backEndAddressPoolControlPlaneName),
							},
						},
					},
				},
				{
					Name: to.Ptr("debudLoadBalancerRule"),
					Properties: &armnetwork.LoadBalancingRulePropertiesFormat{
						FrontendIPConfiguration: &armnetwork.SubResource{
							ID: to.Ptr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/frontendIPConfigurations/" + frontEndIPConfigName),
						},
						FrontendPort: to.Ptr[int32](4000),
						BackendPort:  to.Ptr[int32](4000),
						Protocol:     to.Ptr(armnetwork.TransportProtocolTCP),
						Probe: &armnetwork.SubResource{
							ID: to.Ptr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/probes/" + debugdHealthProbeName),
						},
						DisableOutboundSnat: to.Ptr(true),
						BackendAddressPools: []*armnetwork.SubResource{
							{
								ID: to.Ptr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/backendAddressPools/" + backEndAddressPoolControlPlaneName),
							},
						},
					},
				},
			},
			OutboundRules: []*armnetwork.OutboundRule{
				{
					Name: to.Ptr("outboundRuleControlPlane"),
					Properties: &armnetwork.OutboundRulePropertiesFormat{
						FrontendIPConfigurations: []*armnetwork.SubResource{
							{
								ID: to.Ptr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/frontendIPConfigurations/" + frontEndIPConfigName),
							},
						},
						BackendAddressPool: &armnetwork.SubResource{
							ID: to.Ptr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/backendAddressPools/all"),
						},
						Protocol: to.Ptr(armnetwork.LoadBalancerOutboundRuleProtocolAll),
					},
				},
			},
		},
	}
}
