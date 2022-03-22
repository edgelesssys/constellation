package client

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestCreateServicePrincipal(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		applicationsAPI      applicationsAPI
		servicePrincipalsAPI servicePrincipalsAPI
		roleAssignmentsAPI   roleAssignmentsAPI
		resourceGroupAPI     resourceGroupAPI
		errExpected          bool
	}{
		"successful create": {
			applicationsAPI:      stubApplicationsAPI{},
			servicePrincipalsAPI: stubServicePrincipalsAPI{},
			roleAssignmentsAPI:   &stubRoleAssignmentsAPI{},
			resourceGroupAPI: stubResourceGroupAPI{
				getResourceGroup: armresources.ResourceGroup{
					ID: to.StringPtr("resource-group-id"),
				},
			},
		},
		"failed app create": {
			applicationsAPI: stubApplicationsAPI{
				createErr: someErr,
			},
			errExpected: true,
		},
		"failed service principal create": {
			applicationsAPI: stubApplicationsAPI{},
			servicePrincipalsAPI: stubServicePrincipalsAPI{
				createErr: someErr,
			},
			errExpected: true,
		},
		"failed role assignment": {
			applicationsAPI:      stubApplicationsAPI{},
			servicePrincipalsAPI: stubServicePrincipalsAPI{},
			roleAssignmentsAPI: &stubRoleAssignmentsAPI{
				createErrors: []error{someErr},
			},
			resourceGroupAPI: stubResourceGroupAPI{
				getResourceGroup: armresources.ResourceGroup{
					ID: to.StringPtr("resource-group-id"),
				},
			},
			errExpected: true,
		},
		"failed update creds": {
			applicationsAPI: stubApplicationsAPI{
				updateCredentialsErr: someErr,
			},
			servicePrincipalsAPI: stubServicePrincipalsAPI{},
			roleAssignmentsAPI:   &stubRoleAssignmentsAPI{},
			resourceGroupAPI: stubResourceGroupAPI{
				getResourceGroup: armresources.ResourceGroup{
					ID: to.StringPtr("resource-group-id"),
				},
			},
			errExpected: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()

			client := Client{
				name:                            "name",
				uid:                             "uid",
				resourceGroup:                   "resource-group",
				applicationsAPI:                 tc.applicationsAPI,
				servicePrincipalsAPI:            tc.servicePrincipalsAPI,
				roleAssignmentsAPI:              tc.roleAssignmentsAPI,
				resourceGroupAPI:                tc.resourceGroupAPI,
				adReplicationLagCheckMaxRetries: 2,
			}

			_, err := client.CreateServicePrincipal(ctx)
			if tc.errExpected {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestTerminateServicePrincipal(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		appObjectID     string
		applicationsAPI applicationsAPI
		errExpected     bool
	}{
		"successful terminate": {
			appObjectID:     "object-id",
			applicationsAPI: stubApplicationsAPI{},
		},
		"nothing to terminate": {
			applicationsAPI: stubApplicationsAPI{},
		},
		"failed delete": {
			appObjectID: "object-id",
			applicationsAPI: stubApplicationsAPI{
				deleteErr: someErr,
			},
			errExpected: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()

			client := Client{
				name:            "name",
				uid:             "uid",
				resourceGroup:   "resource-group",
				adAppObjectID:   tc.appObjectID,
				applicationsAPI: tc.applicationsAPI,
			}

			err := client.TerminateServicePrincipal(ctx)
			if tc.errExpected {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestCreateADApplication(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		applicationsAPI applicationsAPI
		errExpected     bool
	}{
		"successful create": {
			applicationsAPI: stubApplicationsAPI{},
		},
		"failed app create": {
			applicationsAPI: stubApplicationsAPI{
				createErr: someErr,
			},
			errExpected: true,
		},
		"app create returns invalid appid": {
			applicationsAPI: stubApplicationsAPI{
				createApplication: &graphrbac.Application{
					ObjectID: proto.String("00000000-0000-0000-0000-000000000001"),
				},
			},
			errExpected: true,
		},
		"app create returns invalid objectid": {
			applicationsAPI: stubApplicationsAPI{
				createApplication: &graphrbac.Application{
					AppID: proto.String("00000000-0000-0000-0000-000000000000"),
				},
			},
			errExpected: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()

			client := Client{
				name:            "name",
				uid:             "uid",
				applicationsAPI: tc.applicationsAPI,
			}

			appCredentials, err := client.createADApplication(ctx)
			if tc.errExpected {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.NotNil(appCredentials)
		})
	}
}

func TestCreateAppServicePrincipal(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		servicePrincipalsAPI servicePrincipalsAPI
		errExpected          bool
	}{
		"successful create": {
			servicePrincipalsAPI: stubServicePrincipalsAPI{},
		},
		"failed service principal create": {
			servicePrincipalsAPI: stubServicePrincipalsAPI{
				createErr: someErr,
			},
			errExpected: true,
		},
		"service principal create returns invalid objectid": {
			servicePrincipalsAPI: stubServicePrincipalsAPI{
				createServicePrincipal: &graphrbac.ServicePrincipal{},
			},
			errExpected: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()

			client := Client{
				name:                 "name",
				uid:                  "uid",
				servicePrincipalsAPI: tc.servicePrincipalsAPI,
			}

			_, err := client.createAppServicePrincipal(ctx, "app-id")
			if tc.errExpected {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestAssignOwnerOfResourceGroup(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		roleAssignmentsAPI roleAssignmentsAPI
		resourceGroupAPI   resourceGroupAPI
		errExpected        bool
	}{
		"successful assign": {
			roleAssignmentsAPI: &stubRoleAssignmentsAPI{},
			resourceGroupAPI: stubResourceGroupAPI{
				getResourceGroup: armresources.ResourceGroup{
					ID: to.StringPtr("resource-group-id"),
				},
			},
		},
		"failed role assignment": {
			roleAssignmentsAPI: &stubRoleAssignmentsAPI{
				createErrors: []error{someErr},
			},
			resourceGroupAPI: stubResourceGroupAPI{
				getResourceGroup: armresources.ResourceGroup{
					ID: to.StringPtr("resource-group-id"),
				},
			},
			errExpected: true,
		},
		"failed resource group get": {
			roleAssignmentsAPI: &stubRoleAssignmentsAPI{},
			resourceGroupAPI: stubResourceGroupAPI{
				getErr: someErr,
			},
			errExpected: true,
		},
		"resource group get returns invalid id": {
			roleAssignmentsAPI: &stubRoleAssignmentsAPI{},
			resourceGroupAPI: stubResourceGroupAPI{
				getResourceGroup: armresources.ResourceGroup{},
			},
			errExpected: true,
		},
		"create returns PrincipalNotFound the first time": {
			roleAssignmentsAPI: &stubRoleAssignmentsAPI{
				createErrors: []error{
					autorest.DetailedError{Original: &azure.RequestError{
						ServiceError: &azure.ServiceError{
							Code: "PrincipalNotFound",
						},
					}},
					nil,
				},
			},
			resourceGroupAPI: stubResourceGroupAPI{
				getResourceGroup: armresources.ResourceGroup{
					ID: to.StringPtr("resource-group-id"),
				},
			},
		},
		"create does not return request error": {
			roleAssignmentsAPI: &stubRoleAssignmentsAPI{
				createErrors: []error{autorest.DetailedError{Original: someErr}},
			},
			resourceGroupAPI: stubResourceGroupAPI{
				getResourceGroup: armresources.ResourceGroup{
					ID: to.StringPtr("resource-group-id"),
				},
			},
			errExpected: true,
		},
		"create service error code is unknown": {
			roleAssignmentsAPI: &stubRoleAssignmentsAPI{
				createErrors: []error{
					autorest.DetailedError{Original: &azure.RequestError{
						ServiceError: &azure.ServiceError{
							Code: "some-unknown-error-code",
						},
					}},
					nil,
				},
			},
			resourceGroupAPI: stubResourceGroupAPI{
				getResourceGroup: armresources.ResourceGroup{
					ID: to.StringPtr("resource-group-id"),
				},
			},
			errExpected: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()

			client := Client{
				name:                            "name",
				uid:                             "uid",
				resourceGroup:                   "resource-group",
				roleAssignmentsAPI:              tc.roleAssignmentsAPI,
				resourceGroupAPI:                tc.resourceGroupAPI,
				adReplicationLagCheckMaxRetries: 2,
			}

			err := client.assignResourceGroupRole(ctx, "principal-id", "role-definition-id")
			if tc.errExpected {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestConvertToCloudServiceAccountURI(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	key := ApplicationCredentials{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}

	cloudServiceAccountURI := key.ConvertToCloudServiceAccountURI()
	uri, err := url.Parse(cloudServiceAccountURI)
	require.NoError(err)
	query := uri.Query()
	assert.Equal("serviceaccount", uri.Scheme)
	assert.Equal("azure", uri.Host)
	assert.Equal(url.Values{
		"client_id":     []string{"client-id"},
		"client_secret": []string{"client-secret"},
	}, query)
}
