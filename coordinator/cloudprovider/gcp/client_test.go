package gcp

import (
	"context"
	"errors"
	"testing"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/role"
	gax "github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		// https://github.com/kubernetes/klog/issues/282, https://github.com/kubernetes/klog/issues/188
		goleak.IgnoreTopFunction("k8s.io/klog/v2.(*loggingT).flushDaemon"),
	)
}

func TestRetrieveInstances(t *testing.T) {
	uid := "1234"
	someErr := errors.New("failed")
	newTestIter := func() *stubInstanceIterator {
		return &stubInstanceIterator{
			instances: []*computepb.Instance{
				{
					Name: proto.String("someInstance"),
					Metadata: &computepb.Metadata{
						Items: []*computepb.Items{
							{
								Key:   proto.String("ssh-keys"),
								Value: proto.String("bob:ssh-rsa bobskey"),
							},
							{
								Key:   proto.String("key-2"),
								Value: proto.String("value-2"),
							},
							{
								Key:   proto.String(core.ConstellationUIDMetadataKey),
								Value: proto.String(uid),
							},
							{
								Key:   proto.String(core.RoleMetadataKey),
								Value: proto.String(role.Coordinator.String()),
							},
						},
					},
					NetworkInterfaces: []*computepb.NetworkInterface{
						{NetworkIP: proto.String("192.0.2.0")},
					},
				},
			},
		}
	}

	testCases := map[string]struct {
		client              stubInstancesClient
		metadata            stubMetadataClient
		instanceIter        *stubInstanceIterator
		instanceIterMutator func(*stubInstanceIterator)
		expectedInstances   []core.Instance
		expectErr           bool
	}{
		"retrieve works": {
			client:       stubInstancesClient{},
			metadata:     stubMetadataClient{InstanceValue: uid},
			instanceIter: newTestIter(),
			expectedInstances: []core.Instance{
				{
					Name:       "someInstance",
					ProviderID: "gce://someProject/someZone/someInstance",
					Role:       role.Coordinator,
					IPs:        []string{"192.0.2.0"},
					SSHKeys:    map[string][]string{"bob": {"ssh-rsa bobskey"}},
				},
			},
		},
		"instance name is null": {
			client:              stubInstancesClient{},
			metadata:            stubMetadataClient{InstanceValue: uid},
			instanceIter:        newTestIter(),
			instanceIterMutator: func(sii *stubInstanceIterator) { sii.instances[0].Name = nil },
			expectErr:           true,
		},
		"no instance with network ip": {
			client:              stubInstancesClient{},
			metadata:            stubMetadataClient{InstanceValue: uid},
			instanceIter:        newTestIter(),
			instanceIterMutator: func(sii *stubInstanceIterator) { sii.instances[0].NetworkInterfaces = nil },
			expectedInstances: []core.Instance{
				{
					Name:       "someInstance",
					ProviderID: "gce://someProject/someZone/someInstance",
					Role:       role.Coordinator,
					IPs:        []string{},
					SSHKeys:    map[string][]string{"bob": {"ssh-rsa bobskey"}},
				},
			},
		},
		"network ip is nil": {
			client:              stubInstancesClient{},
			metadata:            stubMetadataClient{InstanceValue: uid},
			instanceIter:        newTestIter(),
			instanceIterMutator: func(sii *stubInstanceIterator) { sii.instances[0].NetworkInterfaces[0].NetworkIP = nil },
			expectedInstances: []core.Instance{
				{
					Name:       "someInstance",
					ProviderID: "gce://someProject/someZone/someInstance",
					Role:       role.Coordinator,
					IPs:        []string{},
					SSHKeys:    map[string][]string{"bob": {"ssh-rsa bobskey"}},
				},
			},
		},
		"constellation id is not set": {
			client:              stubInstancesClient{},
			metadata:            stubMetadataClient{InstanceValue: uid},
			instanceIter:        newTestIter(),
			instanceIterMutator: func(sii *stubInstanceIterator) { sii.instances[0].Metadata.Items[2].Key = proto.String("") },
			expectedInstances:   []core.Instance{},
		},
		"constellation retrieval fails": {
			client:       stubInstancesClient{},
			metadata:     stubMetadataClient{InstanceErr: someErr},
			instanceIter: newTestIter(),
			expectErr:    true,
		},
		"role is not set": {
			client:              stubInstancesClient{},
			metadata:            stubMetadataClient{InstanceValue: uid},
			instanceIter:        newTestIter(),
			instanceIterMutator: func(sii *stubInstanceIterator) { sii.instances[0].Metadata.Items[3].Key = proto.String("") },
			expectedInstances: []core.Instance{
				{
					Name:       "someInstance",
					ProviderID: "gce://someProject/someZone/someInstance",
					Role:       role.Unknown,
					IPs:        []string{"192.0.2.0"},
					SSHKeys:    map[string][]string{"bob": {"ssh-rsa bobskey"}},
				},
			},
		},
		"instance iterator Next() errors": {
			client:       stubInstancesClient{},
			metadata:     stubMetadataClient{InstanceValue: uid},
			instanceIter: &stubInstanceIterator{nextErr: someErr},
			expectErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			if tc.instanceIterMutator != nil {
				tc.instanceIterMutator(tc.instanceIter)
			}
			tc.client.ListInstanceIterator = tc.instanceIter
			client := Client{
				instanceAPI: tc.client,
				metadataAPI: tc.metadata,
			}

			instances, err := client.RetrieveInstances(context.Background(), "someProject", "someZone")

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedInstances, instances)
		})
	}
}

