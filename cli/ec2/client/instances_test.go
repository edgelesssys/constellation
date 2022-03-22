package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/edgelesssys/constellation/cli/ec2"
	"github.com/stretchr/testify/assert"
)

func TestCreateInstances(t *testing.T) {
	testInstances := []types.Instance{
		{
			InstanceId:       aws.String("id-1"),
			PublicIpAddress:  aws.String("192.0.2.1"),
			PrivateIpAddress: aws.String("192.0.2.2"),
			State:            &stateRunning,
		},
		{
			InstanceId:       aws.String("id-2"),
			PublicIpAddress:  aws.String("192.0.2.3"),
			PrivateIpAddress: aws.String("192.0.2.4"),
			State:            &stateRunning,
		},
		{
			InstanceId:       aws.String("id-3"),
			PublicIpAddress:  aws.String("192.0.2.5"),
			PrivateIpAddress: aws.String("192.0.2.6"),
			State:            &stateRunning,
		},
	}
	someErr := errors.New("failed")
	var noErr error

	testCases := map[string]struct {
		api           stubAPI
		instances     ec2.Instances
		securityGroup string
		errExpected   bool
		wantInstances ec2.Instances
	}{
		"create": {
			api:           stubAPI{instances: testInstances},
			securityGroup: "sg-test",
			wantInstances: ec2.Instances{
				"id-1": {PublicIP: "192.0.2.1", PrivateIP: "192.0.2.2"},
				"id-2": {PublicIP: "192.0.2.3", PrivateIP: "192.0.2.4"},
				"id-3": {PublicIP: "192.0.2.5", PrivateIP: "192.0.2.6"},
			},
		},
		"client already has instances": {
			api:           stubAPI{instances: testInstances},
			instances:     ec2.Instances{"id-4": {}, "id-5": {}},
			securityGroup: "sg-test",
			wantInstances: ec2.Instances{
				"id-1": {PublicIP: "192.0.2.1", PrivateIP: "192.0.2.2"},
				"id-2": {PublicIP: "192.0.2.3", PrivateIP: "192.0.2.4"},
				"id-3": {PublicIP: "192.0.2.5", PrivateIP: "192.0.2.6"},
				"id-4": {},
				"id-5": {},
			},
		},
		"client already has same instance id": {
			api:           stubAPI{instances: testInstances},
			instances:     ec2.Instances{"id-1": {}, "id-4": {}, "id-5": {}},
			securityGroup: "sg-test",
			errExpected:   false,
			wantInstances: ec2.Instances{
				"id-1": {PublicIP: "192.0.2.1", PrivateIP: "192.0.2.2"},
				"id-2": {PublicIP: "192.0.2.3", PrivateIP: "192.0.2.4"},
				"id-3": {PublicIP: "192.0.2.5", PrivateIP: "192.0.2.6"},
				"id-4": {},
				"id-5": {},
			},
		},
		"client has no security group": {
			api:         stubAPI{},
			errExpected: true,
		},
		"run API error": {
			api:           stubAPI{runInstancesErr: someErr},
			securityGroup: "sg-test",
			errExpected:   true,
		},
		"runDryRun API error": {
			api:           stubAPI{runInstancesDryRunErr: &someErr},
			securityGroup: "sg-test",
			errExpected:   true,
		},
		"runDryRun missing expected API error": {
			api:           stubAPI{runInstancesDryRunErr: &noErr},
			securityGroup: "sg-test",
			errExpected:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &Client{
				api:           tc.api,
				instances:     tc.instances,
				timeout:       time.Millisecond,
				securityGroup: tc.securityGroup,
			}
			if client.instances == nil {
				client.instances = make(map[string]ec2.Instance)
			}
			input := CreateInput{
				ImageId:      "test-image",
				InstanceType: "",
				Count:        13,
			}

			err := client.CreateInstances(context.Background(), input)

			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.ElementsMatch(tc.wantInstances.IDs(), client.instances.IDs())
				assert.ElementsMatch(tc.wantInstances.PublicIPs(), client.instances.PublicIPs())
				assert.ElementsMatch(tc.wantInstances.PrivateIPs(), client.instances.PrivateIPs())
			}
		})
	}
}

