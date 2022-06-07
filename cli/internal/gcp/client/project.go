package client

import (
	"context"
	"fmt"

	iampb "google.golang.org/genproto/googleapis/iam/v1"
)

// addIAMPolicyBindings adds a GCP service account to roles specified in the input.
func (c *Client) addIAMPolicyBindings(ctx context.Context, input AddIAMPolicyBindingInput) error {
	getReq := &iampb.GetIamPolicyRequest{
		Resource: "projects/" + c.project,
	}
	policy, err := c.projectsAPI.GetIamPolicy(ctx, getReq)
	if err != nil {
		return fmt.Errorf("retrieving current iam policy failed: %w", err)
	}
	for _, binding := range input.Bindings {
		addIAMPolicy(policy, binding)
	}
	setReq := &iampb.SetIamPolicyRequest{
		Resource: "projects/" + c.project,
		Policy:   policy,
	}
	if _, err := c.projectsAPI.SetIamPolicy(ctx, setReq); err != nil {
		return fmt.Errorf("setting new iam policy failed: %w", err)
	}
	return nil
}

// PolicyBinding is a GCP IAM policy binding.
type PolicyBinding struct {
	ServiceAccount string
	Role           string
}

// addIAMPolicy inserts policy binding for service account and role to an existing iam policy.
func addIAMPolicy(policy *iampb.Policy, policyBinding PolicyBinding) {
	var binding *iampb.Binding
	for _, existingBinding := range policy.Bindings {
		if existingBinding.Role == policyBinding.Role && existingBinding.Condition == nil {
			binding = existingBinding
			break
		}
	}
	if binding == nil {
		binding = &iampb.Binding{
			Role: policyBinding.Role,
		}
		policy.Bindings = append(policy.Bindings, binding)
	}

	// add service account to role, if not already a member
	member := "serviceAccount:" + policyBinding.ServiceAccount
	var alreadyMember bool
	for _, existingMember := range binding.Members {
		if member == existingMember {
			alreadyMember = true
			break
		}
	}
	if !alreadyMember {
		binding.Members = append(binding.Members, member)
	}
}

// AddIAMPolicyBindingInput is the input for an AddIAMPolicyBinding operation.
type AddIAMPolicyBindingInput struct {
	Bindings []PolicyBinding
}
