/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package openstack

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountKeyFromURI(t *testing.T) {
	accountKey := AccountKey{
		AuthURL:           "auth-url",
		Username:          "username",
		Password:          "password",
		ProjectID:         "project-id",
		ProjectName:       "project-name",
		UserDomainName:    "user-domain-name",
		ProjectDomainName: "project-domain-name",
		RegionName:        "region-name",
	}
	testCases := map[string]struct {
		cloudServiceAccountURI string
		wantKey                AccountKey
		wantErr                bool
	}{
		"successful": {
			cloudServiceAccountURI: "serviceaccount://openstack?auth_url=auth-url&username=username&password=password&project_id=project-id&project_name=project-name&user_domain_name=user-domain-name&project_domain_name=project-domain-name&region_name=region-name",
			wantKey:                accountKey,
		},
		"missing auth_url": {
			cloudServiceAccountURI: "serviceaccount://openstack?username=username&password=password&project_id=project-id&project_name=project-name&user_domain_name=user-domain-name&project_domain_name=project-domain-name&region_name=region-name",
			wantErr:                true,
		},
		"missing username": {
			cloudServiceAccountURI: "serviceaccount://openstack?auth_url=auth-url&password=password&project_id=project-id&project_name=project-name&user_domain_name=user-domain-name&project_domain_name=project-domain-name&region_name=region-name",
			wantErr:                true,
		},
		"missing password": {
			cloudServiceAccountURI: "serviceaccount://openstack?auth_url=auth-url&username=username&project_id=project-id&project_name=project-name&user_domain_name=user-domain-name&project_domain_name=project-domain-name&region_name=region-name",
			wantErr:                true,
		},
		"missing project_id": {
			cloudServiceAccountURI: "serviceaccount://openstack?auth_url=auth-url&username=username&password=password&project_name=project-name&user_domain_name=user-domain-name&project_domain_name=project-domain-name&region_name=region-name",
			wantErr:                true,
		},
		"missing project_name": {
			cloudServiceAccountURI: "serviceaccount://openstack?auth_url=auth-url&username=username&password=password&project_id=project-id&user_domain_name=user-domain-name&project_domain_name=project-domain-name&region_name=region-name",
			wantErr:                true,
		},
		"missing user_domain_name": {
			cloudServiceAccountURI: "serviceaccount://openstack?auth_url=auth-url&username=username&password=password&project_id=project-id&project_name=project-name&project_domain_name=project-domain-name&region_name=region-name",
			wantErr:                true,
		},
		"missing project_domain_name": {
			cloudServiceAccountURI: "serviceaccount://openstack?auth_url=auth-url&username=username&password=password&project_id=project-id&project_name=project-name&user_domain_name=user-domain-name&region_name=region-name",
			wantErr:                true,
		},
		"missing region_name": {
			cloudServiceAccountURI: "serviceaccount://openstack?auth_url=auth-url&username=username&password=password&project_id=project-id&project_name=project-name&user_domain_name=user-domain-name&project_domain_name=project-domain-name",
			wantErr:                true,
		},
		"invalid URI fails": {
			cloudServiceAccountURI: "\x00",
			wantErr:                true,
		},
		"incorrect URI scheme fails": {
			cloudServiceAccountURI: "invalid",
			wantErr:                true,
		},
		"incorrect URI host fails": {
			cloudServiceAccountURI: "serviceaccount://incorrect",
			wantErr:                true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			key, err := AccountKeyFromURI(tc.cloudServiceAccountURI)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantKey, key)
		})
	}
}

func TestConvertToCloudServiceAccountURI(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	key := AccountKey{
		AuthURL:           "auth-url",
		Username:          "username",
		Password:          "password",
		ProjectID:         "project-id",
		ProjectName:       "project-name",
		UserDomainName:    "user-domain-name",
		ProjectDomainName: "project-domain-name",
		RegionName:        "region-name",
	}
	accountURI := key.ToCloudServiceAccountURI()
	uri, err := url.Parse(accountURI)
	require.NoError(err)
	query := uri.Query()
	assert.Equal("serviceaccount", uri.Scheme)
	assert.Equal("openstack", uri.Host)
	assert.Equal(url.Values{
		"auth_url":            []string{"auth-url"},
		"username":            []string{"username"},
		"password":            []string{"password"},
		"project_id":          []string{"project-id"},
		"project_name":        []string{"project-name"},
		"user_domain_name":    []string{"user-domain-name"},
		"project_domain_name": []string{"project-domain-name"},
		"region_name":         []string{"region-name"},
	}, query)
}

func TestAccountKeyToCloudINI(t *testing.T) {
	assert := assert.New(t)
	key := AccountKey{
		AuthURL:           "auth-url",
		Username:          "username",
		Password:          "password",
		ProjectID:         "project-id",
		ProjectName:       "project-name",
		UserDomainName:    "user-domain-name",
		ProjectDomainName: "project-domain-name",
		RegionName:        "region-name",
	}
	ini := key.CloudINI()
	assert.Equal(CloudINI{
		AuthURL:          "auth-url",
		Username:         "username",
		Password:         "password",
		ProjectID:        "project-id",
		TenantName:       "project-name",
		UserDomainName:   "user-domain-name",
		TenantDomainName: "project-domain-name",
		Region:           "region-name",
	}, ini)
}

func TestFullConfiguration(t *testing.T) {
	ini := CloudINI{
		AuthURL:          "auth-url",
		Username:         "username",
		Password:         "password",
		ProjectID:        "project-id",
		TenantName:       "project-name",
		UserDomainName:   "user-domain-name",
		TenantDomainName: "project-domain-name",
		Region:           "region-name",
	}
	assert.Equal(t, `[Global]
auth-url = auth-url
username = username
password = password
tenant-id = project-id
tenant-name = project-name
user-domain-name = user-domain-name
tenant-domain-name = project-domain-name
region = region-name
`, ini.FullConfiguration())
}

func TestYawolConfiguration(t *testing.T) {
	ini := CloudINI{
		AuthURL:          "auth-url",
		Username:         "username",
		Password:         "password",
		ProjectID:        "project-id",
		TenantName:       "project-name",
		UserDomainName:   "user-domain-name",
		TenantDomainName: "project-domain-name",
		Region:           "region-name",
	}
	assert.Equal(t, `[Global]
auth-url = auth-url
username = username
password = password
project-id = project-id
domain-name = user-domain-name
region = region-name
`, ini.YawolConfiguration())
}

func TestCinderCSIConfiguration(t *testing.T) {
	ini := CloudINI{
		AuthURL:          "auth-url",
		Username:         "username",
		Password:         "password",
		ProjectID:        "project-id",
		TenantName:       "project-name",
		UserDomainName:   "user-domain-name",
		TenantDomainName: "project-domain-name",
		Region:           "region-name",
	}
	assert.Equal(t, `[Global]
auth-url = auth-url
username = username
password = password
project-id = project-id
project-name = project-name
user-domain-name = user-domain-name
project-domain-name = project-domain-name
region = region-name
`, ini.CinderCSIConfiguration())
}
