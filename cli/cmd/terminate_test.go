package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/cli/azure"
	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/ec2"
	"github.com/edgelesssys/constellation/cli/gcp"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/stretchr/testify/assert"
)

func TestTerminateCmdArgumentValidation(t *testing.T) {
	testCases := map[string]struct {
		args      []string
		expectErr bool
	}{
		"no args":         {[]string{}, false},
		"some args":       {[]string{"hello", "test"}, true},
		"some other args": {[]string{"12", "2"}, true},
	}

	cmd := newTerminateCmd()

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			err := cmd.ValidateArgs(tc.args)
			if tc.expectErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestTerminateEC2(t *testing.T) {
	testState := state.ConstellationState{
		CloudProvider: cloudprovider.AWS.String(),
		EC2Instances: ec2.Instances{
			"id-0": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
			"id-1": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
			"id-3": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
		},
		EC2SecurityGroup: "sg-test",
	}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		existingState state.ConstellationState
		client        ec2client
		errExpected   bool
	}{
		"terminate existing instances": {
			existingState: testState,
			client:        &fakeEc2Client{},
			errExpected:   false,
		},
		"state without instances": {
			existingState: state.ConstellationState{
				CloudProvider: cloudprovider.AWS.String(),
				EC2Instances:  ec2.Instances{},
			},
			client:      &fakeEc2Client{},
			errExpected: true,
		},
		"fail TerminateInstances": {
			existingState: testState,
			client:        &stubEc2Client{terminateInstancesErr: someErr},
			errExpected:   true,
		},
		"fail DeleteSecurityGroup": {
			existingState: testState,
			client:        &stubEc2Client{deleteSecurityGroupErr: someErr},
			errExpected:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := newTerminateCmd()
			out := bytes.NewBufferString("")
			cmd.SetOut(out)
			errOut := bytes.NewBufferString("")
			cmd.SetErr(errOut)

			err := terminateEC2(cmd, tc.client, tc.existingState)
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestTerminateGCP(t *testing.T) {
	testState := state.ConstellationState{
		GCPNodes: gcp.Instances{
			"id-0": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
			"id-1": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
		},
		GCPCoordinators: gcp.Instances{
			"id-c": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
		},
		GCPNodeInstanceGroup:           "nodes-group",
		GCPCoordinatorInstanceGroup:    "coordinator-group",
		GCPNodeInstanceTemplate:        "template",
		GCPCoordinatorInstanceTemplate: "template",
		GCPNetwork:                     "network",
		GCPFirewalls:                   []string{"coordinator", "wireguard", "ssh"},
	}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		existingState state.ConstellationState
		client        gcpclient
		errExpected   bool
	}{
		"terminate existing instances": {
			existingState: testState,
			client:        &fakeGcpClient{},
		},
		"state without instances": {
			existingState: state.ConstellationState{EC2Instances: ec2.Instances{}},
			client:        &fakeGcpClient{},
		},
		"state not found": {
			existingState: testState,
			client:        &fakeGcpClient{},
		},
		"fail setState": {
			existingState: testState,
			client:        &stubGcpClient{setStateErr: someErr},
			errExpected:   true,
		},
		"fail terminateFirewall": {
			existingState: testState,
			client:        &stubGcpClient{terminateFirewallErr: someErr},
			errExpected:   true,
		},
		"fail terminateVPC": {
			existingState: testState,
			client:        &stubGcpClient{terminateVPCsErr: someErr},
			errExpected:   true,
		},
		"fail terminateInstances": {
			existingState: testState,
			client:        &stubGcpClient{terminateInstancesErr: someErr},
			errExpected:   true,
		},
		"fail terminateServiceAccount": {
			existingState: testState,
			client:        &stubGcpClient{terminateServiceAccountErr: someErr},
			errExpected:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := newTerminateCmd()
			out := bytes.NewBufferString("")
			cmd.SetOut(out)
			errOut := bytes.NewBufferString("")
			cmd.SetErr(errOut)

			err := terminateGCP(cmd, tc.client, tc.existingState)
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
				stat, err := tc.client.GetState()
				assert.NoError(err)
				assert.Equal(state.ConstellationState{
					CloudProvider: cloudprovider.GCP.String(),
				}, stat)
			}
		})
	}
}

func TestTerminateAzure(t *testing.T) {
	testState := state.ConstellationState{
		CloudProvider: cloudprovider.Azure.String(),
		AzureNodes: azure.Instances{
			"id-0": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
			"id-1": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
		},
		AzureCoordinators: azure.Instances{
			"id-c": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
		},
		AzureResourceGroup: "test",
	}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		existingState state.ConstellationState
		client        azureclient
		errExpected   bool
	}{
		"terminate existing instances": {
			existingState: testState,
			client:        &fakeAzureClient{},
		},
		"state resource group": {
			existingState: state.ConstellationState{AzureResourceGroup: ""},
			client:        &fakeAzureClient{},
		},
		"state not found": {
			existingState: testState,
			client:        &fakeAzureClient{},
		},
		"fail setState": {
			existingState: testState,
			client:        &stubAzureClient{setStateErr: someErr},
			errExpected:   true,
		},
		"fail resource group termination": {
			existingState: testState,
			client:        &stubAzureClient{terminateResourceGroupErr: someErr},
			errExpected:   true,
		},
		"fail service principal termination": {
			existingState: testState,
			client:        &stubAzureClient{terminateServicePrincipalErr: someErr},
			errExpected:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := newTerminateCmd()
			out := bytes.NewBufferString("")
			cmd.SetOut(out)
			errOut := bytes.NewBufferString("")
			cmd.SetErr(errOut)

			err := terminateAzure(cmd, tc.client, tc.existingState)
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
				stat, err := tc.client.GetState()
				assert.NoError(err)
				assert.Equal(state.ConstellationState{
					CloudProvider: cloudprovider.Azure.String(),
				}, stat)
			}
		})
	}
}
