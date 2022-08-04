package kubectl

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi/resources"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/resource"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type stubClient struct {
	applyOneObjectErr              error
	getObjectsInfos                []*resource.Info
	getObjectsErr                  error
	createConfigMapErr             error
	addTolerationsToDeploymentErr  error
	addNodeSelectorToDeploymentErr error
	waitForCRDErr                  error
}

func (s *stubClient) ApplyOneObject(info *resource.Info, forceConflicts bool) error {
	return s.applyOneObjectErr
}

func (s *stubClient) GetObjects(resources resources.Marshaler) ([]*resource.Info, error) {
	return s.getObjectsInfos, s.getObjectsErr
}

func (s *stubClient) CreateConfigMap(ctx context.Context, configMap corev1.ConfigMap) error {
	return s.createConfigMapErr
}

func (s *stubClient) AddTolerationsToDeployment(ctx context.Context, tolerations []corev1.Toleration, name string, namespace string) error {
	return s.addTolerationsToDeploymentErr
}

func (s *stubClient) AddNodeSelectorsToDeployment(ctx context.Context, selectors map[string]string, name string, namespace string) error {
	return s.addNodeSelectorToDeploymentErr
}

type stubClientGenerator struct {
	applyOneObjectErr              error
	getObjectsInfos                []*resource.Info
	getObjectsErr                  error
	newClientErr                   error
	createConfigMapErr             error
	addTolerationsToDeploymentErr  error
	addNodeSelectorToDeploymentErr error
	waitForCRDErr                  error
}

func (s *stubClient) WaitForCRD(ctx context.Context, crd string) error {
	return s.waitForCRDErr
}

func (s *stubClientGenerator) NewClient(kubeconfig []byte) (Client, error) {
	return &stubClient{
		applyOneObjectErr:              s.applyOneObjectErr,
		getObjectsInfos:                s.getObjectsInfos,
		getObjectsErr:                  s.getObjectsErr,
		createConfigMapErr:             s.createConfigMapErr,
		addTolerationsToDeploymentErr:  s.addTolerationsToDeploymentErr,
		addNodeSelectorToDeploymentErr: s.addNodeSelectorToDeploymentErr,
		waitForCRDErr:                  s.waitForCRDErr,
	}, s.newClientErr
}

type dummyResource struct{}

func (*dummyResource) Marshal() ([]byte, error) {
	panic("dummy")
}

func TestApplyWorks(t *testing.T) {
	assert := assert.New(t)
	kube := Kubectl{
		clientGenerator: &stubClientGenerator{
			getObjectsInfos: []*resource.Info{
				{},
			},
		},
	}
	kube.SetKubeconfig([]byte("someConfig"))

	assert.NoError(kube.Apply(&dummyResource{}, true))
}

func TestKubeconfigUnset(t *testing.T) {
	assert := assert.New(t)
	kube := Kubectl{}

	assert.ErrorIs(kube.Apply(&dummyResource{}, true), ErrKubeconfigNotSet)
}

func TestClientGeneratorFails(t *testing.T) {
	assert := assert.New(t)
	err := errors.New("generator failed")
	kube := Kubectl{
		clientGenerator: &stubClientGenerator{
			newClientErr: err,
		},
	}
	kube.SetKubeconfig([]byte("someConfig"))

	assert.ErrorIs(kube.Apply(&dummyResource{}, true), err)
}

func TestGetObjectsFails(t *testing.T) {
	assert := assert.New(t)
	err := errors.New("getObjects failed")
	kube := Kubectl{
		clientGenerator: &stubClientGenerator{
			getObjectsErr: err,
		},
	}
	kube.SetKubeconfig([]byte("someConfig"))

	assert.ErrorIs(kube.Apply(&dummyResource{}, true), err)
}

func TestApplyOneObjectFails(t *testing.T) {
	assert := assert.New(t)
	err := errors.New("applyOneObject failed")
	kube := Kubectl{
		clientGenerator: &stubClientGenerator{
			getObjectsInfos: []*resource.Info{
				{},
			},
			applyOneObjectErr: err,
		},
	}
	kube.SetKubeconfig([]byte("someConfig"))

	assert.ErrorIs(kube.Apply(&dummyResource{}, true), err)
}
