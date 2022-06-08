package client

import (
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetGetState(t *testing.T) {
	testCases := map[string]struct {
		state   state.ConstellationState
		wantErr bool
	}{
		"valid state": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPProject:                     "proj-id",
				GCPZone:                        "zone-id",
				GCPRegion:                      "region-id",
				Name:                           "name",
				UID:                            "uid",
				GCPNetwork:                     "net-id",
				GCPSubnetwork:                  "subnet-id",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
				GCPServiceAccount:              "service-account",
			},
		},
		"missing nodes": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPProject:                     "proj-id",
				GCPZone:                        "zone-id",
				GCPRegion:                      "region-id",
				Name:                           "name",
				UID:                            "uid",
				GCPNetwork:                     "net-id",
				GCPSubnetwork:                  "subnet-id",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing coordinator": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPProject:                     "proj-id",
				GCPZone:                        "zone-id",
				GCPRegion:                      "region-id",
				Name:                           "name",
				UID:                            "uid",
				GCPNetwork:                     "net-id",
				GCPSubnetwork:                  "subnet-id",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing node group": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPProject:                     "proj-id",
				GCPZone:                        "zone-id",
				GCPRegion:                      "region-id",
				Name:                           "name",
				UID:                            "uid",
				GCPNetwork:                     "net-id",
				GCPSubnetwork:                  "subnet-id",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing coordinator group": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPProject:                     "proj-id",
				GCPZone:                        "zone-id",
				GCPRegion:                      "region-id",
				Name:                           "name",
				UID:                            "uid",
				GCPNetwork:                     "net-id",
				GCPSubnetwork:                  "subnet-id",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing project id": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPZone:                        "zone-id",
				GCPRegion:                      "region-id",
				Name:                           "name",
				UID:                            "uid",
				GCPNetwork:                     "net-id",
				GCPSubnetwork:                  "subnet-id",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing zone": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPProject:                     "proj-id",
				GCPRegion:                      "region-id",
				Name:                           "name",
				UID:                            "uid",
				GCPNetwork:                     "net-id",
				GCPSubnetwork:                  "subnet-id",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing region": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPProject:                     "proj-id",
				GCPZone:                        "zone-id",
				Name:                           "name",
				UID:                            "uid",
				GCPNetwork:                     "net-id",
				GCPSubnetwork:                  "subnet-id",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing name": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPProject:                     "proj-id",
				GCPZone:                        "zone-id",
				UID:                            "uid",
				GCPRegion:                      "region-id",
				GCPNetwork:                     "net-id",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing uid": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPProject:                     "proj-id",
				GCPZone:                        "zone-id",
				Name:                           "name",
				GCPRegion:                      "region-id",
				GCPNetwork:                     "net-id",
				GCPSubnetwork:                  "subnet-id",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing firewalls": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPProject:                     "proj-id",
				GCPZone:                        "zone-id",
				GCPRegion:                      "region-id",
				Name:                           "name",
				UID:                            "uid",
				GCPNetwork:                     "net-id",
				GCPSubnetwork:                  "subnet-id",
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing network": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPProject:                     "proj-id",
				GCPZone:                        "zone-id",
				GCPRegion:                      "region-id",
				Name:                           "name",
				UID:                            "uid",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing external network": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPProject:                     "proj-id",
				GCPZone:                        "zone-id",
				GCPRegion:                      "region-id",
				Name:                           "name",
				UID:                            "uid",
				GCPNetwork:                     "net-id",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing subnetwork": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPProject:                     "proj-id",
				GCPZone:                        "zone-id",
				GCPRegion:                      "region-id",
				Name:                           "name",
				UID:                            "uid",
				GCPNetwork:                     "net-id",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing external subnetwork": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPProject:                     "proj-id",
				GCPZone:                        "zone-id",
				GCPRegion:                      "region-id",
				Name:                           "name",
				UID:                            "uid",
				GCPNetwork:                     "net-id",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:        "temp-id",
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing node template": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:           "group-id",
				GCPCoordinatorInstanceGroup:    "group-id",
				GCPProject:                     "proj-id",
				GCPZone:                        "zone-id",
				GCPRegion:                      "region-id",
				Name:                           "name",
				UID:                            "uid",
				GCPNetwork:                     "net-id",
				GCPSubnetwork:                  "subnet-id",
				GCPFirewalls:                   []string{"fw-1", "fw-2"},
				GCPCoordinatorInstanceTemplate: "temp-id",
			},
			wantErr: true,
		},
		"missing coordinator template": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPNodes: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
					},
				},
				GCPCoordinators: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip3",
						PrivateIP: "ip4",
					},
				},
				GCPNodeInstanceGroup:        "group-id",
				GCPCoordinatorInstanceGroup: "group-id",
				GCPProject:                  "proj-id",
				GCPZone:                     "zone-id",
				GCPRegion:                   "region-id",
				Name:                        "name",
				UID:                         "uid",
				GCPNetwork:                  "net-id",
				GCPSubnetwork:               "subnet-id",
				GCPFirewalls:                []string{"fw-1", "fw-2"},
				GCPNodeInstanceTemplate:     "temp-id",
			},
			wantErr: true,
		},
	}

	t.Run("SetState", func(t *testing.T) {
		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				client := Client{}
				if tc.wantErr {
					assert.Error(client.SetState(tc.state))
				} else {
					assert.NoError(client.SetState(tc.state))
					assert.Equal(tc.state.GCPNodes, client.nodes)
					assert.Equal(tc.state.GCPCoordinators, client.coordinators)
					assert.Equal(tc.state.GCPNodeInstanceGroup, client.nodesInstanceGroup)
					assert.Equal(tc.state.GCPCoordinatorInstanceGroup, client.coordinatorInstanceGroup)
					assert.Equal(tc.state.GCPProject, client.project)
					assert.Equal(tc.state.GCPZone, client.zone)
					assert.Equal(tc.state.Name, client.name)
					assert.Equal(tc.state.UID, client.uid)
					assert.Equal(tc.state.GCPNetwork, client.network)
					assert.Equal(tc.state.GCPFirewalls, client.firewalls)
					assert.Equal(tc.state.GCPCoordinatorInstanceTemplate, client.coordinatorTemplate)
					assert.Equal(tc.state.GCPNodeInstanceTemplate, client.nodeTemplate)
					assert.Equal(tc.state.GCPServiceAccount, client.serviceAccount)
				}
			})
		}
	})

	t.Run("GetState", func(t *testing.T) {
		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				client := Client{
					nodes:                    tc.state.GCPNodes,
					coordinators:             tc.state.GCPCoordinators,
					nodesInstanceGroup:       tc.state.GCPNodeInstanceGroup,
					coordinatorInstanceGroup: tc.state.GCPCoordinatorInstanceGroup,
					project:                  tc.state.GCPProject,
					zone:                     tc.state.GCPZone,
					region:                   tc.state.GCPRegion,
					name:                     tc.state.Name,
					uid:                      tc.state.UID,
					network:                  tc.state.GCPNetwork,
					subnetwork:               tc.state.GCPSubnetwork,
					firewalls:                tc.state.GCPFirewalls,
					nodeTemplate:             tc.state.GCPNodeInstanceTemplate,
					coordinatorTemplate:      tc.state.GCPCoordinatorInstanceTemplate,
					serviceAccount:           tc.state.GCPServiceAccount,
				}
				if tc.wantErr {
					_, err := client.GetState()
					assert.Error(err)
				} else {
					stat, err := client.GetState()
					assert.NoError(err)
					assert.Equal(tc.state, stat)
				}
			})
		}
	})
}

