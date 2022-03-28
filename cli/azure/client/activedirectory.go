package client

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/uuid"
)

const (
	adAppCredentialValidity                   = time.Hour * 24 * 365 * 5                                // ~5 years
	adReplicationLagCheckInterval             = time.Second * 5                                         // 5 seconds
	adReplicationLagCheckMaxRetries           = int((15 * time.Minute) / adReplicationLagCheckInterval) // wait for up to 15 minutes for AD replication
	ownerRoleDefinitionID                     = "8e3af657-a8ff-443c-a75c-2fe8c4bcb635"                  // https://docs.microsoft.com/en-us/azure/role-based-access-control/built-in-roles#owner
	virtualMachineContributorRoleDefinitionID = "9980e02c-c2be-4d73-94e8-173b1dc7cf3c"                  // https://docs.microsoft.com/en-us/azure/role-based-access-control/built-in-roles#virtual-machine-contributor
)

// CreateServicePrincipal creates an Azure AD app with a service principal, gives it "Owner" role on the resource group and creates new credentials.
func (c *Client) CreateServicePrincipal(ctx context.Context) (string, error) {
	createAppRes, err := c.createADApplication(ctx)
	if err != nil {
		return "", err
	}
	c.adAppObjectID = createAppRes.ObjectID
	servicePrincipalObjectID, err := c.createAppServicePrincipal(ctx, createAppRes.AppID)
	if err != nil {
		return "", err
	}

	if err := c.assignResourceGroupRole(ctx, servicePrincipalObjectID, ownerRoleDefinitionID); err != nil {
		return "", err
	}

	clientSecret, err := c.updateAppCredentials(ctx, createAppRes.ObjectID)
	if err != nil {
		return "", err
	}

	return ApplicationCredentials{
		TenantID:     c.tenantID,
		ClientID:     createAppRes.AppID,
		ClientSecret: clientSecret,
	}.ConvertToCloudServiceAccountURI(), nil
}

// TerminateServicePrincipal terminates an Azure AD app together with the service principal.
func (c *Client) TerminateServicePrincipal(ctx context.Context) error {
	if c.adAppObjectID == "" {
		return nil
	}
	if _, err := c.applicationsAPI.Delete(ctx, c.adAppObjectID); err != nil {
		return err
	}
	c.adAppObjectID = ""
	return nil
}

// createADApplication creates a new azure AD app.
func (c *Client) createADApplication(ctx context.Context) (createADApplicationOutput, error) {
	createParameters := graphrbac.ApplicationCreateParameters{
		AvailableToOtherTenants: to.BoolPtr(false),
		DisplayName:             to.StringPtr("constellation-app-" + c.name + "-" + c.uid),
	}
	app, err := c.applicationsAPI.Create(ctx, createParameters)
	if err != nil {
		return createADApplicationOutput{}, err
	}
	if app.AppID == nil || app.ObjectID == nil {
		return createADApplicationOutput{}, errors.New("creating AD application did not result in valid app id and object id")
	}
	return createADApplicationOutput{
		AppID:    *app.AppID,
		ObjectID: *app.ObjectID,
	}, nil
}

// createAppServicePrincipal creates a new service principal for an azure AD app.
func (c *Client) createAppServicePrincipal(ctx context.Context, appID string) (string, error) {
	createParameters := graphrbac.ServicePrincipalCreateParameters{
		AppID:          &appID,
		AccountEnabled: to.BoolPtr(true),
	}
	servicePrincipal, err := c.servicePrincipalsAPI.Create(ctx, createParameters)
	if err != nil {
		return "", err
	}
	if servicePrincipal.ObjectID == nil {
		return "", errors.New("creating AD service principal did not result in a valid object id")
	}
	return *servicePrincipal.ObjectID, nil
}

