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
	coordHealthProbeName := "coordHealthProbe"
	debugdHealthProbeName := "debugdHealthProbe"
	backEndAddressPoolNodeName := BackendAddressPoolWorkerName + "-" + l.UID
	backEndAddressPoolControlPlaneName := BackendAddressPoolControlPlaneName + "-" + l.UID

	return armnetwork.LoadBalancer{
		Name:     to.StringPtr(l.Name),
		Location: to.StringPtr(l.Location),
		SKU:      &armnetwork.LoadBalancerSKU{Name: armnetwork.LoadBalancerSKUNameStandard.ToPtr()},
		Properties: &armnetwork.LoadBalancerPropertiesFormat{
			FrontendIPConfigurations: []*armnetwork.FrontendIPConfiguration{
				{
					Name: to.StringPtr(frontEndIPConfigName),
					Properties: &armnetwork.FrontendIPConfigurationPropertiesFormat{
						PublicIPAddress: &armnetwork.PublicIPAddress{
							ID: to.StringPtr(l.PublicIPID),
						},
					},
				},
			},
			BackendAddressPools: []*armnetwork.BackendAddressPool{
				{
					Name: to.StringPtr(backEndAddressPoolNodeName),
				},
				{
					Name: to.StringPtr(backEndAddressPoolControlPlaneName),
				},
				{
					Name: to.StringPtr("all"),
				},
			},
			Probes: []*armnetwork.Probe{
				{
					Name: to.StringPtr(kubeHealthProbeName),
					Properties: &armnetwork.ProbePropertiesFormat{
						Protocol: armnetwork.ProbeProtocolTCP.ToPtr(),
						Port:     to.Int32Ptr(int32(6443)),
					},
				},
				{
					Name: to.StringPtr(coordHealthProbeName),
					Properties: &armnetwork.ProbePropertiesFormat{
						Protocol: armnetwork.ProbeProtocolTCP.ToPtr(),
						Port:     to.Int32Ptr(int32(constants.CoordinatorPort)),
					},
				},
				{
					Name: to.StringPtr(debugdHealthProbeName),
					Properties: &armnetwork.ProbePropertiesFormat{
						Protocol: armnetwork.ProbeProtocolTCP.ToPtr(),
						Port:     to.Int32Ptr(int32(4000)),
					},
				},
			},
			LoadBalancingRules: []*armnetwork.LoadBalancingRule{
				{
					Name: to.StringPtr("kubeLoadBalancerRule"),
					Properties: &armnetwork.LoadBalancingRulePropertiesFormat{
						FrontendIPConfiguration: &armnetwork.SubResource{
							ID: to.StringPtr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/frontendIPConfigurations/" + frontEndIPConfigName),
						},
						FrontendPort: to.Int32Ptr(int32(6443)),
						BackendPort:  to.Int32Ptr(int32(6443)),
						Protocol:     armnetwork.TransportProtocolTCP.ToPtr(),
						Probe: &armnetwork.SubResource{
							ID: to.StringPtr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/probes/" + kubeHealthProbeName),
						},
						DisableOutboundSnat: to.BoolPtr(true),
						BackendAddressPools: []*armnetwork.SubResource{
							{
								ID: to.StringPtr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/backendAddressPools/" + backEndAddressPoolControlPlaneName),
							},
						},
					},
				},
				{
					Name: to.StringPtr("coordLoadBalancerRule"),
					Properties: &armnetwork.LoadBalancingRulePropertiesFormat{
						FrontendIPConfiguration: &armnetwork.SubResource{
							ID: to.StringPtr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/frontendIPConfigurations/" + frontEndIPConfigName),
						},
						FrontendPort: to.Int32Ptr(int32(constants.CoordinatorPort)),
						BackendPort:  to.Int32Ptr(int32(constants.CoordinatorPort)),
						Protocol:     armnetwork.TransportProtocolTCP.ToPtr(),
						Probe: &armnetwork.SubResource{
							ID: to.StringPtr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/probes/" + coordHealthProbeName),
						},
						DisableOutboundSnat: to.BoolPtr(true),
						BackendAddressPools: []*armnetwork.SubResource{
							{
								ID: to.StringPtr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/backendAddressPools/" + backEndAddressPoolControlPlaneName),
							},
						},
					},
				},
				{
					Name: to.StringPtr("debudLoadBalancerRule"),
					Properties: &armnetwork.LoadBalancingRulePropertiesFormat{
						FrontendIPConfiguration: &armnetwork.SubResource{
							ID: to.StringPtr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/frontendIPConfigurations/" + frontEndIPConfigName),
						},
						FrontendPort: to.Int32Ptr(int32(4000)),
						BackendPort:  to.Int32Ptr(int32(4000)),
						Protocol:     armnetwork.TransportProtocolTCP.ToPtr(),
						Probe: &armnetwork.SubResource{
							ID: to.StringPtr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/probes/" + debugdHealthProbeName),
						},
						DisableOutboundSnat: to.BoolPtr(true),
						BackendAddressPools: []*armnetwork.SubResource{
							{
								ID: to.StringPtr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/backendAddressPools/" + backEndAddressPoolControlPlaneName),
							},
						},
					},
				},
			},
			OutboundRules: []*armnetwork.OutboundRule{
				{
					Name: to.StringPtr("outboundRuleControlPlane"),
					Properties: &armnetwork.OutboundRulePropertiesFormat{
						FrontendIPConfigurations: []*armnetwork.SubResource{
							{
								ID: to.StringPtr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/frontendIPConfigurations/" + frontEndIPConfigName),
							},
						},
						BackendAddressPool: &armnetwork.SubResource{
							ID: to.StringPtr("/subscriptions/" + l.Subscription + "/resourceGroups/" + l.ResourceGroup + "/providers/Microsoft.Network/loadBalancers/" + l.Name + "/backendAddressPools/all"),
						},
						Protocol: armnetwork.LoadBalancerOutboundRuleProtocolAll.ToPtr(),
					},
				},
			},
		},
	}
}
