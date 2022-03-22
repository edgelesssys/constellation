package client

import (
	"testing"

	"github.com/edgelesssys/constellation/cli/ec2"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/stretchr/testify/assert"
)

func TestGetState(t *testing.T) {
	testCases := map[string]struct {
		client    Client
		wantState state.ConstellationState
		wantErr   bool
	}{
		"successful get": {
			client: Client{
				instances:     ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
				securityGroup: "sg",
			},
			wantState: state.ConstellationState{
				CloudProvider:    "AWS",
				EC2Instances:     ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
				EC2SecurityGroup: "sg",
			},
		},
		"client without security group": {
			client: Client{
				instances: ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			},
			wantErr: true,
		},
		"client without instances": {
			client: Client{
				securityGroup: "sg",
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			stat, err := tc.client.GetState()
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantState, stat)
			}
		})
	}
}

func TestSetState(t *testing.T) {
	testCases := map[string]struct {
		state             state.ConstellationState
		wantInstances     ec2.Instances
		wantSecurityGroup string
		wantErr           bool
	}{
		"successful set": {
			state: state.ConstellationState{
				CloudProvider:    "AWS",
				EC2SecurityGroup: "sg-test",
				EC2Instances:     ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			},
			wantInstances:     ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			wantSecurityGroup: "sg-test",
		},
		"state without cloudprovider": {
			state: state.ConstellationState{
				EC2SecurityGroup: "sg-test",
				EC2Instances:     ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			},
			wantErr: true,
		},
		"state with incorrect cloudprovider": {
			state: state.ConstellationState{
				CloudProvider:    "incorrect",
				EC2SecurityGroup: "sg-test",
				EC2Instances:     ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			},
			wantErr: true,
		},
		"state without instances": {
			state: state.ConstellationState{
				CloudProvider:    "AWS",
				EC2SecurityGroup: "sg-test",
			},
			wantErr: true,
		},
		"state without security group": {
			state: state.ConstellationState{
				CloudProvider: "AWS",
				EC2Instances:  ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &Client{}

			err := client.SetState(tc.state)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantInstances, client.instances)
				assert.Equal(tc.wantSecurityGroup, client.securityGroup)
			}
		})
	}
}
