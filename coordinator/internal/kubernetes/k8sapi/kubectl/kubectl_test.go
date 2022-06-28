package kubectl

import (
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/internal/kubernetes/k8sapi/resources"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"k8s.io/cli-runtime/pkg/resource"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type stubClient struct {
	applyOneObjectErr error
	getObjectsInfos   []*resource.Info
	getObjectsErr     error
}

func (s *stubClient) ApplyOneObject(info *resource.Info, forceConflicts bool) error {
	return s.applyOneObjectErr
}

func (s *stubClient) GetObjects(resources resources.Marshaler) ([]*resource.Info, error) {
	return s.getObjectsInfos, s.getObjectsErr
}

type stubClientGenerator struct {
	applyOneObjectErr error
	getObjectsInfos   []*resource.Info
	getObjectsErr     error
	newClientErr      error
}

func (s *stubClientGenerator) NewClient(kubeconfig []byte) (Client, error) {
	return &stubClient{
		s.applyOneObjectErr,
		s.getObjectsInfos,
		s.getObjectsErr,
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
