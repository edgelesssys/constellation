/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"

	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
	"google.golang.org/protobuf/proto"
)

type stubOperation struct {
	*computepb.Operation
}

func (o *stubOperation) Proto() *computepb.Operation {
	return o.Operation
}

type stubInstanceAPI struct {
	listIterator *stubInstanceIterator
}

func (a stubInstanceAPI) Close() error {
	return nil
}

func (a stubInstanceAPI) List(ctx context.Context, req *computepb.ListInstancesRequest,
	opts ...gax.CallOption,
) InstanceIterator {
	return a.listIterator
}

type stubInstanceIterator struct {
	instances []*computepb.Instance
	nextErr   error

	internalCounter int
}

func (i *stubInstanceIterator) Next() (*computepb.Instance, error) {
	if i.nextErr != nil {
		return nil, i.nextErr
	}
	if i.internalCounter >= len(i.instances) {
		i.internalCounter = 0
		return nil, iterator.Done
	}
	resp := i.instances[i.internalCounter]
	i.internalCounter++
	return resp, nil
}

type stubOperationZoneAPI struct {
	waitErr error
}

func (a stubOperationZoneAPI) Close() error {
	return nil
}

func (a stubOperationZoneAPI) Wait(ctx context.Context, req *computepb.WaitZoneOperationRequest,
	opts ...gax.CallOption,
) (*computepb.Operation, error) {
	if a.waitErr != nil {
		return nil, a.waitErr
	}
	return &computepb.Operation{
		Status: computepb.Operation_DONE.Enum(),
	}, nil
}

type stubOperationRegionAPI struct {
	waitErr error
}

func (a stubOperationRegionAPI) Close() error {
	return nil
}

func (a stubOperationRegionAPI) Wait(ctx context.Context, req *computepb.WaitRegionOperationRequest,
	opts ...gax.CallOption,
) (*computepb.Operation, error) {
	if a.waitErr != nil {
		return nil, a.waitErr
	}
	return &computepb.Operation{
		Status: computepb.Operation_DONE.Enum(),
	}, nil
}

type stubOperationGlobalAPI struct {
	waitErr error
}

func (a stubOperationGlobalAPI) Close() error {
	return nil
}

func (a stubOperationGlobalAPI) Wait(ctx context.Context, req *computepb.WaitGlobalOperationRequest,
	opts ...gax.CallOption,
) (*computepb.Operation, error) {
	if a.waitErr != nil {
		return nil, a.waitErr
	}
	return &computepb.Operation{
		Status: computepb.Operation_DONE.Enum(),
	}, nil
}

type stubFirewallsAPI struct {
	deleteErr error
	insertErr error
}

func (a stubFirewallsAPI) Close() error {
	return nil
}

func (a stubFirewallsAPI) Delete(ctx context.Context, req *computepb.DeleteFirewallRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.deleteErr != nil {
		return nil, a.deleteErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name: proto.String("name"),
		},
	}, nil
}

func (a stubFirewallsAPI) Insert(ctx context.Context, req *computepb.InsertFirewallRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.insertErr != nil {
		return nil, a.insertErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name: proto.String("name"),
		},
	}, nil
}

type stubNetworksAPI struct {
	insertErr error
	deleteErr error
}

func (a stubNetworksAPI) Close() error {
	return nil
}

func (a stubNetworksAPI) Insert(ctx context.Context, req *computepb.InsertNetworkRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.insertErr != nil {
		return nil, a.insertErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name: proto.String("name"),
		},
	}, nil
}

func (a stubNetworksAPI) Delete(ctx context.Context, req *computepb.DeleteNetworkRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.deleteErr != nil {
		return nil, a.deleteErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name: proto.String("name"),
		},
	}, nil
}

type stubSubnetworksAPI struct {
	insertErr error
	deleteErr error
}

func (a stubSubnetworksAPI) Close() error {
	return nil
}

func (a stubSubnetworksAPI) Insert(ctx context.Context, req *computepb.InsertSubnetworkRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.insertErr != nil {
		return nil, a.insertErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name:   proto.String("name"),
			Region: proto.String("region"),
		},
	}, nil
}

func (a stubSubnetworksAPI) Delete(ctx context.Context, req *computepb.DeleteSubnetworkRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.deleteErr != nil {
		return nil, a.deleteErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name:   proto.String("name"),
			Region: proto.String("region"),
		},
	}, nil
}

type stubBackendServicesAPI struct {
	insertErr error
	deleteErr error
}

func (a stubBackendServicesAPI) Close() error {
	return nil
}

func (a stubBackendServicesAPI) Insert(ctx context.Context, req *computepb.InsertRegionBackendServiceRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.insertErr != nil {
		return nil, a.insertErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name:   proto.String("name"),
			Region: proto.String("region"),
		},
	}, nil
}

func (a stubBackendServicesAPI) Delete(ctx context.Context, req *computepb.DeleteRegionBackendServiceRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.deleteErr != nil {
		return nil, a.deleteErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name:   proto.String("name"),
			Region: proto.String("region"),
		},
	}, nil
}

type stubForwardingRulesAPI struct {
	insertErr      error
	deleteErr      error
	getErr         error
	setLabelErr    error
	forwardingRule *computepb.ForwardingRule
}

func (a stubForwardingRulesAPI) Close() error {
	return nil
}

func (a stubForwardingRulesAPI) Insert(ctx context.Context, req *computepb.InsertForwardingRuleRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.insertErr != nil {
		return nil, a.insertErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name:   proto.String("name"),
			Region: proto.String("region"),
		},
	}, nil
}

func (a stubForwardingRulesAPI) Delete(ctx context.Context, req *computepb.DeleteForwardingRuleRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.deleteErr != nil {
		return nil, a.deleteErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name:   proto.String("name"),
			Region: proto.String("region"),
		},
	}, nil
}

