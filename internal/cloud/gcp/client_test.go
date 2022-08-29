package gcp

import (
	"context"
	"errors"
	"testing"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/edgelesssys/constellation/bootstrapper/role"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
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
								Key:   proto.String(constellationUIDMetadataKey),
								Value: proto.String(uid),
							},
							{
								Key:   proto.String(roleMetadataKey),
								Value: proto.String(role.ControlPlane.String()),
							},
						},
					},
					NetworkInterfaces: []*computepb.NetworkInterface{
						{
							Name:          proto.String("nic0"),
							NetworkIP:     proto.String("192.0.2.0"),
							AliasIpRanges: []*computepb.AliasIpRange{{IpCidrRange: proto.String("192.0.2.0/16")}},
							AccessConfigs: []*computepb.AccessConfig{{NatIP: proto.String("192.0.2.1")}},
						},
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
		wantInstances       []metadata.InstanceMetadata
		wantErr             bool
	}{
		"retrieve works": {
			client:       stubInstancesClient{},
			metadata:     stubMetadataClient{InstanceValue: uid},
			instanceIter: newTestIter(),
			wantInstances: []metadata.InstanceMetadata{
				{
					Name:          "someInstance",
					ProviderID:    "gce://someProject/someZone/someInstance",
					Role:          role.ControlPlane,
					AliasIPRanges: []string{"192.0.2.0/16"},
					PublicIP:      "192.0.2.1",
					VPCIP:         "192.0.2.0",
					SSHKeys:       map[string][]string{"bob": {"ssh-rsa bobskey"}},
				},
			},
		},
		"instance name is null": {
			client:              stubInstancesClient{},
			metadata:            stubMetadataClient{InstanceValue: uid},
			instanceIter:        newTestIter(),
			instanceIterMutator: func(sii *stubInstanceIterator) { sii.instances[0].Name = nil },
			wantErr:             true,
		},
		"no instance with network ip": {
			client:              stubInstancesClient{},
			metadata:            stubMetadataClient{InstanceValue: uid},
			instanceIter:        newTestIter(),
			instanceIterMutator: func(sii *stubInstanceIterator) { sii.instances[0].NetworkInterfaces = nil },
			wantInstances: []metadata.InstanceMetadata{
				{
					Name:          "someInstance",
					ProviderID:    "gce://someProject/someZone/someInstance",
					Role:          role.ControlPlane,
					AliasIPRanges: []string{},
					PublicIP:      "",
					VPCIP:         "",
					SSHKeys:       map[string][]string{"bob": {"ssh-rsa bobskey"}},
				},
			},
		},
		"network ip is nil": {
			client:              stubInstancesClient{},
			metadata:            stubMetadataClient{InstanceValue: uid},
			instanceIter:        newTestIter(),
			instanceIterMutator: func(sii *stubInstanceIterator) { sii.instances[0].NetworkInterfaces[0].NetworkIP = nil },
			wantInstances: []metadata.InstanceMetadata{
				{
					Name:          "someInstance",
					ProviderID:    "gce://someProject/someZone/someInstance",
					Role:          role.ControlPlane,
					AliasIPRanges: []string{"192.0.2.0/16"},
					PublicIP:      "192.0.2.1",
					VPCIP:         "",
					SSHKeys:       map[string][]string{"bob": {"ssh-rsa bobskey"}},
				},
			},
		},
		"constellation id is not set": {
			client:              stubInstancesClient{},
			metadata:            stubMetadataClient{InstanceValue: uid},
			instanceIter:        newTestIter(),
			instanceIterMutator: func(sii *stubInstanceIterator) { sii.instances[0].Metadata.Items[2].Key = proto.String("") },
			wantInstances:       []metadata.InstanceMetadata{},
		},
		"constellation retrieval fails": {
			client:       stubInstancesClient{},
			metadata:     stubMetadataClient{InstanceErr: someErr},
			instanceIter: newTestIter(),
			wantErr:      true,
		},
		"role is not set": {
			client:              stubInstancesClient{},
			metadata:            stubMetadataClient{InstanceValue: uid},
			instanceIter:        newTestIter(),
			instanceIterMutator: func(sii *stubInstanceIterator) { sii.instances[0].Metadata.Items[3].Key = proto.String("") },
			wantInstances: []metadata.InstanceMetadata{
				{
					Name:          "someInstance",
					ProviderID:    "gce://someProject/someZone/someInstance",
					Role:          role.Unknown,
					AliasIPRanges: []string{"192.0.2.0/16"},
					PublicIP:      "192.0.2.1",
					VPCIP:         "192.0.2.0",
					SSHKeys:       map[string][]string{"bob": {"ssh-rsa bobskey"}},
				},
			},
		},
		"instance iterator Next() errors": {
			client:       stubInstancesClient{},
			metadata:     stubMetadataClient{InstanceValue: uid},
			instanceIter: &stubInstanceIterator{nextErr: someErr},
			wantErr:      true,
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

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantInstances, instances)
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
				{
					Name:          proto.String("nic0"),
					NetworkIP:     proto.String("192.0.2.0"),
					AliasIpRanges: []*computepb.AliasIpRange{{IpCidrRange: proto.String("192.0.2.0/16")}},
					AccessConfigs: []*computepb.AccessConfig{{NatIP: proto.String("192.0.2.1")}},
				},
			},
		}
	}

	testCases := map[string]struct {
		client                stubInstancesClient
		clientInstance        *computepb.Instance
		clientInstanceMutator func(*computepb.Instance)
		wantInstance          metadata.InstanceMetadata
		wantErr               bool
	}{
		"retrieve works": {
			client:         stubInstancesClient{},
			clientInstance: newTestInstance(),
			wantInstance: metadata.InstanceMetadata{
				Name:          "someInstance",
				ProviderID:    "gce://someProject/someZone/someInstance",
				AliasIPRanges: []string{"192.0.2.0/16"},
				PublicIP:      "192.0.2.1",
				VPCIP:         "192.0.2.0",
				SSHKeys:       map[string][]string{},
			},
		},
		"retrieve with SSH key works": {
			client:         stubInstancesClient{},
			clientInstance: newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) {
				i.Metadata.Items[0].Key = proto.String("ssh-keys")
				i.Metadata.Items[0].Value = proto.String("bob:ssh-rsa bobskey")
			},
			wantInstance: metadata.InstanceMetadata{
				Name:          "someInstance",
				ProviderID:    "gce://someProject/someZone/someInstance",
				AliasIPRanges: []string{"192.0.2.0/16"},
				PublicIP:      "192.0.2.1",
				VPCIP:         "192.0.2.0",
				SSHKeys:       map[string][]string{"bob": {"ssh-rsa bobskey"}},
			},
		},
		"retrieve with Role works": {
			client:         stubInstancesClient{},
			clientInstance: newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) {
				i.Metadata.Items[0].Key = proto.String(roleMetadataKey)
				i.Metadata.Items[0].Value = proto.String(role.ControlPlane.String())
			},
			wantInstance: metadata.InstanceMetadata{
				Name:          "someInstance",
				ProviderID:    "gce://someProject/someZone/someInstance",
				AliasIPRanges: []string{"192.0.2.0/16"},
				PublicIP:      "192.0.2.1",
				Role:          role.ControlPlane,
				VPCIP:         "192.0.2.0",
				SSHKeys:       map[string][]string{},
			},
		},
		"retrieve fails": {
			client: stubInstancesClient{
				GetErr: errors.New("retrieve error"),
			},
			clientInstance: nil,
			wantErr:        true,
		},
		"metadata item is null": {
			client:                stubInstancesClient{},
			clientInstance:        newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) { i.Metadata.Items[0] = nil },
			wantInstance: metadata.InstanceMetadata{
				Name:          "someInstance",
				ProviderID:    "gce://someProject/someZone/someInstance",
				AliasIPRanges: []string{"192.0.2.0/16"},
				PublicIP:      "192.0.2.1",
				VPCIP:         "192.0.2.0",
				SSHKeys:       map[string][]string{},
			},
		},
		"metadata key is null": {
			client:                stubInstancesClient{},
			clientInstance:        newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) { i.Metadata.Items[0].Key = nil },
			wantInstance: metadata.InstanceMetadata{
				Name:          "someInstance",
				ProviderID:    "gce://someProject/someZone/someInstance",
				AliasIPRanges: []string{"192.0.2.0/16"},
				PublicIP:      "192.0.2.1",
				VPCIP:         "192.0.2.0",
				SSHKeys:       map[string][]string{},
			},
		},
		"metadata value is null": {
			client:                stubInstancesClient{},
			clientInstance:        newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) { i.Metadata.Items[0].Value = nil },
			wantInstance: metadata.InstanceMetadata{
				Name:          "someInstance",
				ProviderID:    "gce://someProject/someZone/someInstance",
				AliasIPRanges: []string{"192.0.2.0/16"},
				PublicIP:      "192.0.2.1",
				VPCIP:         "192.0.2.0",
				SSHKeys:       map[string][]string{},
			},
		},
		"instance without network ip": {
			client:                stubInstancesClient{},
			clientInstance:        newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) { i.NetworkInterfaces[0] = nil },
			wantInstance: metadata.InstanceMetadata{
				Name:          "someInstance",
				ProviderID:    "gce://someProject/someZone/someInstance",
				AliasIPRanges: []string{},
				PublicIP:      "",
				VPCIP:         "",
				SSHKeys:       map[string][]string{},
			},
		},
		"network ip is nil": {
			client:                stubInstancesClient{},
			clientInstance:        newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) { i.NetworkInterfaces[0].NetworkIP = nil },
			wantInstance: metadata.InstanceMetadata{
				Name:          "someInstance",
				ProviderID:    "gce://someProject/someZone/someInstance",
				AliasIPRanges: []string{"192.0.2.0/16"},
				PublicIP:      "192.0.2.1",
				VPCIP:         "",
				SSHKeys:       map[string][]string{},
			},
		},
		"network alias cidr is nil": {
			client:                stubInstancesClient{},
			clientInstance:        newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) { i.NetworkInterfaces[0].AliasIpRanges[0].IpCidrRange = nil },
			wantInstance: metadata.InstanceMetadata{
				Name:          "someInstance",
				ProviderID:    "gce://someProject/someZone/someInstance",
				AliasIPRanges: []string{},
				PublicIP:      "192.0.2.1",
				VPCIP:         "192.0.2.0",
				SSHKeys:       map[string][]string{},
			},
		},
		"network public ip is nil": {
			client:                stubInstancesClient{},
			clientInstance:        newTestInstance(),
			clientInstanceMutator: func(i *computepb.Instance) { i.NetworkInterfaces[0].AccessConfigs[0].NatIP = nil },
			wantInstance: metadata.InstanceMetadata{
				Name:          "someInstance",
				ProviderID:    "gce://someProject/someZone/someInstance",
				AliasIPRanges: []string{"192.0.2.0/16"},
				PublicIP:      "",
				VPCIP:         "192.0.2.0",
				SSHKeys:       map[string][]string{},
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

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantInstance, instance)
		})
	}
}

