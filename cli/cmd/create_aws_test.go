package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/cli/ec2"
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAWSCmdArgumentValidation(t *testing.T) {
	testCases := map[string]struct {
		args      []string
		expectErr bool
	}{
		"valid size 4XL":              {[]string{"5", "4xlarge"}, false},
		"valid size 8XL":              {[]string{"4", "8xlarge"}, false},
		"valid size 12XL":             {[]string{"3", "12xlarge"}, false},
		"valid size 16XL":             {[]string{"2", "16xlarge"}, false},
		"valid size 24XL":             {[]string{"2", "24xlarge"}, false},
		"valid short 12XL":            {[]string{"4", "12xl"}, false},
		"valid short 24XL":            {[]string{"2", "24xl"}, false},
		"valid capitalized":           {[]string{"3", "24XlARge"}, false},
		"valid short capitalized":     {[]string{"4", "16XL"}, false},
		"invalid to many arguments":   {[]string{"2", "4xl", "2xl"}, true},
		"invalid to many arguments 2": {[]string{"2", "4xl", "2"}, true},
		"invalidOnlyOneInstance":      {[]string{"1", "4xl"}, true},
		"invalid first is no int":     {[]string{"xl", "4xl"}, true},
		"invalid second is no size":   {[]string{"2", "2"}, true},
		"invalid wrong order":         {[]string{"4xl", "2"}, true},
	}

	cmd := newCreateAWSCmd()

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

func TestCreateAWS(t *testing.T) {
	testState := state.ConstellationState{
		CloudProvider: cloudprovider.AWS.String(),
		EC2Instances: ec2.Instances{
			"id-0": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.2",
			},
			"id-1": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.2",
			},
			"id-2": {
				PrivateIP: "192.0.2.1",
				PublicIP:  "192.0.2.2",
			},
		},
		EC2SecurityGroup: "sg-test",
	}
	someErr := errors.New("failed")
	config := config.Default()

	testCases := map[string]struct {
		existingState    *state.ConstellationState
		client           ec2client
		interactive      bool
		interactiveStdin string
		stateExpected    state.ConstellationState
		errExpected      bool
	}{
		"create some instances": {
			client:        &fakeEc2Client{},
			stateExpected: testState,
			errExpected:   false,
		},
		"state already exists": {
			existingState: &testState,
			client:        &fakeEc2Client{},
			errExpected:   true,
		},
		"create some instances interactive": {
			client:           &fakeEc2Client{},
			interactive:      true,
			interactiveStdin: "y\n",
			stateExpected:    testState,
			errExpected:      false,
		},
		"fail CreateSecurityGroup": {
			client:      &stubEc2Client{createSecurityGroupErr: someErr},
			errExpected: true,
		},
		"fail CreateInstances": {
			client:      &stubEc2Client{createInstancesErr: someErr},
			errExpected: true,
		},
		"fail GetState": {
			client:      &stubEc2Client{getStateErr: someErr},
			errExpected: true,
		},
		"error on rollback": {
			client:      &stubEc2Client{createInstancesErr: someErr, deleteSecurityGroupErr: someErr},
			errExpected: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newCreateAWSCmd()
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
				require.NoError(fileHandler.WriteJSON(*config.StatePath, *tc.existingState, false))
			}

			err := createAWS(cmd, tc.client, fileHandler, config, "xlarge", "name", 3)
			if tc.errExpected {
				assert.Error(err)
				if stubClient, ok := tc.client.(*stubEc2Client); ok {
					// Should have made a rollback on error.
					assert.True(stubClient.terminateInstancesCalled)
					assert.True(stubClient.deleteSecurityGroupCalled)
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

func TestCreateAWSCompletion(t *testing.T) {
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
			args:       []string{"23"},
			toComplete: "4xl",
			resultExpected: []string{
				"4xlarge",
				"8xlarge",
				"12xlarge",
				"16xlarge",
				"24xlarge",
			},
			shellCDExpected: cobra.ShellCompDirectiveDefault,
		},
		"third arg": {
			args:            []string{"23", "4xlarge"},
			toComplete:      "xl",
			resultExpected:  []string{},
			shellCDExpected: cobra.ShellCompDirectiveError,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := &cobra.Command{}
			result, shellCD := createAWSCompletion(cmd, tc.args, tc.toComplete)
			assert.Equal(tc.resultExpected, result)
			assert.Equal(tc.shellCDExpected, shellCD)
		})
	}
}