func TestTerminateInstances(t *testing.T) {
	testAWSInstances := []types.Instance{
		{InstanceId: aws.String("id-1"), State: &stateTerminated},
		{InstanceId: aws.String("id-2"), State: &stateTerminated},
		{InstanceId: aws.String("id-3"), State: &stateTerminated},
	}
	someErr := errors.New("failed")
	var noErr error

	testCases := map[string]struct {
		api         stubAPI
		instances   ec2.Instances
		errExpected bool
	}{
		"client with instances": {
			api:         stubAPI{instances: testAWSInstances},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: false,
		},
		"client no instances set": {
			api: stubAPI{},
		},
		"terminate API error": {
			api:         stubAPI{terminateInstancesErr: someErr},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: true,
		},
		"terminateDryRun API error": {
			api:         stubAPI{terminateInstancesDryRunErr: &someErr},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: true,
		},
		"terminateDryRun miss expected API error": {
			api:         stubAPI{terminateInstancesDryRunErr: &noErr},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &Client{
				api:       tc.api,
				instances: tc.instances,
				timeout:   time.Millisecond,
			}
			if client.instances == nil {
				client.instances = make(map[string]ec2.Instance)
			}

			err := client.TerminateInstances(context.Background())
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Empty(client.instances)
			}
		})
	}
}