func TestRetrieveProjectID(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		client    stubMetadataClient
		wantValue string
		wantErr   bool
	}{
		"retrieve works": {
			client:    stubMetadataClient{ProjectIDValue: "someProjectID"},
			wantValue: "someProjectID",
		},
		"retrieve fails": {
			client:  stubMetadataClient{ProjectIDErr: someErr},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{metadataAPI: tc.client}
			value, err := client.RetrieveProjectID()

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantValue, value)
		})
	}
}

func TestRetrieveZone(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		client    stubMetadataClient
		wantValue string
		wantErr   bool
	}{
		"retrieve works": {
			client:    stubMetadataClient{ZoneValue: "someZone"},
			wantValue: "someZone",
		},
		"retrieve fails": {
			client:  stubMetadataClient{ZoneErr: someErr},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{metadataAPI: tc.client}
			value, err := client.RetrieveZone()

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantValue, value)
		})
	}
}

func TestRetrieveInstanceName(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		client    stubMetadataClient
		wantValue string
		wantErr   bool
	}{
		"retrieve works": {
			client:    stubMetadataClient{InstanceNameValue: "someInstanceName"},
			wantValue: "someInstanceName",
		},
		"retrieve fails": {
			client:  stubMetadataClient{InstanceNameErr: someErr},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{metadataAPI: tc.client}
			value, err := client.RetrieveInstanceName()

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantValue, value)
		})
	}
}

