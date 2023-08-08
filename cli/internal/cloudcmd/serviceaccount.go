/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package cmd provides the Constellation CLI.

It is responsible for the interaction with the user.
*/
package cloudcmd

import (
	"fmt"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/internal/cloud/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/openstack"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

// GetMarshaledServiceAccountURI returns the service account URI for the given cloud provider.
func GetMarshaledServiceAccountURI(provider cloudprovider.Provider, config *config.Config, workspace string, log debugLog, fileHandler file.Handler,
) (string, error) {
	log.Debugf("Getting service account URI")
	switch provider {
	case cloudprovider.GCP:
		log.Debugf("Handling case for GCP")
		log.Debugf("GCP service account key path %s", filepath.Join(workspace, config.Provider.GCP.ServiceAccountKeyPath))

		var key gcpshared.ServiceAccountKey
		if err := fileHandler.ReadJSON(config.Provider.GCP.ServiceAccountKeyPath, &key); err != nil {
			return "", fmt.Errorf("reading service account key from path %q: %w", filepath.Join(workspace, config.Provider.GCP.ServiceAccountKeyPath), err)
		}
		log.Debugf("Read GCP service account key from path")
		return key.ToCloudServiceAccountURI(), nil

	case cloudprovider.AWS:
		log.Debugf("Handling case for AWS")
		return "", nil // AWS does not need a service account URI
	case cloudprovider.Azure:
		log.Debugf("Handling case for Azure")

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
		log.Debugf("Handling case for QEMU")
		return "", nil // QEMU does not use service account keys

	default:
		return "", fmt.Errorf("unsupported cloud provider %q", provider)
	}
}
