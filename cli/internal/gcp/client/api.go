/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"

	"github.com/googleapis/gax-go/v2"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	adminpb "google.golang.org/genproto/googleapis/iam/admin/v1"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
)

type instanceAPI interface {
	Close() error
	List(ctx context.Context, req *computepb.ListInstancesRequest,
		opts ...gax.CallOption) InstanceIterator
}

type operationRegionAPI interface {
	Close() error
	Wait(ctx context.Context, req *computepb.WaitRegionOperationRequest,
		opts ...gax.CallOption) (*computepb.Operation, error)
}

type operationZoneAPI interface {
	Close() error
	Wait(ctx context.Context, req *computepb.WaitZoneOperationRequest,
		opts ...gax.CallOption) (*computepb.Operation, error)
}

type operationGlobalAPI interface {
	Close() error
	Wait(ctx context.Context, req *computepb.WaitGlobalOperationRequest,
		opts ...gax.CallOption) (*computepb.Operation, error)
}

type firewallsAPI interface {
	Close() error
	Delete(ctx context.Context, req *computepb.DeleteFirewallRequest,
		opts ...gax.CallOption) (Operation, error)
	Insert(ctx context.Context, req *computepb.InsertFirewallRequest,
		opts ...gax.CallOption) (Operation, error)
}

type forwardingRulesAPI interface {
	Close() error
	Delete(ctx context.Context, req *computepb.DeleteGlobalForwardingRuleRequest,
		opts ...gax.CallOption) (Operation, error)
	Insert(ctx context.Context, req *computepb.InsertGlobalForwardingRuleRequest,
		opts ...gax.CallOption) (Operation, error)
	Get(ctx context.Context, req *computepb.GetGlobalForwardingRuleRequest,
		opts ...gax.CallOption) (*computepb.ForwardingRule, error)
	SetLabels(ctx context.Context, req *computepb.SetLabelsGlobalForwardingRuleRequest,
		opts ...gax.CallOption) (Operation, error)
}

type backendServicesAPI interface {
	Close() error
	Delete(ctx context.Context, req *computepb.DeleteBackendServiceRequest,
		opts ...gax.CallOption) (Operation, error)
	Insert(ctx context.Context, req *computepb.InsertBackendServiceRequest,
		opts ...gax.CallOption) (Operation, error)
}

type healthChecksAPI interface {
	Close() error
	Delete(ctx context.Context, req *computepb.DeleteHealthCheckRequest,
		opts ...gax.CallOption) (Operation, error)
	Insert(ctx context.Context, req *computepb.InsertHealthCheckRequest,
		opts ...gax.CallOption) (Operation, error)
}

type targetTCPProxiesAPI interface {
	Close() error
	Delete(ctx context.Context, req *computepb.DeleteTargetTcpProxyRequest,
		opts ...gax.CallOption) (Operation, error)
	Insert(ctx context.Context, req *computepb.InsertTargetTcpProxyRequest,
		opts ...gax.CallOption) (Operation, error)
}

type networksAPI interface {
	Close() error
	Delete(ctx context.Context, req *computepb.DeleteNetworkRequest,
		opts ...gax.CallOption) (Operation, error)
	Insert(ctx context.Context, req *computepb.InsertNetworkRequest,
		opts ...gax.CallOption) (Operation, error)
}

type subnetworksAPI interface {
	Close() error
	Delete(ctx context.Context, req *computepb.DeleteSubnetworkRequest,
		opts ...gax.CallOption) (Operation, error)
	Insert(ctx context.Context, req *computepb.InsertSubnetworkRequest,
		opts ...gax.CallOption) (Operation, error)
}

type instanceTemplateAPI interface {
	Close() error
	Delete(ctx context.Context, req *computepb.DeleteInstanceTemplateRequest,
		opts ...gax.CallOption) (Operation, error)
	Insert(ctx context.Context, req *computepb.InsertInstanceTemplateRequest,
		opts ...gax.CallOption) (Operation, error)
}

type instanceGroupManagersAPI interface {
	Close() error
	Delete(ctx context.Context, req *computepb.DeleteInstanceGroupManagerRequest,
		opts ...gax.CallOption) (Operation, error)
	Insert(ctx context.Context, req *computepb.InsertInstanceGroupManagerRequest,
		opts ...gax.CallOption) (Operation, error)
	ListManagedInstances(ctx context.Context, req *computepb.ListManagedInstancesInstanceGroupManagersRequest,
		opts ...gax.CallOption) ManagedInstanceIterator
}

type iamAPI interface {
	Close() error
	CreateServiceAccount(ctx context.Context, req *adminpb.CreateServiceAccountRequest,
		opts ...gax.CallOption) (*adminpb.ServiceAccount, error)
	CreateServiceAccountKey(ctx context.Context, req *adminpb.CreateServiceAccountKeyRequest,
		opts ...gax.CallOption) (*adminpb.ServiceAccountKey, error)
	DeleteServiceAccount(ctx context.Context, req *adminpb.DeleteServiceAccountRequest,
		opts ...gax.CallOption) error
}

type projectsAPI interface {
	Close() error
	GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest,
		opts ...gax.CallOption) (*iampb.Policy, error)
	SetIamPolicy(ctx context.Context, req *iampb.SetIamPolicyRequest,
		opts ...gax.CallOption) (*iampb.Policy, error)
}

type addressesAPI interface {
	Close() error
	Insert(ctx context.Context, req *computepb.InsertGlobalAddressRequest,
		opts ...gax.CallOption) (Operation, error)
	Get(ctx context.Context, req *computepb.GetGlobalAddressRequest,
		opts ...gax.CallOption) (*computepb.Address, error)
	Delete(ctx context.Context, req *computepb.DeleteGlobalAddressRequest,
		opts ...gax.CallOption) (Operation, error)
}

type Operation interface {
	Proto() *computepb.Operation
}

type ManagedInstanceIterator interface {
	Next() (*computepb.ManagedInstance, error)
}

type InstanceIterator interface {
	Next() (*computepb.Instance, error)
}
