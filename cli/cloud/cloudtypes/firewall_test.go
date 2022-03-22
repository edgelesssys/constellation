package cloudtypes

import (
	"strconv"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestFirewallGCP(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	testFw := Firewall{
		{
			Name:        "test-1",
			Description: "This is the Test-1 Permission",
			Protocol:    "tcp",
			IPRange:     "",
			Port:        9000,
		},
		{
			Name:        "test-2",
			Description: "This is the Test-2 Permission",
			Protocol:    "udp",
			IPRange:     "",
			Port:        51820,
		},
	}

	firewalls := testFw.GCP()
	assert.Equal(2, len(firewalls))

	// Check permissions
	for i := 0; i < len(testFw); i++ {
		firewall1 := firewalls[i]
		actualPermission1 := firewall1.Allowed[0]

		actualPort, err := strconv.Atoi(actualPermission1.GetPorts()[0])
		require.NoError(err)
		assert.Equal(testFw[i].Port, actualPort)
		assert.Equal(testFw[i].Protocol, actualPermission1.GetIPProtocol())

		assert.Equal(testFw[i].Name, firewall1.GetName())
		assert.Equal(testFw[i].Description, firewall1.GetDescription())
	}
}

func TestFirewallAzure(t *testing.T) {
	assert := assert.New(t)

	input := Firewall{
		{
			Name:        "perm1",
			Description: "perm1 description",
			Protocol:    "TCP",
			IPRange:     "192.0.2.0/24",
			Port:        22,
		},
		{
			Name:        "perm2",
			Description: "perm2 description",
			Protocol:    "udp",
			IPRange:     "192.0.2.0/24",
			Port:        4433,
		},
		{
			Name:        "perm3",
			Description: "perm3 description",
			Protocol:    "tcp",
			IPRange:     "192.0.2.0/24",
			Port:        4433,
		},
	}
	expectedOutput := []*armnetwork.SecurityRule{
		{
			Name: proto.String("perm1"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Description:              proto.String("perm1 description"),
				Protocol:                 armnetwork.SecurityRuleProtocolTCP.ToPtr(),
				SourceAddressPrefix:      proto.String("192.0.2.0/24"),
				SourcePortRange:          proto.String("*"),
				DestinationAddressPrefix: proto.String("192.0.2.0/24"),
				DestinationPortRange:     proto.String("22"),
				Access:                   armnetwork.SecurityRuleAccessAllow.ToPtr(),
				Direction:                armnetwork.SecurityRuleDirectionInbound.ToPtr(),
				Priority:                 proto.Int32(100),
			},
		},
		{
			Name: proto.String("perm2"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Description:              proto.String("perm2 description"),
				Protocol:                 armnetwork.SecurityRuleProtocolUDP.ToPtr(),
				SourceAddressPrefix:      proto.String("192.0.2.0/24"),
				SourcePortRange:          proto.String("*"),
				DestinationAddressPrefix: proto.String("192.0.2.0/24"),
				DestinationPortRange:     proto.String("4433"),
				Access:                   armnetwork.SecurityRuleAccessAllow.ToPtr(),
				Direction:                armnetwork.SecurityRuleDirectionInbound.ToPtr(),
				Priority:                 proto.Int32(200),
			},
		},
		{
			Name: proto.String("perm3"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Description:              proto.String("perm3 description"),
				Protocol:                 armnetwork.SecurityRuleProtocolTCP.ToPtr(),
				SourceAddressPrefix:      proto.String("192.0.2.0/24"),
				SourcePortRange:          proto.String("*"),
				DestinationAddressPrefix: proto.String("192.0.2.0/24"),
				DestinationPortRange:     proto.String("4433"),
				Access:                   armnetwork.SecurityRuleAccessAllow.ToPtr(),
				Direction:                armnetwork.SecurityRuleDirectionInbound.ToPtr(),
				Priority:                 proto.Int32(300),
			},
		},
	}

	out := input.Azure()
	assert.Equal(expectedOutput, out)
}

func TestIPPermissonsToAWS(t *testing.T) {
	assert := assert.New(t)

	input := Firewall{
		{
			Description: "perm1",
			Protocol:    "TCP",
			IPRange:     "192.0.2.0/24",
			Port:        22,
		},
		{
			Description: "perm2",
			Protocol:    "UDP",
			IPRange:     "192.0.2.0/24",
			Port:        4433,
		},
		{
			Description: "perm3",
			Protocol:    "TCP",
			IPRange:     "192.0.2.0/24",
			Port:        4433,
		},
	}
	expectedOutput := []ec2types.IpPermission{
		{
			FromPort:   proto.Int32(int32(22)),
			ToPort:     proto.Int32(int32(22)),
			IpProtocol: proto.String("TCP"),
			IpRanges: []ec2types.IpRange{
				{
					CidrIp:      proto.String("192.0.2.0/24"),
					Description: proto.String("perm1"),
				},
			},
		},
		{
			FromPort:   proto.Int32(int32(4433)),
			ToPort:     proto.Int32(int32(4433)),
			IpProtocol: proto.String("UDP"),
			IpRanges: []ec2types.IpRange{
				{
					CidrIp:      proto.String("192.0.2.0/24"),
					Description: proto.String("perm2"),
				},
			},
		},
		{
			FromPort:   proto.Int32(int32(4433)),
			ToPort:     proto.Int32(int32(4433)),
			IpProtocol: proto.String("TCP"),
			IpRanges: []ec2types.IpRange{
				{
					CidrIp:      proto.String("192.0.2.0/24"),
					Description: proto.String("perm3"),
				},
			},
		},
	}

	out := input.AWS()
	assert.Equal(expectedOutput, out)
}
