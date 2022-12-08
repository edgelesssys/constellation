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

func (a stubProjectAPI) Get(ctx context.Context, req *computepb.GetProjectRequest,
	opts ...gax.CallOption,
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

func (a stubInstanceAPI) Get(ctx context.Context, req *computepb.GetInstanceRequest,
	opts ...gax.CallOption,
) (*computepb.Instance, error) {
	return a.instance, a.getErr
}

type stubInstanceTemplateAPI struct {
	template  *computepb.InstanceTemplate
	getErr    error
	deleteErr error
	insertErr error
}

func (a stubInstanceTemplateAPI) Close() error {
	return nil
}

func (a stubInstanceTemplateAPI) Get(ctx context.Context, req *computepb.GetInstanceTemplateRequest,
	opts ...gax.CallOption,
) (*computepb.InstanceTemplate, error) {
	return a.template, a.getErr
}

func (a stubInstanceTemplateAPI) Delete(ctx context.Context, req *computepb.DeleteInstanceTemplateRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	return &stubOperation{
		&computepb.Operation{
			Name: proto.String("name"),
		},
	}, a.deleteErr
}

func (a stubInstanceTemplateAPI) Insert(ctx context.Context, req *computepb.InsertInstanceTemplateRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	return &stubOperation{
		&computepb.Operation{
			Name: proto.String("name"),
		},
	}, a.insertErr
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

func (a stubInstanceGroupManagersAPI) Get(ctx context.Context, req *computepb.GetInstanceGroupManagerRequest,
	opts ...gax.CallOption,
) (*computepb.InstanceGroupManager, error) {
	return a.instanceGroupManager, a.getErr
}

func (a stubInstanceGroupManagersAPI) AggregatedList(ctx context.Context, req *computepb.AggregatedListInstanceGroupManagersRequest,
	opts ...gax.CallOption,
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

func (a stubInstanceGroupManagersAPI) SetInstanceTemplate(ctx context.Context, req *computepb.SetInstanceTemplateInstanceGroupManagerRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	return &stubOperation{
		&computepb.Operation{
			Name: proto.String("name"),
		},
	}, a.setInstanceTemplateErr
}

func (a stubInstanceGroupManagersAPI) CreateInstances(ctx context.Context, req *computepb.CreateInstancesInstanceGroupManagerRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	return &stubOperation{
		&computepb.Operation{
			Name: proto.String("name"),
		},
	}, a.createInstancesErr
}

func (a stubInstanceGroupManagersAPI) DeleteInstances(ctx context.Context, req *computepb.DeleteInstancesInstanceGroupManagerRequest,
	opts ...gax.CallOption,
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

func (a stubDiskAPI) Get(ctx context.Context, req *computepb.GetDiskRequest,
	opts ...gax.CallOption,
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

func (o *stubOperation) Wait(ctx context.Context, opts ...gax.CallOption) error {
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
