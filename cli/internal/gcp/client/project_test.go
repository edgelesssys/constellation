package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
	"google.golang.org/protobuf/proto"
)

func TestAddIAMPolicyBindings(t *testing.T) {
	someErr := errors.New("someErr")

	testCases := map[string]struct {
		projectsAPI stubProjectsAPI
		input       AddIAMPolicyBindingInput
		wantErr     bool
	}{
		"successful set without new bindings": {
			input: AddIAMPolicyBindingInput{
				Bindings: []PolicyBinding{},
			},
		},
		"successful set with bindings": {
			input: AddIAMPolicyBindingInput{
				Bindings: []PolicyBinding{
					{
						ServiceAccount: "service-account",
						Role:           "role",
					},
				},
			},
		},
		"retrieving iam policy fails": {
			projectsAPI: stubProjectsAPI{
				getPolicyErr: someErr,
			},
			wantErr: true,
		},
		"setting iam policy fails": {
			projectsAPI: stubProjectsAPI{
				setPolicyErr: someErr,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:     "project",
				zone:        "zone",
				name:        "name",
				uid:         "uid",
				projectsAPI: tc.projectsAPI,
			}

			err := client.addIAMPolicyBindings(ctx, tc.input)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestAddIAMPolicy(t *testing.T) {
	testCases := map[string]struct {
		binding    PolicyBinding
		policy     *iampb.Policy
		wantErr    bool
		wantPolicy *iampb.Policy
	}{
		"successful on empty policy": {
			binding: PolicyBinding{
				ServiceAccount: "service-account",
				Role:           "role",
			},
			policy: &iampb.Policy{
				Bindings: []*iampb.Binding{},
			},
			wantPolicy: &iampb.Policy{
				Bindings: []*iampb.Binding{
					{
						Role:    "role",
						Members: []string{"serviceAccount:service-account"},
					},
				},
			},
		},
		"successful on existing policy with different role": {
			binding: PolicyBinding{
				ServiceAccount: "service-account",
				Role:           "role",
			},
			policy: &iampb.Policy{
				Bindings: []*iampb.Binding{
					{
						Role:    "other-role",
						Members: []string{"other-member"},
					},
				},
			},
			wantPolicy: &iampb.Policy{
				Bindings: []*iampb.Binding{
					{
						Role:    "other-role",
						Members: []string{"other-member"},
					},
					{
						Role:    "role",
						Members: []string{"serviceAccount:service-account"},
					},
				},
			},
		},
		"successful on existing policy with existing role": {
			binding: PolicyBinding{
				ServiceAccount: "service-account",
				Role:           "role",
			},
			policy: &iampb.Policy{
				Bindings: []*iampb.Binding{
					{
						Role:    "role",
						Members: []string{"other-member"},
					},
				},
			},
			wantPolicy: &iampb.Policy{
				Bindings: []*iampb.Binding{
					{
						Role:    "role",
						Members: []string{"other-member", "serviceAccount:service-account"},
					},
				},
			},
		},
		"already a member": {
			binding: PolicyBinding{
				ServiceAccount: "service-account",
				Role:           "role",
			},
			policy: &iampb.Policy{
				Bindings: []*iampb.Binding{
					{
						Role:    "role",
						Members: []string{"serviceAccount:service-account"},
					},
				},
			},
			wantPolicy: &iampb.Policy{
				Bindings: []*iampb.Binding{
					{
						Role:    "role",
						Members: []string{"serviceAccount:service-account"},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			addIAMPolicy(tc.policy, tc.binding)
			assert.True(proto.Equal(tc.wantPolicy, tc.policy))
		})
	}
}
