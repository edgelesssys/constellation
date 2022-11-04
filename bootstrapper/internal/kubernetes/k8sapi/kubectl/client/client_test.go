/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/protobuf/proto"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8s "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/meta/testrestmapper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	restfake "k8s.io/client-go/rest/fake"
	"k8s.io/client-go/restmapper"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

var (
	corev1GV        = schema.GroupVersion{Version: "v1"}
	nginxDeployment = &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app": "nginx",
			},
			Name: "my-nginx",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: proto.Int32(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: k8s.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
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
	tolerationsDeployment = appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
			Name:      "test-deployment",
		},
	}
	selectorsDeployment = appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
			Name:      "test-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Template: k8s.PodTemplateSpec{
				Spec: k8s.PodSpec{
					NodeSelector: map[string]string{},
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
		wantObj          runtime.Object
		resourcesYAML    string
		failingClient    bool
		wantErr          bool
	}{
		"apply works": {
			httpResponseData: map[string]string{
				"/deployments/my-nginx?fieldManager=constellation-bootstrapper&force=true": string(nginxDeplJSON),
			},
			wantObj:       nginxDeployment,
			resourcesYAML: string(nginxDeplYAML),
			wantErr:       false,
		},
		"apply fails": {
			httpResponseData: map[string]string{},
			wantObj:          nginxDeployment,
			resourcesYAML:    string(nginxDeplYAML),
			failingClient:    true,
			wantErr:          true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			var client Client
			if tc.failingClient {
				client = newFailingClient(tc.wantObj)
			} else {
				client = newClientWithFakes(t, tc.httpResponseData, tc.wantObj)
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

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestGetObjects(t *testing.T) {
	testCases := map[string]struct {
		wantResources    kubernetes.Marshaler
		httpResponseData map[string]string
		resourcesYAML    string
		wantErr          bool
	}{
		"GetObjects Marshal failure detected": {
			wantResources: &unmarshableResource{},
			resourcesYAML: string(nginxDeplYAML),
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := newClientWithFakes(t, tc.httpResponseData)
			infos, err := client.GetObjects(tc.wantResources)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.NotNil(infos)
		})
	}
}

func TestAddTolerationsToDeployment(t *testing.T) {
	testCases := map[string]struct {
		namespace   string
		name        string
		tolerations []corev1.Toleration
		wantErr     bool
	}{
		"Success": {
			namespace: "test-ns",
			name:      "test-deployment",
		},
		"Specifying non-existent deployment fails": {
			namespace: "test-ns",
			name:      "wrong-name",
			wantErr:   true,
		},
		"Wrong namespace": {
			name:    "test-deployment",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := newClientWithFakes(t, map[string]string{}, &tolerationsDeployment)
			err := client.AddTolerationsToDeployment(context.Background(), tc.tolerations, tc.name, tc.namespace)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestAddNodeSelectorsToDeployment(t *testing.T) {
	testCases := map[string]struct {
		namespace string
		name      string
		selectors map[string]string
		wantErr   bool
	}{
		"Success": {
			namespace: "test-ns",
			name:      "test-deployment",
			selectors: map[string]string{"some-key": "some-value"},
		},
		"Specifying non-existent deployment fails": {
			namespace: "test-ns",
			name:      "wrong-name",
			wantErr:   true,
		},
		"Wrong namespace": {
			name:    "test-deployment",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := newClientWithFakes(t, map[string]string{}, &selectorsDeployment)
			err := client.AddNodeSelectorsToDeployment(context.Background(), tc.selectors, tc.name, tc.namespace)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestWaitForCRD(t *testing.T) {
	testCases := map[string]struct {
		crd      string
		events   []watch.Event
		watchErr error
		wantErr  bool
	}{
		"Success": {
			crd: "test-crd",
			events: []watch.Event{
				{
					Type: watch.Added,
					Object: &apiextensionsv1.CustomResourceDefinition{
						Status: apiextensionsv1.CustomResourceDefinitionStatus{
							Conditions: []apiextensionsv1.CustomResourceDefinitionCondition{
								{
									Type:   apiextensionsv1.Established,
									Status: apiextensionsv1.ConditionTrue,
								},
							},
						},
					},
				},
			},
		},
		"watch error": {
			crd:      "test-crd",
			watchErr: errors.New("watch error"),
			wantErr:  true,
		},
		"crd deleted": {
			crd:     "test-crd",
			events:  []watch.Event{{Type: watch.Deleted}},
			wantErr: true,
		},
		"other error": {
			crd:     "test-crd",
			events:  []watch.Event{{Type: watch.Error}},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				apiextensionClient: &stubCRDWatcher{events: tc.events, watchErr: tc.watchErr},
			}
			err := client.WaitForCRD(context.Background(), tc.crd)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

type stubCRDWatcher struct {
	events   []watch.Event
	watchErr error

	apiextensionsclientv1.ApiextensionsV1Interface
}

func (w *stubCRDWatcher) CustomResourceDefinitions() apiextensionsclientv1.CustomResourceDefinitionInterface {
	return &stubCustomResourceDefinitions{
		events:   w.events,
		watchErr: w.watchErr,
	}
}

type stubCustomResourceDefinitions struct {
	events   []watch.Event
	watchErr error

	apiextensionsclientv1.CustomResourceDefinitionInterface
}

func (c *stubCustomResourceDefinitions) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	eventChan := make(chan watch.Event, len(c.events))
	for _, event := range c.events {
		eventChan <- event
	}
	return &stubCRDWatch{events: eventChan}, c.watchErr
}

type stubCRDWatch struct {
	events chan watch.Event
}

func (w *stubCRDWatch) Stop() {
	close(w.events)
}

func (w *stubCRDWatch) ResultChan() <-chan watch.Event {
	return w.events
}