func TestRetrieveInstanceMetadata(t *testing.T) {
	someErr := errors.New("failed")
	attr := "someAttribute"

	testCases := map[string]struct {
		client    stubMetadataClient
		attr      string
		wantValue string
		wantErr   bool
	}{
		"retrieve works": {
			client: stubMetadataClient{
				InstanceValue: "someValue",
				InstanceErr:   nil,
			},
			wantValue: "someValue",
		},
		"retrieve fails": {
			client: stubMetadataClient{
				InstanceValue: "",
				InstanceErr:   someErr,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{metadataAPI: tc.client}
			value, err := client.RetrieveInstanceMetadata(attr)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantValue, value)
		})
	}
}

func TestSetInstanceMetadata(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		client  stubInstancesClient
		wantErr bool
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
			wantErr: true,
		},
		"retrieve returns nil": {
			wantErr: true,
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
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{instanceAPI: tc.client}
			err := client.SetInstanceMetadata(context.Background(), "project", "zone", "instanceName", "key", "value")

			if tc.wantErr {
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
		client  stubInstancesClient
		wantErr bool
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
			client:  stubInstancesClient{GetErr: someErr},
			wantErr: true,
		},
		"retrieve returns nil": {
			wantErr: true,
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
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{instanceAPI: tc.client}
			err := client.UnsetInstanceMetadata(context.Background(), "project", "zone", "instanceName", "key")

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestRetrieveSubnetworkAliasCIDR(t *testing.T) {
	aliasCIDR := "192.0.2.1/24"
	someErr := errors.New("some error")
	testCases := map[string]struct {
		stubInstancesClient   stubInstancesClient
		stubSubnetworksClient stubSubnetworksClient
		wantAliasCIDR         string
		wantErr               bool
	}{
		"RetrieveSubnetworkAliasCIDR works": {
			stubInstancesClient: stubInstancesClient{
				GetInstance: &computepb.Instance{
					NetworkInterfaces: []*computepb.NetworkInterface{
						{
							Subnetwork: proto.String("projects/project/regions/region/subnetworks/subnetwork"),
						},
					},
				},
			},
			stubSubnetworksClient: stubSubnetworksClient{
				GetSubnetwork: &computepb.Subnetwork{
					SecondaryIpRanges: []*computepb.SubnetworkSecondaryRange{
						{
							IpCidrRange: proto.String(aliasCIDR),
						},
					},
				},
			},
			wantAliasCIDR: aliasCIDR,
		},
		"instance has no network interface": {
			stubInstancesClient: stubInstancesClient{
				GetInstance: &computepb.Instance{
					NetworkInterfaces: []*computepb.NetworkInterface{},
				},
			},
			wantErr: true,
		},
		"cannot get instance": {
			stubInstancesClient: stubInstancesClient{
				GetErr: someErr,
			},
			wantErr: true,
		},
		"cannot get subnetwork": {
			stubInstancesClient: stubInstancesClient{
				GetInstance: &computepb.Instance{
					NetworkInterfaces: []*computepb.NetworkInterface{
						{
							Subnetwork: proto.String("projects/project/regions/region/subnetworks/subnetwork"),
						},
					},
				},
			},
			stubSubnetworksClient: stubSubnetworksClient{
				GetErr: someErr,
			},
			wantErr: true,
		},
		"subnetwork has no cidr range": {
			stubInstancesClient: stubInstancesClient{
				GetInstance: &computepb.Instance{
					NetworkInterfaces: []*computepb.NetworkInterface{
						{
							Subnetwork: proto.String("projects/project/regions/region/subnetworks/subnetwork"),
						},
					},
				},
			},
			stubSubnetworksClient: stubSubnetworksClient{
				GetSubnetwork: &computepb.Subnetwork{},
			},
			wantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{instanceAPI: tc.stubInstancesClient, subnetworkAPI: tc.stubSubnetworksClient}
			aliasCIDR, err := client.RetrieveSubnetworkAliasCIDR(context.Background(), "project", "us-central1-a", "subnetwork")

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantAliasCIDR, aliasCIDR)
		})
	}
}

