package client

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/edgelesssys/constellation/cli/cloud/cloudtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSecurityGroup(t *testing.T) {
	testInput := SecurityGroupInput{
		Inbound: cloudtypes.Firewall{
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
		},
		Outbound: cloudtypes.Firewall{
			{
				Description: "perm3",
				Protocol:    "TCP",
				IPRange:     "192.0.2.0/24",
				Port:        4040,
			},
		},
	}
	someErr := errors.New("failed")
	var noErr error

	testCases := map[string]struct {
		api                   stubAPI
		securityGroup         string
		input                 SecurityGroupInput
		errExpected           bool
		securityGroupExpected string
	}{
		"create security group": {
			api:                   stubAPI{securityGroup: types.SecurityGroup{GroupId: aws.String("sg-test")}},
			input:                 testInput,
			securityGroupExpected: "sg-test",
		},
		"create security group without permissions": {
			api:                   stubAPI{securityGroup: types.SecurityGroup{GroupId: aws.String("sg-test")}},
			input:                 SecurityGroupInput{},
			securityGroupExpected: "sg-test",
		},
		"client already has security group": {
			api:           stubAPI{},
			securityGroup: "sg-test",
			input:         testInput,
			errExpected:   true,
		},
		"create returns nil security group ID": {
			api:         stubAPI{securityGroup: types.SecurityGroup{GroupId: nil}},
			input:       testInput,
			errExpected: true,
		},
		"create API error": {
			api:         stubAPI{createSecurityGroupErr: someErr},
			input:       testInput,
			errExpected: true,
		},
		"create DryRun API error": {
			api:         stubAPI{createSecurityGroupDryRunErr: &someErr},
			input:       testInput,
			errExpected: true,
		},
		"create DryRun missing expected error": {
			api:         stubAPI{createSecurityGroupDryRunErr: &noErr},
			input:       testInput,
			errExpected: true,
		},
		"authorize error": {
			api: stubAPI{
				securityGroup:                    types.SecurityGroup{GroupId: aws.String("sg-test")},
				authorizeSecurityGroupIngressErr: someErr,
			},
			input:       testInput,
			errExpected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client, err := newClient(tc.api)
			require.NoError(err)
			client.securityGroup = tc.securityGroup

			err = client.CreateSecurityGroup(context.Background(), tc.input)
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.securityGroupExpected, client.securityGroup)
			}
		})
	}
}

func TestDeleteSecurityGroup(t *testing.T) {
	someErr := errors.New("failed")
	var noErr error

	testCases := map[string]struct {
		api           stubAPI
		securityGroup string
		errExpected   bool
	}{
		"delete security group": {
			api:           stubAPI{},
			securityGroup: "sg-test",
		},
		"client without security group": {
			api: stubAPI{},
		},
		"delete API error": {
			api:           stubAPI{deleteSecurityGroupErr: someErr},
			securityGroup: "sg-test",
			errExpected:   true,
		},
		"delete DryRun API error": {
			api:           stubAPI{deleteSecurityGroupDryRunErr: &someErr},
			securityGroup: "sg-test",
			errExpected:   true,
		},
		"delete DryRun missing expected error": {
			api:           stubAPI{deleteSecurityGroupDryRunErr: &noErr},
			securityGroup: "sg-test",
			errExpected:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client, err := newClient(tc.api)
			require.NoError(err)
			client.securityGroup = tc.securityGroup

			err = client.DeleteSecurityGroup(context.Background())
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Empty(client.securityGroup)
			}
		})
	}
}

func TestAuthorizeSecurityGroup(t *testing.T) {
	testInput := SecurityGroupInput{
		Inbound: cloudtypes.Firewall{
			{
				Description: "perm1",
				Protocol:    "TCP",
				IPRange: " 	192.0.2.0/24",
				Port: 22,
			},
			{
				Description: "perm2",
				Protocol:    "UDP",
				IPRange:     "192.0.2.0/24",
				Port:        4433,
			},
		},
		Outbound: cloudtypes.Firewall{
			{
				Description: "perm3",
				Protocol:    "TCP",
				IPRange:     "192.0.2.0/24",
				Port:        4040,
			},
		},
	}
	someErr := errors.New("failed")
	var noErr error

	testCases := map[string]struct {
		api           stubAPI
		securityGroup string
		input         SecurityGroupInput
		errExpected   bool
	}{
		"authorize": {
			api:           stubAPI{},
			securityGroup: "sg-test",
			input:         testInput,
			errExpected:   false,
		},
		"client without security group": {
			api:         stubAPI{},
			input:       testInput,
			errExpected: true,
		},
		"authorizeIngress API error": {
			api:           stubAPI{authorizeSecurityGroupIngressErr: someErr},
			securityGroup: "sg-test",
			input:         testInput,
			errExpected:   true,
		},
		"authorizeIngress DryRun API error": {
			api:           stubAPI{authorizeSecurityGroupIngressDryRunErr: &someErr},
			securityGroup: "sg-test",
			input:         testInput,
			errExpected:   true,
		},
		"authorizeIngress DryRun missing expected error": {
			api:           stubAPI{authorizeSecurityGroupIngressDryRunErr: &noErr},
			securityGroup: "sg-test",
			input:         testInput,
			errExpected:   true,
		},
		"authorizeEgress API error": {
			api:           stubAPI{authorizeSecurityGroupEgressErr: someErr},
			securityGroup: "sg-test",
			input:         testInput,
			errExpected:   true,
		},
		"authorizeEgress DryRun API error": {
			api:           stubAPI{authorizeSecurityGroupEgressDryRunErr: &someErr},
			securityGroup: "sg-test",
			input:         testInput,
			errExpected:   true,
		},
		"authorizeEgress DryRun missing expected error": {
			api:           stubAPI{authorizeSecurityGroupEgressDryRunErr: &noErr},
			securityGroup: "sg-test",
			input:         testInput,
			errExpected:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client, err := newClient(tc.api)
			require.NoError(err)
			client.securityGroup = tc.securityGroup

			err = client.authorizeSecurityGroup(context.Background(), tc.input)
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