func (a stubForwardingRulesAPI) Get(ctx context.Context, req *computepb.GetForwardingRuleRequest,
	opts ...gax.CallOption,
) (*computepb.ForwardingRule, error) {
	if a.getErr != nil {
		return nil, a.getErr
	}
	return a.forwardingRule, nil
}

func (a stubForwardingRulesAPI) SetLabels(ctx context.Context, req *computepb.SetLabelsForwardingRuleRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.deleteErr != nil {
		return nil, a.setLabelErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name:   proto.String("name"),
			Region: proto.String("region"),
		},
	}, nil
}

type stubHealthChecksAPI struct {
	insertErr error
	deleteErr error
}

func (a stubHealthChecksAPI) Close() error {
	return nil
}

func (a stubHealthChecksAPI) Insert(ctx context.Context, req *computepb.InsertRegionHealthCheckRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.insertErr != nil {
		return nil, a.insertErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name:   proto.String("name"),
			Region: proto.String("region"),
		},
	}, nil
}

func (a stubHealthChecksAPI) Delete(ctx context.Context, req *computepb.DeleteRegionHealthCheckRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.deleteErr != nil {
		return nil, a.deleteErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name:   proto.String("name"),
			Region: proto.String("region"),
		},
	}, nil
}

type stubInstanceTemplateAPI struct {
	deleteErr error
	insertErr error
}

func (a stubInstanceTemplateAPI) Close() error {
	return nil
}

func (a stubInstanceTemplateAPI) Delete(ctx context.Context, req *computepb.DeleteInstanceTemplateRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.deleteErr != nil {
		return nil, a.deleteErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name: proto.String("name"),
		},
	}, nil
}

func (a stubInstanceTemplateAPI) Insert(ctx context.Context, req *computepb.InsertInstanceTemplateRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.insertErr != nil {
		return nil, a.insertErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name: proto.String("name"),
		},
	}, nil
}

type stubInstanceGroupManagersAPI struct {
	listIterator *stubManagedInstanceIterator

	deleteErr error
	insertErr error
}

func (a stubInstanceGroupManagersAPI) Close() error {
	return nil
}

func (a stubInstanceGroupManagersAPI) Delete(ctx context.Context, req *computepb.DeleteInstanceGroupManagerRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.deleteErr != nil {
		return nil, a.deleteErr
	}
	return &stubOperation{
		&computepb.Operation{
			Zone: proto.String("zone"),
			Name: proto.String("name"),
		},
	}, nil
}

func (a stubInstanceGroupManagersAPI) Insert(ctx context.Context, req *computepb.InsertInstanceGroupManagerRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	if a.insertErr != nil {
		return nil, a.insertErr
	}
	return &stubOperation{
		&computepb.Operation{
			Zone: proto.String("zone"),
			Name: proto.String("name"),
		},
	}, nil
}

func (a stubInstanceGroupManagersAPI) ListManagedInstances(ctx context.Context, req *computepb.ListManagedInstancesInstanceGroupManagersRequest,
	opts ...gax.CallOption,
) ManagedInstanceIterator {
	return a.listIterator
}

type stubProjectsAPI struct {
	getPolicyErr error
	setPolicyErr error
}

func (a stubProjectsAPI) Close() error {
	return nil
}

func (a stubProjectsAPI) GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error) {
	if a.getPolicyErr != nil {
		return nil, a.getPolicyErr
	}
	return &iampb.Policy{
		Version: 3,
		Bindings: []*iampb.Binding{
			{
				Role: "role",
				Members: []string{
					"member",
				},
			},
		},
		Etag: []byte("etag"),
	}, nil
}

func (a stubProjectsAPI) SetIamPolicy(ctx context.Context, req *iampb.SetIamPolicyRequest, opts ...gax.CallOption) (*iampb.Policy, error) {
	if a.setPolicyErr != nil {
		return nil, a.setPolicyErr
	}
	return &iampb.Policy{
		Version: 3,
		Bindings: []*iampb.Binding{
			{
				Role: "role",
				Members: []string{
					"member",
				},
			},
		},
		Etag: []byte("etag"),
	}, nil
}

type stubManagedInstanceIterator struct {
	instances []*computepb.ManagedInstance
	nextErr   error

	internalCounter int
}

func (i *stubManagedInstanceIterator) Next() (*computepb.ManagedInstance, error) {
	if i.nextErr != nil {
		return nil, i.nextErr
	}
	if i.internalCounter >= len(i.instances) {
		i.internalCounter = 0
		return nil, iterator.Done
	}
	resp := i.instances[i.internalCounter]
	i.internalCounter++
	return resp, nil
}

type stubAddressesAPI struct {
	insertErr error
	getAddr   *string
	getErr    error
	deleteErr error
}

func (a stubAddressesAPI) Insert(context.Context, *computepb.InsertAddressRequest,
	...gax.CallOption,
) (Operation, error) {
	if a.insertErr != nil {
		return nil, a.insertErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name:   proto.String("name"),
			Region: proto.String("region"),
		},
	}, nil
}

func (a stubAddressesAPI) Get(ctx context.Context, req *computepb.GetAddressRequest,
	opts ...gax.CallOption,
) (*computepb.Address, error) {
	return &computepb.Address{Address: a.getAddr}, a.getErr
}

func (a stubAddressesAPI) Delete(context.Context, *computepb.DeleteAddressRequest,
	...gax.CallOption,
) (Operation, error) {
	if a.deleteErr != nil {
		return nil, a.deleteErr
	}
	return &stubOperation{
		&computepb.Operation{
			Name:   proto.String("name"),
			Region: proto.String("region"),
		},
	}, nil
}

func (a stubAddressesAPI) Close() error {
	return nil
}
