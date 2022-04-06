package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/cli/azure"
	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAzureCmdArgumentValidation(t *testing.T) {
	testCases := map[string]struct {
		args      []string
		expectErr bool
	}{
		"valid create 1":              {[]string{"3", "3", "Standard_DC2as_v5"}, false},
		"valid create 2":              {[]string{"3", "7", "Standard_DC4as_v5"}, false},
		"valid create 3":              {[]string{"1", "2", "Standard_DC8as_v5"}, false},
		"invalid to many arguments":   {[]string{"3", "2", "Standard_DC2as_v5", "Standard_DC2as_v5"}, true},
		"invalid to many arguments 2": {[]string{"3", "2", "Standard_DC2as_v5", "2"}, true},
		"invalid no coordinators":     {[]string{"0", "1", "Standard_DC2as_v5"}, true},
		"invalid no nodes":            {[]string{"1", "0", "Standard_DC2as_v5"}, true},
		"invalid first is no int":     {[]string{"Standard_DC2as_v5", "1", "Standard_DC2as_v5"}, true},
		"invalid second is no int":    {[]string{"1", "Standard_DC2as_v5", "Standard_DC2as_v5"}, true},
		"invalid third is no size":    {[]string{"2", "2", "2"}, true},
		"invalid wrong order":         {[]string{"Standard_DC2as_v5", "2", "2"}, true},
	}

	cmd := newCreateAzureCmd()

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

func TestCreateAzure(t *testing.T) {
	testState := state.ConstellationState{
		CloudProvider: cloudprovider.Azure.String(),
		AzureNodes: azure.Instances{
			"0": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
			"1": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
		},
		AzureCoordinators: azure.Instances{
			"0": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
			"1": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
			"2": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
		},
		AzureResourceGroup:        "resource-group",
		AzureSubnet:               "subnet",
		AzureNetworkSecurityGroup: "network-security-group",
		AzureNodesScaleSet:        "nodes-scale-set",
		AzureCoordinatorsScaleSet: "coordinators-scale-set",
	}
	someErr := errors.New("failed")
	config := config.Default()

	testCases := map[string]struct {
		existingState    *state.ConstellationState
		client           azureclient
		interactive      bool
		interactiveStdin string
		stateExpected    state.ConstellationState
		errExpected      bool
	}{
		"create some instances": {
			client:        &fakeAzureClient{},
			stateExpected: testState,
		},
		"state already exists": {
			existingState: &testState,
			client:        &fakeAzureClient{},
			errExpected:   true,
		},
		"create some instances interactive": {
			client:           &fakeAzureClient{},
			interactive:      true,
			interactiveStdin: "y\n",
			stateExpected:    testState,
			errExpected:      false,
		},
		"fail getState": {
			client:      &stubAzureClient{getStateErr: someErr},
			errExpected: true,
		},
		"fail createVirtualNetwork": {
			client:      &stubAzureClient{createVirtualNetworkErr: someErr},
			errExpected: true,
		},
		"fail createSecurityGroup": {
			client:      &stubAzureClient{createSecurityGroupErr: someErr},
			errExpected: true,
		},
		"fail createInstances": {
			client:      &stubAzureClient{createInstancesErr: someErr},
			errExpected: true,
		},
		"fail createResourceGroup": {
			client:      &stubAzureClient{createResourceGroupErr: someErr},
			errExpected: true,
		},
		"error on rollback": {
			client:      &stubAzureClient{createInstancesErr: someErr, terminateResourceGroupErr: someErr},
			errExpected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newCreateAzureCmd()
			cmd.Flags().BoolP("yes", "y", false, "")
			out := bytes.NewBufferString("")
			cmd.SetOut(out)
			errOut := bytes.NewBufferString("")
			cmd.SetErr(errOut)
			in := bytes.NewBufferString(tc.interactiveStdin)
			cmd.SetIn(in)
			if !tc.interactive {
				require.NoError(cmd.Flags().Set("yes", "true")) // disable interactivity
			}

			fs := afero.NewMemMapFs()
			fileHandler := file.NewHandler(fs)
			if tc.existingState != nil {
				require.NoError(fileHandler.WriteJSON(constants.StateFilename, *tc.existingState, file.OptNone))
			}

			err := createAzure(cmd, tc.client, fileHandler, config, "Standard_D2s_v3", 3, 2)
			if tc.errExpected {
				assert.Error(err)
				if stubClient, ok := tc.client.(*stubAzureClient); ok {
					// Should have made a rollback on error.
					assert.True(stubClient.terminateResourceGroupCalled)
				}
			} else {
				assert.NoError(err)
				var state state.ConstellationState
				err := fileHandler.ReadJSON(constants.StateFilename, &state)
				assert.NoError(err)
				assert.Equal(tc.stateExpected, state)
			}
		})
	}
}

func TestCreateAzureCompletion(t *testing.T) {
	testCases := map[string]struct {
		args            []string
		toComplete      string
		resultExpected  []string
		shellCDExpected cobra.ShellCompDirective
	}{
		"first arg": {
			args:            []string{},
			toComplete:      "21",
			resultExpected:  []string{},
			shellCDExpected: cobra.ShellCompDirectiveNoFileComp,
		},
		"second arg": {
			args:            []string{"23"},
			toComplete:      "21",
			resultExpected:  []string{},
			shellCDExpected: cobra.ShellCompDirectiveNoFileComp,
		},
		"third arg": {
			args:            []string{"23", "24"},
			toComplete:      "Standard_D",
			resultExpected:  azure.InstanceTypes,
			shellCDExpected: cobra.ShellCompDirectiveDefault,
		},
		"fourth arg": {
			args:            []string{"23", "24", "Standard_D2s_v3"},
			toComplete:      "Standard_D",
			resultExpected:  []string{},
			shellCDExpected: cobra.ShellCompDirectiveError,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := &cobra.Command{}
			result, shellCD := createAzureCompletion(cmd, tc.args, tc.toComplete)
			assert.Equal(tc.resultExpected, result)
			assert.Equal(tc.shellCDExpected, shellCD)
		})
	}
}
