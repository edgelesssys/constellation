/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	computeREST "google.golang.org/api/compute/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"
)

type stubProjectAPI struct {
	project *computepb.Project
	getErr  error
}

func (a stubProjectAPI) Close() error {
	return nil
}

func (a stubProjectAPI) Get(_ context.Context, _ *computepb.GetProjectRequest,
	_ ...gax.CallOption,
) (*computepb.Project, error) {
	return a.project, a.getErr
}

type stubInstanceAPI struct {
	instance *computepb.Instance
	getErr   error
}

func (a stubInstanceAPI) Close() error {
	return nil
}

func (a stubInstanceAPI) Get(_ context.Context, _ *computepb.GetInstanceRequest,
	_ ...gax.CallOption,
) (*computepb.Instance, error) {
	return a.instance, a.getErr
}

type stubInstanceTemplateAPI struct {
	template  *computeREST.InstanceTemplate
	getErr    error
	deleteErr error
	insertErr error
}

func (a stubInstanceTemplateAPI) Close() error {
	return nil
}

func (a stubInstanceTemplateAPI) Get(_, _ string) (*computeREST.InstanceTemplate, error) {
	return a.template, a.getErr
}

func (a stubInstanceTemplateAPI) Delete(_, _ string) (*computeREST.Operation, error) {
	return &computeREST.Operation{}, a.deleteErr
}

func (a stubInstanceTemplateAPI) Insert(_ string, _ *computeREST.InstanceTemplate) (*computeREST.Operation, error) {
	return &computeREST.Operation{}, a.insertErr
}

type stubInstanceGroupManagersAPI struct {
	instanceGroupManager   *computepb.InstanceGroupManager
	getErr                 error
	aggregatedListErr      error
	setInstanceTemplateErr error
	createInstancesErr     error
	deleteInstancesErr     error
}

func (a stubInstanceGroupManagersAPI) Close() error {
	return nil
}

func (a stubInstanceGroupManagersAPI) Get(_ context.Context, _ *computepb.GetInstanceGroupManagerRequest,
	_ ...gax.CallOption,
) (*computepb.InstanceGroupManager, error) {
	return a.instanceGroupManager, a.getErr
}

func (a stubInstanceGroupManagersAPI) AggregatedList(_ context.Context, _ *computepb.AggregatedListInstanceGroupManagersRequest,
	_ ...gax.CallOption,
) InstanceGroupManagerScopedListIterator {
	return &stubInstanceGroupManagerScopedListIterator{
		pairs: []compute.InstanceGroupManagersScopedListPair{
			{
				Key: "key",
				Value: &computepb.InstanceGroupManagersScopedList{
					InstanceGroupManagers: []*computepb.InstanceGroupManager{
						a.instanceGroupManager,
					},
				},
			},
		},
		nextErr: a.aggregatedListErr,
	}
}

func (a stubInstanceGroupManagersAPI) SetInstanceTemplate(_ context.Context, _ *computepb.SetInstanceTemplateInstanceGroupManagerRequest,
	_ ...gax.CallOption,
) (Operation, error) {
	return &stubOperation{
		&computepb.Operation{
			Name: proto.String("name"),
		},
	}, a.setInstanceTemplateErr
}

func (a stubInstanceGroupManagersAPI) CreateInstances(_ context.Context, _ *computepb.CreateInstancesInstanceGroupManagerRequest,
	_ ...gax.CallOption,
) (Operation, error) {
	return &stubOperation{
		&computepb.Operation{
			Name: proto.String("name"),
		},
	}, a.createInstancesErr
}

func (a stubInstanceGroupManagersAPI) DeleteInstances(_ context.Context, _ *computepb.DeleteInstancesInstanceGroupManagerRequest,
	_ ...gax.CallOption,
) (Operation, error) {
	if a.deleteInstancesErr != nil {
		return nil, a.deleteInstancesErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name: proto.String("name"),
		},
	}, nil
}

type stubDiskAPI struct {
	disk   *computepb.Disk
	getErr error
}

func (a stubDiskAPI) Close() error {
	return nil
}

func (a stubDiskAPI) Get(_ context.Context, _ *computepb.GetDiskRequest,
	_ ...gax.CallOption,
) (*computepb.Disk, error) {
	return a.disk, a.getErr
}

type stubOperation struct {
	*computepb.Operation
}

func (o *stubOperation) Proto() *computepb.Operation {
	return o.Operation
}

func (o *stubOperation) Done() bool {
	return true
}

func (o *stubOperation) Wait(_ context.Context, _ ...gax.CallOption) error {
	return nil
}

type stubInstanceGroupManagerScopedListIterator struct {
	pairs   []compute.InstanceGroupManagersScopedListPair
	nextErr error

	internalCounter int
}

func (i *stubInstanceGroupManagerScopedListIterator) Next() (compute.InstanceGroupManagersScopedListPair, error) {
	if i.nextErr != nil {
		return compute.InstanceGroupManagersScopedListPair{}, i.nextErr
	}
	if i.internalCounter >= len(i.pairs) {
		return compute.InstanceGroupManagersScopedListPair{}, iterator.Done
	}
	pair := i.pairs[i.internalCounter]
	i.internalCounter++
	return pair, nil
}
