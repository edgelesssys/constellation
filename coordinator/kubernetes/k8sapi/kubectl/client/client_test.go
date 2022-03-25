package client

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	apps "k8s.io/api/apps/v1"
	k8s "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/meta/testrestmapper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	restfake "k8s.io/client-go/rest/fake"
	"k8s.io/client-go/restmapper"
)

var (
	corev1GV        = schema.GroupVersion{Version: "v1"}
	nginxDeployment = &apps.Deployment{
		TypeMeta: v1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				"app": "nginx",
			},
			Name: "my-nginx",
		},
		Spec: apps.DeploymentSpec{
			Replicas: proto.Int32(3),
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: k8s.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: k8s.PodSpec{
					Containers: []k8s.Container{
						{
							Name:  "nginx",
							Image: "nginx:1.14.2",
							Ports: []k8s.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	nginxDeplJSON, _ = marshalJSON(nginxDeployment)
	nginxDeplYAML, _ = marshalYAML(nginxDeployment)
)

type unmarshableResource struct{}

func (*unmarshableResource) Marshal() ([]byte, error) {
	return nil, errors.New("someErr")
}

func stringBody(body string) io.ReadCloser {
	return io.NopCloser(bytes.NewReader([]byte(body)))
}

func fakeClientWith(t *testing.T, testName string, data map[string]string) resource.FakeClientFunc {
	return func(version schema.GroupVersion) (resource.RESTClient, error) {
		return &restfake.RESTClient{
			GroupVersion:         corev1GV,
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
			Client: restfake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
				p := req.URL.Path
				q := req.URL.RawQuery
				if len(q) != 0 {
					p = p + "?" + q
				}
				body, ok := data[p]
				if !ok {
					t.Fatalf("%s: unexpected request: %s (%s)\n%#v", testName, p, req.URL, req)
				}
				header := http.Header{}
				header.Set("Content-Type", runtime.ContentTypeJSON)
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       stringBody(body),
				}, nil
			}),
		}, nil
	}
}

func newClientWithFakes(t *testing.T, data map[string]string, objects ...runtime.Object) Client {
	clientset := fake.NewSimpleClientset(objects...)
	builder := resource.NewFakeBuilder(
		fakeClientWith(t, "", data),
		func() (meta.RESTMapper, error) {
			return testrestmapper.TestOnlyStaticRESTMapper(scheme.Scheme), nil
		},
		func() (restmapper.CategoryExpander, error) {
			return resource.FakeCategoryExpander, nil
		}).
		Unstructured()
	client := Client{
		clientset: clientset,
		builder:   builder,
	}
	return client
}

func failingClient() resource.FakeClientFunc {
	return func(version schema.GroupVersion) (resource.RESTClient, error) {
		return &restfake.RESTClient{
			GroupVersion:         corev1GV,
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
			Resp:                 &http.Response{StatusCode: 501},
		}, nil
	}
}

func newFailingClient(objects ...runtime.Object) Client {
	clientset := fake.NewSimpleClientset(objects...)
	builder := resource.NewFakeBuilder(
		failingClient(),
		func() (meta.RESTMapper, error) {
			return testrestmapper.TestOnlyStaticRESTMapper(scheme.Scheme), nil
		},
		func() (restmapper.CategoryExpander, error) {
			return resource.FakeCategoryExpander, nil
		}).
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...)
	client := Client{
		clientset: clientset,
		builder:   builder,
	}
	return client
}

func marshalJSON(obj runtime.Object) ([]byte, error) {
	serializer := json.NewSerializer(json.DefaultMetaFactory, nil, nil, false)
	var buf bytes.Buffer
	if err := serializer.Encode(obj, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func marshalYAML(obj runtime.Object) ([]byte, error) {
	serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
	var buf bytes.Buffer
	if err := serializer.Encode(obj, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestApplyOneObject(t *testing.T) {
	testCases := map[string]struct {
		httpResponseData map[string]string
		expectedObj      runtime.Object
		resourcesYAML    string
		failingClient    bool
		expectErr        bool
	}{
		"apply works": {
			httpResponseData: map[string]string{
				"/deployments/my-nginx?fieldManager=constellation-coordinator&force=true": string(nginxDeplJSON),
			},
			expectedObj:   nginxDeployment,
			resourcesYAML: string(nginxDeplYAML),
			expectErr:     false,
		},
		"apply fails": {
			httpResponseData: map[string]string{},
			expectedObj:      nginxDeployment,
			resourcesYAML:    string(nginxDeplYAML),
			failingClient:    true,
			expectErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			var client Client
			if tc.failingClient {
				client = newFailingClient(tc.expectedObj)
			} else {
				client = newClientWithFakes(t, tc.httpResponseData, tc.expectedObj)
			}

			reader := bytes.NewReader([]byte(tc.resourcesYAML))
			res := client.builder.
				ContinueOnError().
				Stream(reader, "yaml").
				Flatten().
				Do()
			assert.NoError(res.Err())
			infos, err := res.Infos()
			assert.NoError(err)
			require.Len(infos, 1)

			err = client.ApplyOneObject(infos[0], true)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestGetObjects(t *testing.T) {
	testCases := map[string]struct {
		expectedResources resources.Marshaler
		httpResponseData  map[string]string
		resourcesYAML     string
		expectErr         bool
	}{
		"GetObjects works on flannel deployment": {
			expectedResources: resources.NewDefaultFlannelDeployment(),
			resourcesYAML:     string(nginxDeplYAML),
			expectErr:         false,
		},
		"GetObjects works on cluster-autoscaler deployment": {
			expectedResources: resources.NewDefaultFlannelDeployment(),
			resourcesYAML:     string(nginxDeplYAML),
			expectErr:         false,
		},
		"GetObjects works on cloud-controller-manager deployment": {
			expectedResources: resources.NewDefaultCloudControllerManagerDeployment("someProvider", "someImage", "somePath", nil, nil, nil, nil),
			resourcesYAML:     string(nginxDeplYAML),
			expectErr:         false,
		},
		"GetObjects Marshal failure detected": {
			expectedResources: &unmarshableResource{},
			resourcesYAML:     string(nginxDeplYAML),
			expectErr:         true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := newClientWithFakes(t, tc.httpResponseData)
			infos, err := client.GetObjects(tc.expectedResources)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.NotNil(infos)
		})
	}
}
