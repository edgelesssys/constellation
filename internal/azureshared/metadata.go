/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azureshared

import (
	"errors"
	"fmt"
	"regexp"
)

var azureVMSSProviderIDRegexp = regexp.MustCompile(`^azure:///subscriptions/([^/]+)/resourceGroups/([^/]+)/providers/Microsoft.Compute/virtualMachineScaleSets/([^/]+)/virtualMachines/([^/]+)$`)

// BasicsFromProviderID extracts subscriptionID and resourceGroup from both types of valid azure providerID.
func BasicsFromProviderID(providerID string) (subscriptionID, resourceGroup string, err error) {
	subscriptionID, resourceGroup, _, _, err = ScaleSetInformationFromProviderID(providerID)
	if err == nil {
		return subscriptionID, resourceGroup, nil
	}
	return "", "", fmt.Errorf("providerID %v is malformatted", providerID)
}

// ScaleSetInformationFromProviderID splits a provider's id belonging to an azure scaleset into core components.
// A providerID for scale set VMs is build after the following schema:
// - 'azure:///subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.Compute/virtualMachineScaleSets/<scale-set-name>/virtualMachines/<instance-id>'
func ScaleSetInformationFromProviderID(providerID string) (subscriptionID, resourceGroup, scaleSet, instanceID string, err error) {
	matches := azureVMSSProviderIDRegexp.FindStringSubmatch(providerID)
	if len(matches) != 5 {
		return "", "", "", "", errors.New("error splitting providerID")
	}
	return matches[1], matches[2], matches[3], matches[4], nil
}
