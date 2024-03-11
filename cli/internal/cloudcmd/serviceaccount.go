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
	"github.com/edgelesssys/constellation/v2/internal/cloud/openstack/clouds"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constellation"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

// GetMarshaledServiceAccountURI returns the service account URI for the given cloud provider.
func GetMarshaledServiceAccountURI(config *config.Config, fileHandler file.Handler) (string, error) {
	payload := constellation.ServiceAccountPayload{}
	switch config.GetProvider() {
	case cloudprovider.GCP:
		var key gcpshared.ServiceAccountKey
		if err := fileHandler.ReadJSON(config.Provider.GCP.ServiceAccountKeyPath, &key); err != nil {
			return "", fmt.Errorf("reading service account key: %w", err)
		}
		payload.GCP = key

	case cloudprovider.Azure:
		payload.Azure = azureshared.ApplicationCredentials{
			TenantID:            config.Provider.Azure.TenantID,
			Location:            config.Provider.Azure.Location,
			PreferredAuthMethod: azureshared.AuthMethodUserAssignedIdentity,
			UamiResourceID:      config.Provider.Azure.UserAssignedIdentity,
		}

	case cloudprovider.OpenStack:
		cloudsYAML, err := clouds.ReadCloudsYAML(fileHandler, config.Provider.OpenStack.CloudsYAMLPath)
		if err != nil {
			return "", fmt.Errorf("reading clouds.yaml: %w", err)
		}
		cloud, ok := cloudsYAML.Clouds[config.Provider.OpenStack.Cloud]
		if !ok {
			return "", fmt.Errorf("cloud %q not found in clouds.yaml", config.Provider.OpenStack.Cloud)
		}
		payload.OpenStack = openstack.AccountKey{
			AuthURL:           cloud.AuthInfo.AuthURL,
			Username:          cloud.AuthInfo.Username,
			Password:          cloud.AuthInfo.Password,
			ProjectID:         cloud.AuthInfo.ProjectID,
			ProjectName:       cloud.AuthInfo.ProjectName,
			UserDomainName:    cloud.AuthInfo.UserDomainName,
			ProjectDomainName: cloud.AuthInfo.ProjectDomainName,
			RegionName:        cloud.RegionName,
		}

	}
	return constellation.MarshalServiceAccountURI(config.GetProvider(), payload)
}
