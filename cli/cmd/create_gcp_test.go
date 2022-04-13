package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/cli/gcp"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGCPCmdArgumentValidation(t *testing.T) {
	testCases := map[string]struct {
		args      []string
		expectErr bool
	}{
		"valid create 1":               {[]string{"3", "3", "n2d-standard-2"}, false},
		"valid create 2":               {[]string{"3", "7", "n2d-standard-16"}, false},
		"valid create 3":               {[]string{"1", "2", "n2d-standard-96"}, false},
		"invalid too many arguments":   {[]string{"3", "2", "n2d-standard-2", "n2d-standard-2"}, true},
		"invalid too many arguments 2": {[]string{"3", "2", "n2d-standard-2", "2"}, true},
		"invalid no coordinators":      {[]string{"0", "1", "n2d-standard-2"}, true},
		"invalid no nodes":             {[]string{"1", "0", "n2d-standard-2"}, true},
		"invalid first is no int":      {[]string{"n2d-standard-2", "1", "n2d-standard-2"}, true},
		"invalid second is no int":     {[]string{"3", "n2d-standard-2", "n2d-standard-2"}, true},
		"invalid third is no size":     {[]string{"2", "2", "2"}, true},
		"invalid wrong order":          {[]string{"n2d-standard-2", "2", "2"}, true},
	}

	cmd := newCreateGCPCmd()

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

func TestCreateGCP(t *testing.T) {
	testState := state.ConstellationState{
		CloudProvider: cloudprovider.GCP.String(),
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
			"id-0": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
			"id-1": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
			"id-2": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.1",
			},
		},
		GCPNodeInstanceGroup:           "nodes-group",
		GCPCoordinatorInstanceGroup:    "coordinator-group",
		GCPNodeInstanceTemplate:        "node-template",
		GCPCoordinatorInstanceTemplate: "coordinator-template",
		GCPNetwork:                     "network",
		GCPSubnetwork:                  "subnetwork",
		GCPFirewalls:                   []string{"coordinator", "wireguard", "ssh"},
	}
	someErr := errors.New("failed")
	config := config.Default()

	testCases := map[string]struct {
		existingState    *state.ConstellationState
		client           gcpclient
		interactive      bool
		interactiveStdin string
		stateExpected    state.ConstellationState
		errExpected      bool
	}{
		"create some instances": {
			client:        &fakeGcpClient{},
			stateExpected: testState,
		},
		"state already exists": {
			existingState: &testState,
			client:        &fakeGcpClient{},
			errExpected:   true,
		},
		"create some instances interactive": {
			client:           &fakeGcpClient{},
			interactive:      true,
			interactiveStdin: "y\n",
			stateExpected:    testState,
			errExpected:      false,
		},
		"fail getState": {
			client:      &stubGcpClient{getStateErr: someErr},
			errExpected: true,
		},
		"fail createVPCs": {
			client:      &stubGcpClient{createVPCsErr: someErr},
			errExpected: true,
		},
		"fail createFirewall": {
			client:      &stubGcpClient{createFirewallErr: someErr},
			errExpected: true,
		},
		"fail createInstances": {
			client:      &stubGcpClient{createInstancesErr: someErr},
			errExpected: true,
		},
		"error on rollback": {
			client:      &stubGcpClient{createInstancesErr: someErr, terminateVPCsErr: someErr},
			errExpected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newCreateGCPCmd()
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
				require.NoError(fileHandler.WriteJSON(*config.StatePath, *tc.existingState, file.OptNone))
			}

			err := createGCP(cmd, tc.client, fileHandler, config, "n2d-standard-2", 3, 2)
			if tc.errExpected {
				assert.Error(err)
				if stubClient, ok := tc.client.(*stubGcpClient); ok {
					// Should have made a rollback on error.
					assert.True(stubClient.terminateFirewallCalled)
					assert.True(stubClient.terminateInstancesCalled)
					assert.True(stubClient.terminateVPCsCalled)
				}
			} else {
				assert.NoError(err)
				var stat state.ConstellationState
				err := fileHandler.ReadJSON(*config.StatePath, &stat)
				assert.NoError(err)
				assert.Equal(tc.stateExpected, stat)
			}
		})
	}
}

func TestCreateGCPCompletion(t *testing.T) {
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
			toComplete:      "n2d-stan",
			resultExpected:  gcp.InstanceTypes,
			shellCDExpected: cobra.ShellCompDirectiveDefault,
		},
		"fourth arg": {
			args:            []string{"23", "24", "n2d-standard-2"},
			toComplete:      "n2d-stan",
			resultExpected:  []string{},
			shellCDExpected: cobra.ShellCompDirectiveError,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := &cobra.Command{}
			result, shellCD := createGCPCompletion(cmd, tc.args, tc.toComplete)
			assert.Equal(tc.resultExpected, result)
			assert.Equal(tc.shellCDExpected, shellCD)
		})
	}
}
