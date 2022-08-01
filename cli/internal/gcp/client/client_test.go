package client

import (
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

func TestSetGetState(t *testing.T) {
	state := state.ConstellationState{
		CloudProvider: cloudprovider.GCP.String(),
		GCPWorkerInstances: cloudtypes.Instances{
			"id-1": {
				PublicIP:  "ip1",
				PrivateIP: "ip2",
			},
		},
		GCPControlPlaneInstances: cloudtypes.Instances{
			"id-1": {
				PublicIP:  "ip3",
				PrivateIP: "ip4",
			},
		},
		GCPWorkerInstanceGroup:          "group-id",
		GCPControlPlaneInstanceGroup:    "group-id",
		GCPProject:                      "proj-id",
		GCPZone:                         "zone-id",
		GCPRegion:                       "region-id",
		Name:                            "name",
		UID:                             "uid",
		BootstrapperHost:                "ip3",
		GCPNetwork:                      "net-id",
		GCPSubnetwork:                   "subnet-id",
		GCPFirewalls:                    []string{"fw-1", "fw-2"},
		GCPWorkerInstanceTemplate:       "temp-id",
		GCPControlPlaneInstanceTemplate: "temp-id",
		GCPServiceAccount:               "service-account",
		GCPBackendService:               "backend-service-id",
		GCPHealthCheck:                  "health-check-id",
		GCPForwardingRule:               "forwarding-rule-id",
	}

	t.Run("SetState", func(t *testing.T) {
		assert := assert.New(t)

		client := Client{}
		client.SetState(state)
		assert.Equal(state.GCPWorkerInstances, client.workers)
		assert.Equal(state.GCPControlPlaneInstances, client.controlPlanes)
		assert.Equal(state.GCPWorkerInstanceGroup, client.workerInstanceGroup)
		assert.Equal(state.GCPControlPlaneInstanceGroup, client.controlPlaneInstanceGroup)
		assert.Equal(state.GCPProject, client.project)
		assert.Equal(state.GCPZone, client.zone)
		assert.Equal(state.Name, client.name)
		assert.Equal(state.UID, client.uid)
		assert.Equal(state.GCPNetwork, client.network)
		assert.Equal(state.GCPFirewalls, client.firewalls)
		assert.Equal(state.GCPControlPlaneInstanceTemplate, client.controlPlaneTemplate)
		assert.Equal(state.GCPWorkerInstanceTemplate, client.workerTemplate)
		assert.Equal(state.GCPServiceAccount, client.serviceAccount)
	})

	t.Run("GetState", func(t *testing.T) {
		assert := assert.New(t)

		client := Client{
			workers:                   state.GCPWorkerInstances,
			controlPlanes:             state.GCPControlPlaneInstances,
			workerInstanceGroup:       state.GCPWorkerInstanceGroup,
			controlPlaneInstanceGroup: state.GCPControlPlaneInstanceGroup,
			project:                   state.GCPProject,
			zone:                      state.GCPZone,
			region:                    state.GCPRegion,
			name:                      state.Name,
			uid:                       state.UID,
			network:                   state.GCPNetwork,
			subnetwork:                state.GCPSubnetwork,
			firewalls:                 state.GCPFirewalls,
			workerTemplate:            state.GCPWorkerInstanceTemplate,
			controlPlaneTemplate:      state.GCPControlPlaneInstanceTemplate,
			serviceAccount:            state.GCPServiceAccount,
			healthCheck:               state.GCPHealthCheck,
			backendService:            state.GCPBackendService,
			forwardingRule:            state.GCPForwardingRule,
		}

		stat := client.GetState()

		assert.Equal(state, stat)
	})
}

func TestInit(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	client := Client{}
	require.NoError(client.init("project", "zone", "region", "name"))
	assert.Equal("project", client.project)
	assert.Equal("zone", client.zone)
	assert.Equal("region", client.region)
	assert.Equal("name", client.name)
}

func TestCloseAll(t *testing.T) {
	assert := assert.New(t)

	closers := []closer{&someCloser{}, &someCloser{}, &someCloser{}}
	assert.NoError(closeAll(closers))
	for _, c := range closers {
		assert.True(c.(*someCloser).closed)
	}

	someErr := errors.New("failed")
	closers = []closer{&someCloser{}, &someCloser{closeErr: someErr}, &someCloser{}}
	assert.Error(closeAll(closers))
	for _, c := range closers {
		assert.True(c.(*someCloser).closed)
	}
}

type someCloser struct {
	closeErr error
	closed   bool
}

func (c *someCloser) Close() error {
	c.closed = true
	return c.closeErr
}
