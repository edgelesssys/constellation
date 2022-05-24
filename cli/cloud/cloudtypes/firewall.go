package cloudtypes

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/edgelesssys/constellation/internal/config"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

type FirewallRule = config.FirewallRule

type Firewall config.Firewall

func (f Firewall) GCP() ([]*computepb.Firewall, error) {
	var fw []*computepb.Firewall
	for _, rule := range f {
		var srcRange []string
		if rule.IPRange != "" {
			srcRange = []string{rule.IPRange}
		}

		var ports []string
		if rule.FromPort != 0 || rule.ToPort != 0 {
			port, err := portOrRange(rule.FromPort, rule.ToPort)
			if err != nil {
				return nil, err
			}
			ports = []string{port}
		}

		fw = append(fw, &computepb.Firewall{
			Allowed: []*computepb.Allowed{
				{
					IPProtocol: proto.String(rule.Protocol),
					Ports:      ports,
				},
			},
			Description:  proto.String(rule.Description),
			SourceRanges: srcRange,
			Name:         proto.String(rule.Name),
		})
	}
	return fw, nil
}

func (f Firewall) Azure() ([]*armnetwork.SecurityRule, error) {
	var fw []*armnetwork.SecurityRule
	for i, rule := range f {
		// format string according to armnetwork.SecurityRuleProtocol specification
		protocol := cases.Title(language.English).String(rule.Protocol)

		dstPortRange, err := portOrRange(rule.FromPort, rule.ToPort)
		if err != nil {
			return nil, err
		}
		fw = append(fw, &armnetwork.SecurityRule{
			Name: proto.String(rule.Name),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Description:              proto.String(rule.Description),
				Protocol:                 (*armnetwork.SecurityRuleProtocol)(proto.String(protocol)),
				SourceAddressPrefix:      proto.String(rule.IPRange),
				SourcePortRange:          proto.String("*"),
				DestinationAddressPrefix: proto.String(rule.IPRange),
				DestinationPortRange:     proto.String(dstPortRange),
				Access:                   armnetwork.SecurityRuleAccessAllow.ToPtr(),
				Direction:                armnetwork.SecurityRuleDirectionInbound.ToPtr(),
				// Each security role needs a unique priority
				Priority: proto.Int32(int32(100 * (i + 1))),
			},
		})
	}
	return fw, nil
}

func (f Firewall) AWS() []ec2types.IpPermission {
	var fw []ec2types.IpPermission
	for _, rule := range f {
		fw = append(fw, ec2types.IpPermission{
			FromPort:   proto.Int32(int32(rule.FromPort)),
			ToPort:     proto.Int32(int32(rule.ToPort)),
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

const (
	MinPort = 0
	MaxPort = 65535
)

// PortOutOfRangeError occurs when either FromPort or ToPort are out of range
// of [MinPort-MaxPort].
type PortOutOfRangeError struct {
	FromPort int
	ToPort   int
}

func (p *PortOutOfRangeError) Error() string {
	return fmt.Sprintf(
		"[%d-%d] not in allowed port range of [%d-%d]",
		p.FromPort, p.ToPort, MinPort, MaxPort,
	)
}

// portOrRange returns "fromPort" as single port, if toPort is zero.
// If toPort is >0 a port range of form "fromPort-toPort".
// If either value is negative PortOutOfRangeError is returned.
func portOrRange(fromPort, toPort int) (string, error) {
	if fromPort < MinPort || toPort < MinPort || fromPort > MaxPort || toPort > MaxPort {
		return "", &PortOutOfRangeError{FromPort: fromPort, ToPort: toPort}
	}
	if toPort == MinPort || fromPort == toPort {
		return fmt.Sprintf("%d", fromPort), nil
	}
	if toPort > MinPort {
		return fmt.Sprintf("%d-%d", fromPort, toPort), nil
	}
	return "", &PortOutOfRangeError{FromPort: fromPort, ToPort: toPort}
}
