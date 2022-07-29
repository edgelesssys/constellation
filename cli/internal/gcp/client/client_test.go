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
	testCases := map[string]struct {
		state   state.ConstellationState
		wantErr bool
	}{
		"valid state": {
			state: state.ConstellationState{
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
			},
		},
		"missing workers": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
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
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing control plane": {
			state: state.ConstellationState{
				CloudProvider: cloudprovider.GCP.String(),
				GCPWorkerInstances: cloudtypes.Instances{
					"id-1": {
						PublicIP:  "ip1",
						PrivateIP: "ip2",
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
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing worker group": {
			state: state.ConstellationState{
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
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing control plane group": {
			state: state.ConstellationState{
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
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing project id": {
			state: state.ConstellationState{
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
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing zone": {
			state: state.ConstellationState{
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
				GCPRegion:                       "region-id",
				Name:                            "name",
				UID:                             "uid",
				BootstrapperHost:                "ip3",
				GCPNetwork:                      "net-id",
				GCPSubnetwork:                   "subnet-id",
				GCPFirewalls:                    []string{"fw-1", "fw-2"},
				GCPWorkerInstanceTemplate:       "temp-id",
				GCPControlPlaneInstanceTemplate: "temp-id",
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing region": {
			state: state.ConstellationState{
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
				Name:                            "name",
				UID:                             "uid",
				BootstrapperHost:                "ip3",
				GCPNetwork:                      "net-id",
				GCPSubnetwork:                   "subnet-id",
				GCPFirewalls:                    []string{"fw-1", "fw-2"},
				GCPWorkerInstanceTemplate:       "temp-id",
				GCPControlPlaneInstanceTemplate: "temp-id",
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing name": {
			state: state.ConstellationState{
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
				UID:                             "uid",
				BootstrapperHost:                "ip3",
				GCPRegion:                       "region-id",
				GCPNetwork:                      "net-id",
				GCPFirewalls:                    []string{"fw-1", "fw-2"},
				GCPWorkerInstanceTemplate:       "temp-id",
				GCPControlPlaneInstanceTemplate: "temp-id",
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing uid": {
			state: state.ConstellationState{
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
				Name:                            "name",
				BootstrapperHost:                "ip3",
				GCPRegion:                       "region-id",
				GCPNetwork:                      "net-id",
				GCPSubnetwork:                   "subnet-id",
				GCPFirewalls:                    []string{"fw-1", "fw-2"},
				GCPWorkerInstanceTemplate:       "temp-id",
				GCPControlPlaneInstanceTemplate: "temp-id",
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing firewalls": {
			state: state.ConstellationState{
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
				GCPWorkerInstanceTemplate:       "temp-id",
				GCPControlPlaneInstanceTemplate: "temp-id",
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing network": {
			state: state.ConstellationState{
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
				GCPFirewalls:                    []string{"fw-1", "fw-2"},
				GCPWorkerInstanceTemplate:       "temp-id",
				GCPControlPlaneInstanceTemplate: "temp-id",
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing external network": {
			state: state.ConstellationState{
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
				GCPFirewalls:                    []string{"fw-1", "fw-2"},
				GCPWorkerInstanceTemplate:       "temp-id",
				GCPControlPlaneInstanceTemplate: "temp-id",
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing subnetwork": {
			state: state.ConstellationState{
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
				GCPFirewalls:                    []string{"fw-1", "fw-2"},
				GCPWorkerInstanceTemplate:       "temp-id",
				GCPControlPlaneInstanceTemplate: "temp-id",
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing external subnetwork": {
			state: state.ConstellationState{
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
				GCPFirewalls:                    []string{"fw-1", "fw-2"},
				GCPWorkerInstanceTemplate:       "temp-id",
				GCPControlPlaneInstanceTemplate: "temp-id",
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing worker template": {
			state: state.ConstellationState{
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
				GCPControlPlaneInstanceTemplate: "temp-id",
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing control plane template": {
			state: state.ConstellationState{
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
				GCPWorkerInstanceGroup:       "group-id",
				GCPControlPlaneInstanceGroup: "group-id",
				GCPProject:                   "proj-id",
				GCPZone:                      "zone-id",
				GCPRegion:                    "region-id",
				Name:                         "name",
				UID:                          "uid",
				BootstrapperHost:             "ip3",
				GCPNetwork:                   "net-id",
				GCPSubnetwork:                "subnet-id",
				GCPFirewalls:                 []string{"fw-1", "fw-2"},
				GCPWorkerInstanceTemplate:    "temp-id",
				GCPBackendService:            "backend-service-id",
				GCPHealthCheck:               "health-check-id",
				GCPForwardingRule:            "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing backend service": {
			state: state.ConstellationState{
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
				GCPHealthCheck:                  "health-check-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing health check": {
			state: state.ConstellationState{
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
				GCPBackendService:               "backend-service-id",
				GCPForwardingRule:               "forwarding-rule-id",
			},
			wantErr: true,
		},
		"missing forwarding rule": {
			state: state.ConstellationState{
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
				GCPBackendService:               "backend-service-id",
				GCPHealthCheck:                  "health-check-id",
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
					assert.Equal(tc.state.GCPWorkerInstances, client.workers)
					assert.Equal(tc.state.GCPControlPlaneInstances, client.controlPlanes)
					assert.Equal(tc.state.GCPWorkerInstanceGroup, client.workerInstanceGroup)
					assert.Equal(tc.state.GCPControlPlaneInstanceGroup, client.controlPlaneInstanceGroup)
					assert.Equal(tc.state.GCPProject, client.project)
					assert.Equal(tc.state.GCPZone, client.zone)
					assert.Equal(tc.state.Name, client.name)
					assert.Equal(tc.state.UID, client.uid)
					assert.Equal(tc.state.GCPNetwork, client.network)
					assert.Equal(tc.state.GCPFirewalls, client.firewalls)
					assert.Equal(tc.state.GCPControlPlaneInstanceTemplate, client.controlPlaneTemplate)
					assert.Equal(tc.state.GCPWorkerInstanceTemplate, client.workerTemplate)
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
					workers:                   tc.state.GCPWorkerInstances,
					controlPlanes:             tc.state.GCPControlPlaneInstances,
					workerInstanceGroup:       tc.state.GCPWorkerInstanceGroup,
					controlPlaneInstanceGroup: tc.state.GCPControlPlaneInstanceGroup,
					project:                   tc.state.GCPProject,
					zone:                      tc.state.GCPZone,
					region:                    tc.state.GCPRegion,
					name:                      tc.state.Name,
					uid:                       tc.state.UID,
					network:                   tc.state.GCPNetwork,
					subnetwork:                tc.state.GCPSubnetwork,
					firewalls:                 tc.state.GCPFirewalls,
					workerTemplate:            tc.state.GCPWorkerInstanceTemplate,
					controlPlaneTemplate:      tc.state.GCPControlPlaneInstanceTemplate,
					serviceAccount:            tc.state.GCPServiceAccount,
					healthCheck:               tc.state.GCPHealthCheck,
					backendService:            tc.state.GCPBackendService,
					forwardingRule:            tc.state.GCPForwardingRule,
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
		GCPBackendService:               "backend-service-id",
		GCPHealthCheck:                  "health-check-id",
		GCPForwardingRule:               "forwarding-rule-id",
	}
	assert.Error(client.SetState(stateMissingCloudProvider))
	stateIncorrectCloudProvider := state.ConstellationState{
		CloudProvider: "incorrect",
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
		GCPBackendService:               "backend-service-id",
		GCPHealthCheck:                  "health-check-id",
		GCPForwardingRule:               "forwarding-rule-id",
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
