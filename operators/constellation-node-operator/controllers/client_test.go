/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package controllers

import (
	"context"
	"reflect"
	"testing"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type stubReaderClient struct {
	objects map[schema.GroupVersionKind]map[client.ObjectKey]runtime.Object
	getErr  error
	listErr error
	scheme  *runtime.Scheme
	client.Client
}

func newStubReaderClient(t *testing.T, objects []runtime.Object, getErr, listErr error) *stubReaderClient {
	scheme := getScheme(t)
	objectsMap := make(map[schema.GroupVersionKind]map[client.ObjectKey]runtime.Object)
	for _, obj := range objects {
		if obj == nil {
			continue
		}
		gvks, _, err := scheme.ObjectKinds(obj)
		if err != nil {
			panic(err)
		}
		gvk := gvks[0]
		if _, ok := objectsMap[gvk]; !ok {
			objectsMap[gvk] = make(map[client.ObjectKey]runtime.Object)
		}
		objectsMap[gvk][client.ObjectKey{Namespace: obj.(client.Object).GetNamespace(), Name: obj.(client.Object).GetName()}] = obj
	}
	return &stubReaderClient{
		objects: objectsMap,
		getErr:  getErr,
		listErr: listErr,
		scheme:  scheme,
	}
}

func (c *stubReaderClient) Get(_ context.Context, key client.ObjectKey, out client.Object, _ ...client.GetOption) error {
	gvks, _, err := c.scheme.ObjectKinds(out)
	if err != nil {
		panic(err)
	}
	gvk := gvks[0]
	result := c.objects[gvk][key]
	if result == nil {
		return c.getErr
	}
	obj := result.DeepCopyObject()
	outVal := reflect.ValueOf(out)
	objVal := reflect.ValueOf(obj)
	if !objVal.Type().AssignableTo(outVal.Type()) {
		panic("type mismatch")
	}
	reflect.Indirect(outVal).Set(reflect.Indirect(objVal))
	out.GetObjectKind().SetGroupVersionKind(gvk)
	return c.getErr
}

func (c *stubReaderClient) List(_ context.Context, out client.ObjectList, _ ...client.ListOption) error {
	gvks, _, err := c.scheme.ObjectKinds(out)
	if err != nil {
		panic(err)
	}
	listGVK := gvks[0]
	itemGVK := listGVK.GroupVersion().WithKind(listGVK.Kind[:len(listGVK.Kind)-len("List")])
	results := c.objects[itemGVK]
	runtimeObjs := make([]runtime.Object, 0, len(results))
	for _, item := range results {
		outObj := item.DeepCopyObject()
		outObj.GetObjectKind().SetGroupVersionKind(itemGVK)
		runtimeObjs = append(runtimeObjs, outObj)
	}
	if err := apimeta.SetList(out, runtimeObjs); err != nil {
		panic(err)
	}
	out.GetObjectKind().SetGroupVersionKind(listGVK)
	return c.listErr
}

type stubWriterClient struct {
	createErr      error
	deleteErr      error
	updateErr      error
	patchErr       error
	deleteAllOfErr error
	statusWriter   stubStatusWriter
	client.Client
}

func (c *stubWriterClient) Create(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
	return c.createErr
}

func (c *stubWriterClient) Delete(_ context.Context, _ client.Object, _ ...client.DeleteOption) error {
	return c.deleteErr
}

func (c *stubWriterClient) Update(_ context.Context, _ client.Object, _ ...client.UpdateOption) error {
	return c.updateErr
}

func (c *stubWriterClient) Patch(_ context.Context, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
	return c.patchErr
}

func (c *stubWriterClient) DeleteAllOf(_ context.Context, _ client.Object, _ ...client.DeleteAllOfOption) error {
	return c.deleteAllOfErr
}

func (c *stubWriterClient) Status() client.StatusWriter {
	return &c.statusWriter
}

type stubReadWriterClient struct {
	stubReaderClient
	stubWriterClient
	client.Client
}

func (c *stubReadWriterClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return c.stubReaderClient.Get(ctx, key, obj, opts...)
}

func (c *stubReadWriterClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return c.stubReaderClient.List(ctx, list, opts...)
}

func (c *stubReadWriterClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return c.stubWriterClient.Create(ctx, obj, opts...)
}

func (c *stubReadWriterClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return c.stubWriterClient.Delete(ctx, obj, opts...)
}

func (c *stubReadWriterClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return c.stubWriterClient.Update(ctx, obj, opts...)
}

func (c *stubReadWriterClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return c.stubWriterClient.Patch(ctx, obj, patch, opts...)
}

func (c *stubReadWriterClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return c.stubWriterClient.DeleteAllOf(ctx, obj, opts...)
}

func (c *stubReadWriterClient) Status() client.StatusWriter {
	return c.stubWriterClient.Status()
}

type stubStatusWriter struct {
	createErr error
	updateErr error
	patchErr  error
}

func (w *stubStatusWriter) Create(_ context.Context, _ client.Object, _ client.Object, _ ...client.SubResourceCreateOption) error {
	return w.createErr
}

func (w *stubStatusWriter) Update(_ context.Context, _ client.Object, _ ...client.SubResourceUpdateOption) error {
	return w.updateErr
}

func (w *stubStatusWriter) Patch(_ context.Context, _ client.Object, _ client.Patch, _ ...client.SubResourcePatchOption) error {
	return w.patchErr
}
