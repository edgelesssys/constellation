package client

import (
	"errors"
	"net/http"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/api/googleapi"
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
		LoadBalancerIP:                  "ip5",
		GCPNetwork:                      "net-id",
		GCPSubnetwork:                   "subnet-id",
		GCPFirewalls:                    []string{"fw-1", "fw-2"},
		GCPWorkerInstanceTemplate:       "temp-id",
		GCPControlPlaneInstanceTemplate: "temp-id",
		GCPLoadbalancers:                []string{"lb-1", "lb-2"},
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
		assert.Equal(state.LoadBalancerIP, client.loadbalancerIP)
		for _, lb := range client.loadbalancers {
			assert.Contains(state.GCPLoadbalancers, lb.name)
			assert.True(lb.hasBackendService)
			assert.True(lb.hasHealthCheck)
			assert.True(lb.hasForwardingRules)
		}
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
			loadbalancerIP:            state.LoadBalancerIP,
			loadbalancerIPname:        state.GCPLoadbalancerIPname,
		}
		for _, lbName := range state.GCPLoadbalancers {
			client.loadbalancers = append(client.loadbalancers, &loadBalancer{name: lbName})
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

func TestResourceURI(t *testing.T) {
	testCases := map[string]struct {
		scope        resourceScope
		resourceType string
		resourceName string
		wantURI      string
	}{
		"global resource": {
			scope:        scopeGlobal,
			resourceType: "healthChecks",
			resourceName: "name",
			wantURI:      "https://www.googleapis.com/compute/v1/projects/project/global/healthChecks/name",
		},
		"regional resource": {
			scope:        scopeRegion,
			resourceType: "healthChecks",
			resourceName: "name",
			wantURI:      "https://www.googleapis.com/compute/v1/projects/project/regions/region/healthChecks/name",
		},
		"zonal resource": {
			scope:        scopeZone,
			resourceType: "instanceGroups",
			resourceName: "name",
			wantURI:      "https://www.googleapis.com/compute/v1/projects/project/zones/zone/instanceGroups/name",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := Client{
				project: "project",
				zone:    "zone",
				region:  "region",
			}

			uri := client.resourceURI(tc.scope, tc.resourceType, tc.resourceName)

			assert.Equal(tc.wantURI, uri)
		})
	}
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

func TestIsNotFoundError(t *testing.T) {
	testCases := map[string]struct {
		err    error
		result bool
	}{
		"not found error": {err: &googleapi.Error{Code: http.StatusNotFound}, result: true},
		"nil error":       {err: nil, result: false},
		"other error":     {err: errors.New("failed"), result: false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.result, isNotFoundError(tc.err))
		})
	}
}
