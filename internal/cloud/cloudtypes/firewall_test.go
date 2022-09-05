/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudtypes

import (
	"strconv"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
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
			FromPort:    9000,
		},
		{
			Name:        "test-2",
			Description: "This is the Test-2 Permission",
			Protocol:    "udp",
			IPRange:     "",
			FromPort:    51820,
		},
		{
			Name:        "test-3",
			Description: "This is the Test-3 Permission",
			Protocol:    "tcp",
			IPRange:     "192.0.2.0/24",
			FromPort:    4000,
		},
	}

	firewalls, err := testFw.GCP()
	assert.NoError(err)
	assert.Equal(len(testFw), len(firewalls))

	// Check permissions
	for i := 0; i < len(testFw); i++ {
		firewall1 := firewalls[i]
		actualPermission1 := firewall1.Allowed[0]

		actualPort, err := strconv.Atoi(actualPermission1.GetPorts()[0])
		require.NoError(err)
		assert.Equal(testFw[i].FromPort, actualPort)
		assert.Equal(testFw[i].Protocol, actualPermission1.GetIPProtocol())

		assert.Equal(testFw[i].Name, firewall1.GetName())
		assert.Equal(testFw[i].Description, firewall1.GetDescription())

		if testFw[i].IPRange != "" {
			require.Len(firewall1.GetSourceRanges(), 1)
			assert.Equal(testFw[i].IPRange, firewall1.GetSourceRanges()[0])
		}
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
			FromPort:    22,
		},
		{
			Name:        "perm2",
			Description: "perm2 description",
			Protocol:    "udp",
			IPRange:     "192.0.2.0/24",
			FromPort:    4433,
		},
		{
			Name:        "perm3",
			Description: "perm3 description",
			Protocol:    "tcp",
			IPRange:     "192.0.2.0/24",
			FromPort:    4433,
		},
	}
	wantOutput := []*armnetwork.SecurityRule{
		{
			Name: proto.String("perm1"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Description:              proto.String("perm1 description"),
				Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
				SourceAddressPrefix:      proto.String("192.0.2.0/24"),
				SourcePortRange:          proto.String("*"),
				DestinationAddressPrefix: proto.String("192.0.2.0/24"),
				DestinationPortRange:     proto.String("22"),
				Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
				Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
				Priority:                 proto.Int32(100),
			},
		},
		{
			Name: proto.String("perm2"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Description:              proto.String("perm2 description"),
				Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolUDP),
				SourceAddressPrefix:      proto.String("192.0.2.0/24"),
				SourcePortRange:          proto.String("*"),
				DestinationAddressPrefix: proto.String("192.0.2.0/24"),
				DestinationPortRange:     proto.String("4433"),
				Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
				Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
				Priority:                 proto.Int32(200),
			},
		},
		{
			Name: proto.String("perm3"),
			Properties: &armnetwork.SecurityRulePropertiesFormat{
				Description:              proto.String("perm3 description"),
				Protocol:                 to.Ptr(armnetwork.SecurityRuleProtocolTCP),
				SourceAddressPrefix:      proto.String("192.0.2.0/24"),
				SourcePortRange:          proto.String("*"),
				DestinationAddressPrefix: proto.String("192.0.2.0/24"),
				DestinationPortRange:     proto.String("4433"),
				Access:                   to.Ptr(armnetwork.SecurityRuleAccessAllow),
				Direction:                to.Ptr(armnetwork.SecurityRuleDirectionInbound),
				Priority:                 proto.Int32(300),
			},
		},
	}

	out, err := input.Azure()
	assert.NoError(err)
	assert.Equal(wantOutput, out)
}

func TestIPPermissonsToAWS(t *testing.T) {
	assert := assert.New(t)

	input := Firewall{
		{
			Description: "perm1",
			Protocol:    "TCP",
			IPRange:     "192.0.2.0/24",
			FromPort:    22,
			ToPort:      22,
		},
		{
			Description: "perm2",
			Protocol:    "UDP",
			IPRange:     "192.0.2.0/24",
			FromPort:    4433,
			ToPort:      4433,
		},
		{
			Description: "perm3",
			Protocol:    "TCP",
			IPRange:     "192.0.2.0/24",
			FromPort:    4433,
			ToPort:      4433,
		},
	}
	wantOutput := []ec2types.IpPermission{
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
	assert.Equal(wantOutput, out)
}

func TestPortOrRange(t *testing.T) {
	testCases := map[string]struct {
		fromPort int
		toPort   int
		result   string
		wantErr  bool
	}{
		"ssh": {
			fromPort: 22,
			result:   "22",
		},
		"https": {
			fromPort: 443,
			result:   "443",
		},
		"nodePorts": {
			fromPort: 30000,
			toPort:   32767,
			result:   "30000-32767",
		},
		"negative fromPort": {
			fromPort: -1,
			wantErr:  true,
		},
		"negative toPort": {
			toPort:  -1,
			wantErr: true,
		},
		"same value no range": {
			fromPort: 22,
			toPort:   22,
			result:   "22",
		},
		"from zero to ssh": {
			toPort: 22,
			result: "0-22",
		},
		"from max": {
			fromPort: MaxPort,
			result:   "65535",
		},
		"from max+1": {
			fromPort: MaxPort + 1,
			wantErr:  true,
		},
		"to max": {
			toPort: MaxPort,
			result: "0-65535",
		},
		"to max+1": {
			toPort:  MaxPort + 1,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			got, err := portOrRange(tc.fromPort, tc.toPort)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.result, got)
		})
	}
}