func TestRetrieveLoadBalancerEndpoint(t *testing.T) {
	loadBalancerIP := "192.0.2.1"
	uid := "uid"
	someErr := errors.New("some error")
	testCases := map[string]struct {
		stubForwardingRulesClient stubForwardingRulesClient
		stubMetadataClient        stubMetadataClient
		wantLoadBalancerIP        string
		wantErr                   bool
	}{
		"works": {
			stubMetadataClient: stubMetadataClient{InstanceValue: uid},
			stubForwardingRulesClient: stubForwardingRulesClient{
				ForwardingRuleIterator: &stubForwardingRuleIterator{
					rules: []*computepb.ForwardingRule{
						{
							IPAddress: proto.String(loadBalancerIP),
							Ports:     []string{"100"},
							Labels:    map[string]string{"constellation-uid": uid},
						},
					},
				},
			},
			wantLoadBalancerIP: loadBalancerIP,
		},
		"fails when no matching load balancers exists": {
			stubMetadataClient: stubMetadataClient{InstanceValue: uid},
			stubForwardingRulesClient: stubForwardingRulesClient{
				ForwardingRuleIterator: &stubForwardingRuleIterator{
					rules: []*computepb.ForwardingRule{
						{
							IPAddress: proto.String(loadBalancerIP),
							Ports:     []string{"100"},
						},
					},
				},
			},
			wantErr: true,
		},
		"fails when retrieving uid": {
			stubMetadataClient: stubMetadataClient{InstanceErr: someErr},
			stubForwardingRulesClient: stubForwardingRulesClient{
				ForwardingRuleIterator: &stubForwardingRuleIterator{
					rules: []*computepb.ForwardingRule{
						{
							IPAddress: proto.String(loadBalancerIP),
							Ports:     []string{"100"},
							Labels:    map[string]string{"constellation-uid": uid},
						},
					},
				},
			},
			wantErr: true,
		},
		"fails when answer has empty port range": {
			stubMetadataClient: stubMetadataClient{InstanceErr: someErr},
			stubForwardingRulesClient: stubForwardingRulesClient{
				ForwardingRuleIterator: &stubForwardingRuleIterator{
					rules: []*computepb.ForwardingRule{
						{
							IPAddress: proto.String(loadBalancerIP),
							Labels:    map[string]string{"constellation-uid": uid},
						},
					},
				},
			},
			wantErr: true,
		},
		"fails when retrieving loadbalancer IP": {
			stubMetadataClient: stubMetadataClient{},
			stubForwardingRulesClient: stubForwardingRulesClient{
				ForwardingRuleIterator: &stubForwardingRuleIterator{
					nextErr: someErr,
					rules: []*computepb.ForwardingRule{
						{
							IPAddress: proto.String(loadBalancerIP),
							Ports:     []string{"100"},
							Labels:    map[string]string{"constellation-uid": uid},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{forwardingRulesAPI: tc.stubForwardingRulesClient, metadataAPI: tc.stubMetadataClient}
			aliasCIDR, err := client.RetrieveLoadBalancerEndpoint(context.Background(), "project", "us-central1-a")

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantLoadBalancerIP+":100", aliasCIDR)
		})
	}
}

func TestClose(t *testing.T) {
	someErr := errors.New("failed")

	assert := assert.New(t)

	client := Client{instanceAPI: stubInstancesClient{}, subnetworkAPI: stubSubnetworksClient{}, forwardingRulesAPI: stubForwardingRulesClient{}}
	assert.NoError(client.Close())

	client = Client{instanceAPI: stubInstancesClient{CloseErr: someErr}, subnetworkAPI: stubSubnetworksClient{}, forwardingRulesAPI: stubForwardingRulesClient{}}
	assert.Error(client.Close())

	client = Client{instanceAPI: stubInstancesClient{}, subnetworkAPI: stubSubnetworksClient{CloseErr: someErr}, forwardingRulesAPI: stubForwardingRulesClient{}}
	assert.Error(client.Close())

	client = Client{instanceAPI: stubInstancesClient{}, subnetworkAPI: stubSubnetworksClient{}, forwardingRulesAPI: stubForwardingRulesClient{CloseErr: someErr}}
	assert.Error(client.Close())
}

func TestFetchSSHKeys(t *testing.T) {
	testCases := map[string]struct {
		metadata map[string]string
		wantKeys map[string][]string
	}{
		"fetch works": {
			metadata: map[string]string{"ssh-keys": "bob:ssh-rsa bobskey"},
			wantKeys: map[string][]string{"bob": {"ssh-rsa bobskey"}},
		},
		"google ssh key metadata is ignored": {
			metadata: map[string]string{"ssh-keys": "bob:ssh-rsa bobskey google-ssh {\"userName\":\"bob\",\"expireOn\":\"2021-06-14T16:59:03+0000\"}"},
			wantKeys: map[string][]string{"bob": {"ssh-rsa bobskey"}},
		},
		"ssh key format error is ignored": {
			metadata: map[string]string{"ssh-keys": "incorrect-format"},
			wantKeys: map[string][]string{},
		},
		"ssh key format space error is ignored": {
			metadata: map[string]string{"ssh-keys": "user:incorrect-key-format"},
			wantKeys: map[string][]string{},
		},
		"metadata field empty": {
			metadata: map[string]string{"ssh-keys": ""},
			wantKeys: map[string][]string{},
		},
		"metadata field missing": {
			metadata: map[string]string{},
			wantKeys: map[string][]string{},
		},
		"multiple keys": {
			metadata: map[string]string{"ssh-keys": "bob:ssh-rsa bobskey\nalice:ssh-rsa alicekey"},
			wantKeys: map[string][]string{
				"bob":   {"ssh-rsa bobskey"},
				"alice": {"ssh-rsa alicekey"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			keys := extractSSHKeys(tc.metadata)
			assert.Equal(tc.wantKeys, keys)
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

type stubSubnetworksClient struct {
	GetSubnetwork      *computepb.Subnetwork
	GetErr             error
	SubnetworkIterator SubnetworkIterator
	CloseErr           error
}

func (s stubSubnetworksClient) Get(ctx context.Context, req *computepb.GetSubnetworkRequest, opts ...gax.CallOption) (*computepb.Subnetwork, error) {
	return s.GetSubnetwork, s.GetErr
}

func (s stubSubnetworksClient) List(ctx context.Context, req *computepb.ListSubnetworksRequest, opts ...gax.CallOption) SubnetworkIterator {
	return s.SubnetworkIterator
}

func (s stubSubnetworksClient) Close() error {
	return s.CloseErr
}

type stubForwardingRuleIterator struct {
	rules   []*computepb.ForwardingRule
	nextErr error

	internalCounter int
}

func (i *stubForwardingRuleIterator) Next() (*computepb.ForwardingRule, error) {
	if i.nextErr != nil {
		return nil, i.nextErr
	}
	if i.internalCounter >= len(i.rules) {
		i.internalCounter = 0
		return nil, iterator.Done
	}
	resp := i.rules[i.internalCounter]
	i.internalCounter++
	return resp, nil
}

type stubForwardingRulesClient struct {
	ForwardingRuleIterator ForwardingRuleIterator
	GetErr                 error
	CloseErr               error
}

func (s stubForwardingRulesClient) List(ctx context.Context, req *computepb.ListForwardingRulesRequest, opts ...gax.CallOption) ForwardingRuleIterator {
	return s.ForwardingRuleIterator
}

func (s stubForwardingRulesClient) Close() error {
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