func TestRetrieveInstance(t *testing.T) {
	newTestInstance := func() *computepb.Instance {
		return &computepb.Instance{
			Name: proto.String("someInstance"),
			Metadata: &computepb.Metadata{
				Items: []*computepb.Items{
					{
						Key:   proto.String("key-1"),
						Value: proto.String("value-1"),
					},
					{
						Key:   proto.String("key-2"),
						Value: proto.String("value-2"),
					},
				},
			},
			NetworkInterfaces: []*computepb.NetworkInterface{
				{NetworkIP: proto.String("192.0.2.0")},
			},
		}
	}

	testCases := map[string]struct {
		client                stubInstancesClient
		clientInstance        *computepb.Instance
		clientInstanceMutator func(*computepb.Instance)
		expectedInstance      core.Instance
		expectErr             bool
	}{
		"retrieve works": {
			client:         stubInstancesClient{},
			clientInstance: newTestInstance(),
			expectedInstance: core.Instance{
				Name:       "someInstance",
				ProviderID: "gce://someProject/someZone/someInstance",
				IPs:        []string{"192.0.2.0"},
				SSHKeys:    map[string][]string{},
			},
		},
		"retrieve with SSH key works": {
			client:         stubInstancesClient{},
			clientInstance: newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) {
				i.Metadata.Items[0].Key = proto.String("ssh-keys")
				i.Metadata.Items[0].Value = proto.String("bob:ssh-rsa bobskey")
			},
			expectedInstance: core.Instance{
				Name:       "someInstance",
				ProviderID: "gce://someProject/someZone/someInstance",
				IPs:        []string{"192.0.2.0"},
				SSHKeys:    map[string][]string{"bob": {"ssh-rsa bobskey"}},
			},
		},
		"retrieve with Role works": {
			client:         stubInstancesClient{},
			clientInstance: newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) {
				i.Metadata.Items[0].Key = proto.String(core.RoleMetadataKey)
				i.Metadata.Items[0].Value = proto.String(role.Coordinator.String())
			},
			expectedInstance: core.Instance{
				Name:       "someInstance",
				ProviderID: "gce://someProject/someZone/someInstance",
				Role:       role.Coordinator,
				IPs:        []string{"192.0.2.0"},
				SSHKeys:    map[string][]string{},
			},
		},
		"retrieve fails": {
			client: stubInstancesClient{
				GetErr: errors.New("retrieve error"),
			},
			clientInstance: nil,
			expectErr:      true,
		},
		"metadata item is null": {
			client:                stubInstancesClient{},
			clientInstance:        newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) { i.Metadata.Items[0] = nil },
			expectedInstance: core.Instance{
				Name:       "someInstance",
				ProviderID: "gce://someProject/someZone/someInstance",
				IPs:        []string{"192.0.2.0"},
				SSHKeys:    map[string][]string{},
			},
		},
		"metadata key is null": {
			client:                stubInstancesClient{},
			clientInstance:        newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) { i.Metadata.Items[0].Key = nil },
			expectedInstance: core.Instance{
				Name:       "someInstance",
				ProviderID: "gce://someProject/someZone/someInstance",
				IPs:        []string{"192.0.2.0"},
				SSHKeys:    map[string][]string{},
			},
		},
		"metadata value is null": {
			client:                stubInstancesClient{},
			clientInstance:        newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) { i.Metadata.Items[0].Value = nil },
			expectedInstance: core.Instance{
				Name:       "someInstance",
				ProviderID: "gce://someProject/someZone/someInstance",
				IPs:        []string{"192.0.2.0"},
				SSHKeys:    map[string][]string{},
			},
		},
		"instance without network ip": {
			client:                stubInstancesClient{},
			clientInstance:        newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) { i.NetworkInterfaces[0] = nil },
			expectedInstance: core.Instance{
				Name:       "someInstance",
				ProviderID: "gce://someProject/someZone/someInstance",
				IPs:        []string{},
				SSHKeys:    map[string][]string{},
			},
		},
		"network ip is nil": {
			client:                stubInstancesClient{},
			clientInstance:        newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) { i.NetworkInterfaces[0].NetworkIP = nil },
			expectedInstance: core.Instance{
				Name:       "someInstance",
				ProviderID: "gce://someProject/someZone/someInstance",
				IPs:        []string{},
				SSHKeys:    map[string][]string{},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			if tc.clientInstanceMutator != nil {
				tc.clientInstanceMutator(tc.clientInstance)
			}
			tc.client.GetInstance = tc.clientInstance
			client := Client{instanceAPI: tc.client}

			instance, err := client.RetrieveInstance(context.Background(), "someProject", "someZone", "someInstance")

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedInstance, instance)
		})
	}
}

