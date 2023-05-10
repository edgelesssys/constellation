/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"flag"
	"os"
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"k8s.io/client-go/discovery"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	awsclient "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/internal/cloud/aws/client"
	azureclient "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/internal/cloud/azure/client"
	cloudfake "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/internal/cloud/fake/client"
	gcpclient "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/internal/cloud/gcp/client"
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/internal/deploy"
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/internal/upgrade"

	nodemaintenancev1beta1 "github.com/edgelesssys/constellation/v2/3rdparty/node-maintenance-operator/api/v1beta1"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/controllers"
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/internal/etcd"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	defaultAzureCloudConfigPath = "/etc/azure/azure.json"
	defaultGCPCloudConfigPath   = "/etc/gce/gce.conf"
	// constellationCSP is the environment variable stating which Cloud Service Provider Constellation is running on.
	constellationCSP = "CONSTEL_CSP"
	// constellationUID is the environment variable stating which uid is used to tag / label cloud provider resources belonging to one constellation.
	constellationUID = "constellation-uid"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(nodemaintenancev1beta1.AddToScheme(scheme))
	utilruntime.Must(updatev1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var cloudConfigPath string
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&cloudConfigPath, "cloud-config", "", "Path to provider specific cloud config. Optional.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Create CSP client
	var cspClient cspAPI
	var clientErr error
	csp := strings.ToLower(os.Getenv(constellationCSP))
	switch csp {
	case "azure":
		if cloudConfigPath == "" {
			cloudConfigPath = defaultAzureCloudConfigPath
		}
		cspClient, clientErr = azureclient.NewFromDefault(cloudConfigPath)
		if clientErr != nil {
			setupLog.Error(clientErr, "Unable to create Azure client")
			os.Exit(1)
		}
	case "gcp":
		if cloudConfigPath == "" {
			cloudConfigPath = defaultGCPCloudConfigPath
		}
		cspClient, clientErr = gcpclient.New(context.Background(), cloudConfigPath)
		if clientErr != nil {
			setupLog.Error(clientErr, "unable to create GCP client")
			os.Exit(1)
		}
	case "aws":
		cspClient, clientErr = awsclient.New(context.Background())
		if clientErr != nil {
			setupLog.Error(clientErr, "unable to create AWS client")
			os.Exit(1)
		}
	default:
		setupLog.Info("CSP does not support upgrades", "csp", csp)
		cspClient = &cloudfake.Client{}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "38cc1645.edgeless.systems",
	})
	if err != nil {
		setupLog.Error(err, "Unable to start manager")
		os.Exit(1)
	}

	k8sClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "Unable to create k8s client")
		os.Exit(1)
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		setupLog.Error(err, "Unable to create discovery client")
		os.Exit(1)
	}
	etcdClient, err := etcd.New(k8sClient)
	if err != nil {
		setupLog.Error(err, "Unable to create etcd client")
		os.Exit(1)
	}
	defer etcdClient.Close()

	imageInfo := deploy.NewImageInfo()
	if err := deploy.InitialResources(context.Background(), k8sClient, imageInfo, cspClient, os.Getenv(constellationUID)); err != nil {
		setupLog.Error(err, "Unable to deploy initial resources")
		os.Exit(1)
	}
	// Create Controllers
	if csp == "azure" || csp == "gcp" || csp == "aws" {
		if err = controllers.NewNodeVersionReconciler(
			cspClient, etcdClient, upgrade.NewClient(), discoveryClient, mgr.GetClient(), mgr.GetScheme(),
		).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "Unable to create controller", "controller", "NodeVersion")
			os.Exit(1)
		}
		if err = (&controllers.AutoscalingStrategyReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "Unable to create controller", "controller", "AutoscalingStrategy")
			os.Exit(1)
		}
		if err = controllers.NewScalingGroupReconciler(
			cspClient, mgr.GetClient(), mgr.GetScheme(),
		).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "Unable to create controller", "controller", "ScalingGroup")
			os.Exit(1)
		}
		if err = controllers.NewPendingNodeReconciler(
			cspClient, mgr.GetClient(), mgr.GetScheme(),
		).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "Unable to create controller", "controller", "PendingNode")
			os.Exit(1)
		}
	}

	if err = controllers.NewJoiningNodesReconciler(
		mgr.GetClient(),
		mgr.GetScheme(),
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "JoiningNode")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Problem running manager")
		os.Exit(1)
	}
}

type cspAPI interface {
	// GetNodeImage retrieves the image currently used by a node.
	GetNodeImage(ctx context.Context, providerID string) (string, error)
	// GetScalingGroupID retrieves the scaling group that a node is part of.
	GetScalingGroupID(ctx context.Context, providerID string) (string, error)
	// CreateNode creates a new node inside a specified scaling group at the CSP and returns its future name and provider id.
	CreateNode(ctx context.Context, scalingGroupID string) (nodeName, providerID string, err error)
	// DeleteNode starts the termination of the node at the CSP.
	DeleteNode(ctx context.Context, providerID string) error
	// GetNodeState retrieves the state of a pending node from a CSP.
	GetNodeState(ctx context.Context, providerID string) (updatev1alpha1.CSPNodeState, error)
	// GetScalingGroupImage retrieves the image currently used by a scaling group.
	GetScalingGroupImage(ctx context.Context, scalingGroupID string) (string, error)
	// SetScalingGroupImage sets the image to be used by newly created nodes in a scaling group.
	SetScalingGroupImage(ctx context.Context, scalingGroupID, imageURI string) error
	// GetScalingGroupName retrieves the name of a scaling group.
	GetScalingGroupName(scalingGroupID string) (string, error)
	// GetAutoscalingGroupName retrieves the name of a scaling group as needed by the cluster-autoscaler.
	GetAutoscalingGroupName(scalingGroupID string) (string, error)
	// ListScalingGroups retrieves a list of scaling groups for the cluster.
	ListScalingGroups(ctx context.Context, uid string) (controlPlaneGroupIDs []string, workerGroupIDs []string, err error)
	// AutoscalingCloudProvider returns the cloud-provider name as used by k8s cluster-autoscaler.
	AutoscalingCloudProvider() string
}
