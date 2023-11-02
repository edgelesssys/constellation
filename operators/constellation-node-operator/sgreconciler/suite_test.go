//go:build integration

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package sgreconciler

import (
	"context"
	"path/filepath"
	"testing"

	//revive:disable:dot-imports
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	//revive:enable:dot-imports

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/internal/executor"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg                    *rest.Config
	k8sClient              client.Client
	testEnv                *envtest.Environment
	ctx                    context.Context
	cancel                 context.CancelFunc
	fakes                  = newFakes()
	triggerReconcile       func()
	stopAndWaitForExecutor func()
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, reporterConfig := GinkgoConfiguration()
	// If you want to debug a specific seed, set it here.
	// suiteConfig.RandomSeed = 1679587116
	reporterConfig.VeryVerbose = true
	RunSpecs(t, "Controller Suite", suiteConfig, reporterConfig)
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	ctx, cancel = context.WithCancel(context.Background())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "config", "crd", "bases"),
		},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = updatev1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	extScalingGroupReconciler := NewExternalScalingGroupReconciler(
		"uid",
		fakes.scalingGroupDiscoverer,
		k8sClient,
	)

	exec := executor.New(extScalingGroupReconciler, executor.Config{})
	triggerReconcile = exec.Trigger

	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		stopAndWaitForExecutor = exec.Start(ctx)
	}()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	defer stopAndWaitForExecutor()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

type fakeCollection struct {
	scalingGroupDiscoverer *stubScalingGroupDiscoverer
}

func (c *fakeCollection) reset() {
	c.scalingGroupDiscoverer.reset()
}

func newFakes() *fakeCollection {
	return &fakeCollection{
		scalingGroupDiscoverer: &stubScalingGroupDiscoverer{},
	}
}

func resetEnv() error {
	// cleanup all nodes
	nodeList := &corev1.NodeList{}
	if err := k8sClient.List(context.Background(), nodeList); err != nil {
		return err
	}
	for _, node := range nodeList.Items {
		if err := k8sClient.Delete(context.Background(), &node); err != nil {
			return err
		}
	}
	// cleanup all node versions
	nodeVersionList := &updatev1alpha1.NodeVersionList{}
	if err := k8sClient.List(context.Background(), nodeVersionList); err != nil {
		return err
	}
	for _, nodeVersion := range nodeVersionList.Items {
		if err := k8sClient.Delete(context.Background(), &nodeVersion); err != nil {
			return err
		}
	}
	// cleanup all scaling groups
	scalingGroupList := &updatev1alpha1.ScalingGroupList{}
	if err := k8sClient.List(context.Background(), scalingGroupList); err != nil {
		return err
	}
	for _, scalingGroup := range scalingGroupList.Items {
		if err := k8sClient.Delete(context.Background(), &scalingGroup); err != nil {
			return err
		}
	}
	// cleanup all pending nodes
	pendingNodeList := &updatev1alpha1.PendingNodeList{}
	if err := k8sClient.List(context.Background(), pendingNodeList); err != nil {
		return err
	}
	for _, pendingNode := range pendingNodeList.Items {
		if err := k8sClient.Delete(context.Background(), &pendingNode); err != nil {
			return err
		}
	}
	// cleanup all joining nodes
	joiningNodeList := &updatev1alpha1.JoiningNodeList{}
	if err := k8sClient.List(context.Background(), joiningNodeList); err != nil {
		return err
	}
	for _, joiningNode := range joiningNodeList.Items {
		if err := k8sClient.Delete(context.Background(), &joiningNode); err != nil {
			return err
		}
	}
	// cleanup all autoscaling strategies
	autoscalingStrategyList := &updatev1alpha1.AutoscalingStrategyList{}
	if err := k8sClient.List(context.Background(), autoscalingStrategyList); err != nil {
		return err
	}
	for _, autoscalingStrategy := range autoscalingStrategyList.Items {
		if err := k8sClient.Delete(context.Background(), &autoscalingStrategy); err != nil {
			return err
		}
	}
	// cleanup all deployments
	deploymentList := &appsv1.DeploymentList{}
	if err := k8sClient.List(context.Background(), deploymentList); err != nil {
		return err
	}
	for _, deployment := range deploymentList.Items {
		if err := k8sClient.Delete(context.Background(), &deployment); err != nil {
			return err
		}
	}
	fakes.reset()
	return nil
}