func TestRetrieveProjectID(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		client        stubMetadataClient
		expectedValue string
		expectErr     bool
	}{
		"retrieve works": {
			client:        stubMetadataClient{ProjectIDValue: "someProjectID"},
			expectedValue: "someProjectID",
		},
		"retrieve fails": {
			client:    stubMetadataClient{ProjectIDErr: someErr},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{metadataAPI: tc.client}
			value, err := client.RetrieveProjectID()

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedValue, value)
		})
	}
}

func TestRetrieveZone(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		client        stubMetadataClient
		expectedValue string
		expectErr     bool
	}{
		"retrieve works": {
			client:        stubMetadataClient{ZoneValue: "someZone"},
			expectedValue: "someZone",
		},
		"retrieve fails": {
			client:    stubMetadataClient{ZoneErr: someErr},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{metadataAPI: tc.client}
			value, err := client.RetrieveZone()

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedValue, value)
		})
	}
}

func TestRetrieveInstanceName(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		client        stubMetadataClient
		expectedValue string
		expectErr     bool
	}{
		"retrieve works": {
			client:        stubMetadataClient{InstanceNameValue: "someInstanceName"},
			expectedValue: "someInstanceName",
		},
		"retrieve fails": {
			client:    stubMetadataClient{InstanceNameErr: someErr},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{metadataAPI: tc.client}
			value, err := client.RetrieveInstanceName()

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedValue, value)
		})
	}
}

func TestRetrieveInstanceMetadata(t *testing.T) {
	someErr := errors.New("failed")
	attr := "someAttribute"

	testCases := map[string]struct {
		client        stubMetadataClient
		attr          string
		expectedValue string
		expectErr     bool
	}{
		"retrieve works": {
			client: stubMetadataClient{
				InstanceValue: "someValue",
				InstanceErr:   nil,
			},
			expectedValue: "someValue",
		},
		"retrieve fails": {
			client: stubMetadataClient{
				InstanceValue: "",
				InstanceErr:   someErr,
			},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{metadataAPI: tc.client}
			value, err := client.RetrieveInstanceMetadata(attr)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.expectedValue, value)
		})
	}
}

