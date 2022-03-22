package cloudtypes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

type FirewallRule struct {
	Name        string
	Description string
	Protocol    string
	IPRange     string
	Port        int
}

type Firewall []FirewallRule

func (f Firewall) GCP() []*computepb.Firewall {
	var fw []*computepb.Firewall
	for _, rule := range f {
		var destRange []string = nil
		if rule.IPRange != "" {
			destRange = append(destRange, rule.IPRange)
		}

		fw = append(fw, &computepb.Firewall{
			Allowed: []*computepb.Allowed{
				{
					IPProtocol: proto.String(rule.Protocol),
					Ports:      []string{fmt.Sprint(rule.Port)},
				},
			},
			Description:       proto.String(rule.Description),
			DestinationRanges: destRange,
			Name:              proto.String(rule.Name),
		})
	}
	return fw
}

func (f Firewall) Azure() []*armnetwork.SecurityRule {
	var fw []*armnetwork.SecurityRule
	for i, rule := range f {
		// format string according to armnetwork.SecurityRuleProtocol specification
		protocol := strings.Title(strings.ToLower(rule.Protocol))

		fw = append(fw, &armnetwork.SecurityRule{
			Name: proto.String(rule.Name),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Description:              proto.String(rule.Description),
				Protocol:                 (*armnetwork.SecurityRuleProtocol)(proto.String(protocol)),
				SourceAddressPrefix:      proto.String(rule.IPRange),
				SourcePortRange:          proto.String("*"),
				DestinationAddressPrefix: proto.String(rule.IPRange),
				DestinationPortRange:     proto.String(strconv.Itoa(rule.Port)),
				Access:                   armnetwork.SecurityRuleAccessAllow.ToPtr(),
				Direction:                armnetwork.SecurityRuleDirectionInbound.ToPtr(),
				// Each security role needs a unique priority
				Priority: proto.Int32(int32(100 * (i + 1))),
			},
		})
	}
	return fw
}

func (f Firewall) AWS() []ec2types.IpPermission {
	var fw []ec2types.IpPermission
	for _, rule := range f {
		fw = append(fw, ec2types.IpPermission{
			FromPort:   proto.Int32(int32(rule.Port)),
			ToPort:     proto.Int32(int32(rule.Port)),
			IpProtocol: proto.String(rule.Protocol),
			IpRanges: []ec2types.IpRange{
				{
					CidrIp:      proto.String(rule.IPRange),
					Description: proto.String(rule.Description),
				},
			},
		})
	}
	return fw
}