func TestSetStateCloudProvider(t *testing.T) {
	assert := assert.New(t)

	client := Client{}
	stateMissingCloudProvider := state.ConstellationState{
		GCPNodes: cloudtypes.Instances{
			"id-1": {
				PublicIP:  "ip1",
				PrivateIP: "ip2",
			},
		},
		GCPCoordinators: cloudtypes.Instances{
			"id-1": {
				PublicIP:  "ip3",
				PrivateIP: "ip4",
			},
		},
		GCPNodeInstanceGroup:           "group-id",
		GCPCoordinatorInstanceGroup:    "group-id",
		GCPProject:                     "proj-id",
		GCPZone:                        "zone-id",
		GCPRegion:                      "region-id",
		Name:                           "name",
		UID:                            "uid",
		GCPNetwork:                     "net-id",
		GCPSubnetwork:                  "subnet-id",
		GCPFirewalls:                   []string{"fw-1", "fw-2"},
		GCPNodeInstanceTemplate:        "temp-id",
		GCPCoordinatorInstanceTemplate: "temp-id",
	}
	assert.Error(client.SetState(stateMissingCloudProvider))
	stateIncorrectCloudProvider := state.ConstellationState{
		CloudProvider: "incorrect",
		GCPNodes: cloudtypes.Instances{
			"id-1": {
				PublicIP:  "ip1",
				PrivateIP: "ip2",
			},
		},
		GCPCoordinators: cloudtypes.Instances{
			"id-1": {
				PublicIP:  "ip3",
				PrivateIP: "ip4",
			},
		},
		GCPNodeInstanceGroup:           "group-id",
		GCPCoordinatorInstanceGroup:    "group-id",
		GCPProject:                     "proj-id",
		GCPZone:                        "zone-id",
		GCPRegion:                      "region-id",
		Name:                           "name",
		UID:                            "uid",
		GCPNetwork:                     "net-id",
		GCPSubnetwork:                  "subnet-id",
		GCPFirewalls:                   []string{"fw-1", "fw-2"},
		GCPNodeInstanceTemplate:        "temp-id",
		GCPCoordinatorInstanceTemplate: "temp-id",
	}
	assert.Error(client.SetState(stateIncorrectCloudProvider))
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

func TestComposedErr(t *testing.T) {
	assert := assert.New(t)

	noErrs := []error{nil, nil, nil}
	assert.NoError(composeErr(noErrs))

	someErrs := []error{
		errors.New("failed 4"),
		errors.New("failed 7"),
		nil,
		nil,
		errors.New("failed 9"),
	}
	err := composeErr(someErrs)
	assert.Error(err)
	assert.Contains(err.Error(), "4")
	assert.Contains(err.Error(), "7")
	assert.Contains(err.Error(), "9")
}