func TestWaitStateRunning(t *testing.T) {
	testCases := map[string]struct {
		api         api
		instances   ec2.Instances
		errExpected bool
	}{
		"instances are running": {
			api: stubAPI{instances: []types.Instance{
				{
					InstanceId: aws.String("id-1"),
					State:      &stateRunning,
				},
				{
					InstanceId: aws.String("id-2"),
					State:      &stateRunning,
				},
				{
					InstanceId: aws.String("id-3"),
					State:      &stateRunning,
				},
			}},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: false,
		},
		"one instance running, rest nil": {
			api: stubAPI{instances: []types.Instance{
				{
					InstanceId: aws.String("id-1"),
					State:      &stateRunning,
				},
				{InstanceId: aws.String("id-2")},
				{InstanceId: aws.String("id-3")},
			}},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: false,
		},
		"one instance terminated, rest nil": {
			api: stubAPI{instances: []types.Instance{
				{
					InstanceId: aws.String("id-1"),
					State:      &stateTerminated,
				},
				{InstanceId: aws.String("id-2")},
				{InstanceId: aws.String("id-3")},
			}},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: true,
		},
		"instances with different state": {
			api: stubAPI{instances: []types.Instance{
				{
					InstanceId: aws.String("id-1"),
					State:      &stateTerminated,
				},
				{
					InstanceId: aws.String("id-2"),
					State:      &stateRunning,
				},
				{InstanceId: aws.String("id-3")},
			}},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: true,
		},
		"all instances have nil state": {
			api: stubAPI{instances: []types.Instance{
				{InstanceId: aws.String("id-1")},
				{InstanceId: aws.String("id-2")},
				{InstanceId: aws.String("id-3")},
			}},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: true,
		},
		"client has no instances": {
			api:         &stubAPI{},
			errExpected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			client := &Client{
				api:       tc.api,
				instances: tc.instances,
				timeout:   time.Millisecond,
			}
			if client.instances == nil {
				client.instances = make(map[string]ec2.Instance)
			}

			err := client.waitStateRunning(context.Background())
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestWaitStateTerminated(t *testing.T) {
	testCases := map[string]struct {
		api         api
		instances   ec2.Instances
		errExpected bool
	}{
		"instances are terminated": {
			api: stubAPI{instances: []types.Instance{
				{
					InstanceId: aws.String("id-1"),
					State:      &stateTerminated,
				},
				{
					InstanceId: aws.String("id-2"),
					State:      &stateTerminated,
				},
				{
					InstanceId: aws.String("id-3"),
					State:      &stateTerminated,
				},
			}},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: false,
		},
		"one instance terminated, rest nil": {
			api: stubAPI{instances: []types.Instance{
				{
					InstanceId: aws.String("id-1"),
					State:      &stateTerminated,
				},
				{InstanceId: aws.String("id-2")},
				{InstanceId: aws.String("id-3")},
			}},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: false,
		},
		"one instance running, rest nil": {
			api: stubAPI{instances: []types.Instance{
				{
					InstanceId: aws.String("id-1"),
					State:      &stateRunning,
				},
				{InstanceId: aws.String("id-2")},
				{InstanceId: aws.String("id-3")},
			}},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: true,
		},
		"instances with different state": {
			api: stubAPI{instances: []types.Instance{
				{
					InstanceId: aws.String("id-1"),
					State:      &stateTerminated,
				},
				{
					InstanceId: aws.String("id-2"),
					State:      &stateRunning,
				},
				{InstanceId: aws.String("id-3")},
			}},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: true,
		},
		"all instances have nil state": {
			api: stubAPI{instances: []types.Instance{
				{InstanceId: aws.String("id-1")},
				{InstanceId: aws.String("id-2")},
				{InstanceId: aws.String("id-3")},
			}},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: true,
		},
		"client has no instances": {
			api:         &stubAPI{},
			errExpected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &Client{
				api:       tc.api,
				instances: tc.instances,
				timeout:   time.Millisecond,
			}
			if client.instances == nil {
				client.instances = make(map[string]ec2.Instance)
			}

			err := client.waitStateTerminated(context.Background())
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestTagInstances(t *testing.T) {
	testTags := ec2.Tags{
		{Key: "Name", Value: "Test"},
		{Key: "Foo", Value: "Bar"},
	}

	testCases := map[string]struct {
		api         stubAPI
		instances   ec2.Instances
		errExpected bool
	}{
		"tag": {
			api:         stubAPI{},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: false,
		},
		"client without instances": {
			api:         stubAPI{createTagsErr: errors.New("failed")},
			errExpected: true,
		},
		"tag API error": {
			api:         stubAPI{createTagsErr: errors.New("failed")},
			instances:   ec2.Instances{"id-1": {}, "id-2": {}, "id-3": {}},
			errExpected: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &Client{
				api:       tc.api,
				instances: tc.instances,
				timeout:   time.Millisecond,
			}
			if client.instances == nil {
				client.instances = make(map[string]ec2.Instance)
			}

			err := client.tagInstances(context.Background(), testTags)
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestEc2RunInstanceInput(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		in          CreateInput
		outExpected awsec2.RunInstancesInput
	}{
		{
			in: CreateInput{
				ImageId:          "test-image",
				InstanceType:     "4xlarge",
				Count:            13,
				securityGroupIds: []string{"test-sec-group"},
			},
			outExpected: awsec2.RunInstancesInput{
				ImageId:          aws.String("test-image"),
				InstanceType:     types.InstanceTypeC5a4xlarge,
				MinCount:         aws.Int32(int32(13)),
				MaxCount:         aws.Int32(int32(13)),
				EnclaveOptions:   &types.EnclaveOptionsRequest{Enabled: aws.Bool(true)},
				SecurityGroupIds: []string{"test-sec-group"},
			},
		},
		{
			in: CreateInput{
				ImageId:          "test-image-2",
				InstanceType:     "12xlarge",
				Count:            2,
				securityGroupIds: []string{"test-sec-group-2"},
			},
			outExpected: awsec2.RunInstancesInput{
				ImageId:          aws.String("test-image-2"),
				InstanceType:     types.InstanceTypeC5a12xlarge,
				MinCount:         aws.Int32(int32(2)),
				MaxCount:         aws.Int32(int32(2)),
				EnclaveOptions:   &types.EnclaveOptionsRequest{Enabled: aws.Bool(true)},
				SecurityGroupIds: []string{"test-sec-group-2"},
			},
		},
	}

	for _, tc := range testCases {
		out := tc.in.AWS()
		assert.Equal(tc.outExpected, *out)
	}
}