// updateAppCredentials sets app client-secret for authentication.
func (c *Client) updateAppCredentials(ctx context.Context, objectID string) (string, error) {
	keyID := uuid.New().String()
	clientSecret, err := generateClientSecret()
	if err != nil {
		return "", fmt.Errorf("generating client secret failed: %w", err)
	}
	updateParameters := graphrbac.PasswordCredentialsUpdateParameters{
		Value: &[]graphrbac.PasswordCredential{
			{
				StartDate: &date.Time{Time: time.Now()},
				EndDate:   &date.Time{Time: time.Now().Add(adAppCredentialValidity)},
				Value:     to.StringPtr(clientSecret),
				KeyID:     to.StringPtr(keyID),
			},
		},
	}
	_, err = c.applicationsAPI.UpdatePasswordCredentials(ctx, objectID, updateParameters)
	if err != nil {
		return "", err
	}
	return clientSecret, nil
}

// assignResourceGroupRole assigns the service principal a role at resource group scope.
func (c *Client) assignResourceGroupRole(ctx context.Context, principalID, roleDefinitionID string) error {
	resourceGroup, err := c.resourceGroupAPI.Get(ctx, c.resourceGroup, nil)
	if err != nil || resourceGroup.ID == nil {
		return fmt.Errorf("unable to retrieve resource group id for group %v: %w", c.resourceGroup, err)
	}
	roleAssignmentID := uuid.New().String()
	createParameters := authorization.RoleAssignmentCreateParameters{
		Properties: &authorization.RoleAssignmentProperties{
			PrincipalID:      to.StringPtr(principalID),
			RoleDefinitionID: to.StringPtr(fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s", c.subscriptionID, roleDefinitionID)),
		},
	}

	// due to an azure AD replication lag, retry role assignment if principal does not exist yet
	// reference: https://docs.microsoft.com/en-us/azure/role-based-access-control/role-assignments-rest#new-service-principal
	// proper fix: use API version 2018-09-01-preview or later
	// azure go sdk currently uses version 2015-07-01: https://github.com/Azure/azure-sdk-for-go/blob/v62.0.0/services/authorization/mgmt/2015-07-01/authorization/roleassignments.go#L95
	// the newer version "armauthorization.RoleAssignmentsClient" is currently broken: https://github.com/Azure/azure-sdk-for-go/issues/17071
	for i := 0; i < c.adReplicationLagCheckMaxRetries; i++ {
		_, err = c.roleAssignmentsAPI.Create(ctx, *resourceGroup.ID, roleAssignmentID, createParameters)
		var detailedErr autorest.DetailedError
		var ok bool
		if detailedErr, ok = err.(autorest.DetailedError); !ok {
			return err
		}
		var requestErr *azure.RequestError
		if requestErr, ok = detailedErr.Original.(*azure.RequestError); !ok || requestErr.ServiceError == nil {
			return err
		}
		if requestErr.ServiceError.Code != "PrincipalNotFound" {
			return err
		}
		time.Sleep(c.adReplicationLagCheckInterval)
	}
	return err
}

// ApplicationCredentials is a set of Azure AD application credentials.
// It is the equivalent of a service account key in other cloud providers.
type ApplicationCredentials struct {
	TenantID     string
	ClientID     string
	ClientSecret string
}

// ConvertToCloudServiceAccountURI converts the ApplicationCredentials into a cloud service account URI.
func (c ApplicationCredentials) ConvertToCloudServiceAccountURI() string {
	query := url.Values{}
	query.Add("tenant_id", c.TenantID)
	query.Add("client_id", c.ClientID)
	query.Add("client_secret", c.ClientSecret)
	uri := url.URL{
		Scheme:   "serviceaccount",
		Host:     "azure",
		RawQuery: query.Encode(),
	}
	return uri.String()
}

type createADApplicationOutput struct {
	AppID    string
	ObjectID string
}

func generateClientSecret() (string, error) {
	letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	pwLen := 64
	pw := make([]byte, 0, pwLen)
	for i := 0; i < pwLen; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		pw = append(pw, letters[n.Int64()])
	}
	return string(pw), nil
}
