package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/edgelesssys/constellation/internal/gcpshared"
	adminpb "google.golang.org/genproto/googleapis/iam/admin/v1"
)

// CreateServiceAccount creates a new GCP service account and returns an account key as service account URI.
func (c *Client) CreateServiceAccount(ctx context.Context, input ServiceAccountInput) (string, error) {
	insertInput := insertServiceAccountInput{
		Project:     c.project,
		AccountID:   "constellation-app-" + c.uid,
		DisplayName: "constellation-app-" + c.uid,
		Description: "This service account belongs to a Constellation cluster.",
	}

	email, err := c.insertServiceAccount(ctx, insertInput)
	if err != nil {
		return "", err
	}
	c.serviceAccount = email

	iamInput := input.addIAMPolicyBindingInput(c.serviceAccount)
	if err := c.addIAMPolicyBindings(ctx, iamInput); err != nil {
		return "", err
	}

	key, err := c.createServiceAccountKey(ctx, email)
	if err != nil {
		return "", err
	}

	return key.ToCloudServiceAccountURI(), nil
}

func (c *Client) TerminateServiceAccount(ctx context.Context) error {
	if c.serviceAccount != "" {
		req := &adminpb.DeleteServiceAccountRequest{
			Name: "projects/-/serviceAccounts/" + c.serviceAccount,
		}
		if err := c.iamAPI.DeleteServiceAccount(ctx, req); err != nil {
			return fmt.Errorf("deleting service account failed: %w", err)
		}
		c.serviceAccount = ""
	}
	return nil
}

type ServiceAccountInput struct {
	Roles []string
}

func (i ServiceAccountInput) addIAMPolicyBindingInput(serviceAccount string) AddIAMPolicyBindingInput {
	iamPolicyBindingInput := AddIAMPolicyBindingInput{
		Bindings: make([]PolicyBinding, len(i.Roles)),
	}
	for i, role := range i.Roles {
		iamPolicyBindingInput.Bindings[i] = PolicyBinding{
			ServiceAccount: serviceAccount,
			Role:           role,
		}
	}
	return iamPolicyBindingInput
}

func (c *Client) insertServiceAccount(ctx context.Context, input insertServiceAccountInput) (string, error) {
	req := input.createServiceAccountRequest()
	account, err := c.iamAPI.CreateServiceAccount(ctx, req)
	if err != nil {
		return "", err
	}

	return account.Email, nil
}

func (c *Client) createServiceAccountKey(ctx context.Context, email string) (gcpshared.ServiceAccountKey, error) {
	req := createServiceAccountKeyRequest(email)
	key, err := c.iamAPI.CreateServiceAccountKey(ctx, req)
	if err != nil {
		return gcpshared.ServiceAccountKey{}, fmt.Errorf("creating service account key failed: %w", err)
	}
	var serviceAccountKey gcpshared.ServiceAccountKey
	if err := json.Unmarshal(key.PrivateKeyData, &serviceAccountKey); err != nil {
		return gcpshared.ServiceAccountKey{}, fmt.Errorf("decoding service account key JSON failed: %w", err)
	}

	return serviceAccountKey, nil
}

// insertServiceAccountInput is the input for a createServiceAccount operation.
type insertServiceAccountInput struct {
	Project     string
	AccountID   string
	DisplayName string
	Description string
}

func (c insertServiceAccountInput) createServiceAccountRequest() *adminpb.CreateServiceAccountRequest {
	return &adminpb.CreateServiceAccountRequest{
		Name:      "projects/" + c.Project,
		AccountId: c.AccountID,
		ServiceAccount: &adminpb.ServiceAccount{
			DisplayName: c.DisplayName,
			Description: c.Description,
		},
	}
}

func createServiceAccountKeyRequest(email string) *adminpb.CreateServiceAccountKeyRequest {
	return &adminpb.CreateServiceAccountKeyRequest{
		Name: "projects/-/serviceAccounts/" + email,
	}
}
