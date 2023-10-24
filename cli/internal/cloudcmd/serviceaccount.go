/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/cloud/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/openstack"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

// GetMarshaledServiceAccountURI returns the service account URI for the given cloud provider.
func GetMarshaledServiceAccountURI(config *config.Config, fileHandler file.Handler) (string, error) {
	switch config.GetProvider() {
	case cloudprovider.GCP:
		var key gcpshared.ServiceAccountKey
		if err := fileHandler.ReadJSON(config.Provider.GCP.ServiceAccountKeyPath, &key); err != nil {
			return "", fmt.Errorf("reading service account key: %w", err)
		}
		return key.ToCloudServiceAccountURI(), nil

	case cloudprovider.AWS:
		return "", nil // AWS does not need a service account URI

	case cloudprovider.Azure:
		authMethod := azureshared.AuthMethodUserAssignedIdentity

		creds := azureshared.ApplicationCredentials{
			TenantID:            config.Provider.Azure.TenantID,
			Location:            config.Provider.Azure.Location,
			PreferredAuthMethod: authMethod,
			UamiResourceID:      config.Provider.Azure.UserAssignedIdentity,
		}
		return creds.ToCloudServiceAccountURI(), nil

	case cloudprovider.OpenStack:
		creds := openstack.AccountKey{
			AuthURL:           config.Provider.OpenStack.AuthURL,
			Username:          config.Provider.OpenStack.Username,
			Password:          config.Provider.OpenStack.Password,
			ProjectID:         config.Provider.OpenStack.ProjectID,
			ProjectName:       config.Provider.OpenStack.ProjectName,
			UserDomainName:    config.Provider.OpenStack.UserDomainName,
			ProjectDomainName: config.Provider.OpenStack.ProjectDomainName,
			RegionName:        config.Provider.OpenStack.RegionName,
		}
		return creds.ToCloudServiceAccountURI(), nil

	case cloudprovider.QEMU:
		return "", nil // QEMU does not use service account keys

	default:
		return "", fmt.Errorf("unsupported cloud provider %q", config.GetProvider())
	}
}