func TestSetInstanceMetadata(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		client    stubInstancesClient
		expectErr bool
	}{
		"set works": {
			client: stubInstancesClient{
				GetInstance: &computepb.Instance{
					Metadata: &computepb.Metadata{
						Fingerprint: proto.String("someFingerprint"),
						Kind:        proto.String("compute#metadata"),
						Items:       []*computepb.Items{},
					},
				},
			},
		},
		"retrieve fails": {
			client: stubInstancesClient{
				GetErr: someErr,
			},
			expectErr: true,
		},
		"retrieve returns nil": {
			expectErr: true,
		},
		"setting fails": {
			client: stubInstancesClient{
				GetInstance: &computepb.Instance{
					Metadata: &computepb.Metadata{
						Fingerprint: proto.String("someFingerprint"),
						Kind:        proto.String("compute#metadata"),
						Items:       []*computepb.Items{},
					},
				},
				SetMetadataErr: someErr,
			},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{instanceAPI: tc.client}
			err := client.SetInstanceMetadata(context.Background(), "project", "zone", "instanceName", "key", "value")

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestUnsetInstanceMetadata(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		client    stubInstancesClient
		expectErr bool
	}{
		"unset works": {
			client: stubInstancesClient{
				GetInstance: &computepb.Instance{
					Metadata: &computepb.Metadata{
						Fingerprint: proto.String("someFingerprint"),
						Kind:        proto.String("compute#metadata"),
						Items:       []*computepb.Items{},
					},
				},
			},
		},
		"unset with existing key works": {
			client: stubInstancesClient{
				GetInstance: &computepb.Instance{
					Metadata: &computepb.Metadata{
						Fingerprint: proto.String("someFingerprint"),
						Kind:        proto.String("compute#metadata"),
						Items: []*computepb.Items{
							{
								Key:   proto.String("key"),
								Value: proto.String("value"),
							},
						},
					},
				},
			},
		},
		"retrieve fails": {
			client:    stubInstancesClient{GetErr: someErr},
			expectErr: true,
		},
		"retrieve returns nil": {
			expectErr: true,
		},
		"setting fails": {
			client: stubInstancesClient{
				GetInstance: &computepb.Instance{
					Metadata: &computepb.Metadata{
						Fingerprint: proto.String("someFingerprint"),
						Kind:        proto.String("compute#metadata"),
						Items:       []*computepb.Items{},
					},
				},
				SetMetadataErr: someErr,
			},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{instanceAPI: tc.client}
			err := client.UnsetInstanceMetadata(context.Background(), "project", "zone", "instanceName", "key")

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestClose(t *testing.T) {
	someErr := errors.New("failed")

	assert := assert.New(t)

	client := Client{instanceAPI: stubInstancesClient{}}
	assert.NoError(client.Close())

	client = Client{instanceAPI: stubInstancesClient{CloseErr: someErr}}
	assert.Error(client.Close())
}

func TestFetchSSHKeys(t *testing.T) {
	testCases := map[string]struct {
		metadata     map[string]string
		expectedKeys map[string][]string
	}{
		"fetch works": {
			metadata:     map[string]string{"ssh-keys": "bob:ssh-rsa bobskey"},
			expectedKeys: map[string][]string{"bob": {"ssh-rsa bobskey"}},
		},
		"google ssh key metadata is ignored": {
			metadata:     map[string]string{"ssh-keys": "bob:ssh-rsa bobskey google-ssh {\"userName\":\"bob\",\"expireOn\":\"2021-06-14T16:59:03+0000\"}"},
			expectedKeys: map[string][]string{"bob": {"ssh-rsa bobskey"}},
		},
		"ssh key format error is ignored": {
			metadata:     map[string]string{"ssh-keys": "incorrect-format"},
			expectedKeys: map[string][]string{},
		},
		"ssh key format space error is ignored": {
			metadata:     map[string]string{"ssh-keys": "user:incorrect-key-format"},
			expectedKeys: map[string][]string{},
		},
		"metadata field empty": {
			metadata:     map[string]string{"ssh-keys": ""},
			expectedKeys: map[string][]string{},
		},
		"metadata field missing": {
			metadata:     map[string]string{},
			expectedKeys: map[string][]string{},
		},
		"multiple keys": {
			metadata: map[string]string{"ssh-keys": "bob:ssh-rsa bobskey\nalice:ssh-rsa alicekey"},
			expectedKeys: map[string][]string{
				"bob":   {"ssh-rsa bobskey"},
				"alice": {"ssh-rsa alicekey"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			keys := extractSSHKeys(tc.metadata)
			assert.Equal(tc.expectedKeys, keys)
		})
	}
}

type stubInstanceIterator struct {
	instances []*computepb.Instance
	nextErr   error

	internalCounter int
}

func (i *stubInstanceIterator) Next() (*computepb.Instance, error) {
	if i.nextErr != nil {
		return nil, i.nextErr
	}
	if i.internalCounter >= len(i.instances) {
		i.internalCounter = 0
		return nil, iterator.Done
	}
	resp := i.instances[i.internalCounter]
	i.internalCounter++
	return resp, nil
}

type stubInstancesClient struct {
	GetInstance          *computepb.Instance
	GetErr               error
	ListInstanceIterator InstanceIterator
	SetMetadataOperation *compute.Operation
	SetMetadataErr       error
	CloseErr             error
}

func (s stubInstancesClient) Get(ctx context.Context, req *computepb.GetInstanceRequest, opts ...gax.CallOption) (*computepb.Instance, error) {
	return s.GetInstance, s.GetErr
}

func (s stubInstancesClient) List(ctx context.Context, req *computepb.ListInstancesRequest, opts ...gax.CallOption) InstanceIterator {
	return s.ListInstanceIterator
}

func (s stubInstancesClient) SetMetadata(ctx context.Context, req *computepb.SetMetadataInstanceRequest, opts ...gax.CallOption) (*compute.Operation, error) {
	return s.SetMetadataOperation, s.SetMetadataErr
}

func (s stubInstancesClient) Close() error {
	return s.CloseErr
}

type stubMetadataClient struct {
	InstanceValue     string
	InstanceErr       error
	ProjectIDValue    string
	ProjectIDErr      error
	ZoneValue         string
	ZoneErr           error
	InstanceNameValue string
	InstanceNameErr   error
}

func (s stubMetadataClient) InstanceAttributeValue(attr string) (string, error) {
	return s.InstanceValue, s.InstanceErr
}

func (s stubMetadataClient) ProjectID() (string, error) {
	return s.ProjectIDValue, s.ProjectIDErr
}

func (s stubMetadataClient) Zone() (string, error) {
	return s.ZoneValue, s.ZoneErr
}

func (s stubMetadataClient) InstanceName() (string, error) {
	return s.InstanceNameValue, s.InstanceNameErr
}
