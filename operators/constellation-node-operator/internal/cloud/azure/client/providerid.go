/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package client

import (
	"errors"
	"regexp"
)

// azureVMSSProviderIDRegexp describes the format of a kubernetes providerID for an azure virtual machine scale set (VMSS) VM.
var azureVMSSProviderIDRegexp = regexp.MustCompile(`^azure:///subscriptions/([^/]+)/resourceGroups/([^/]+)/providers/Microsoft.Compute/virtualMachineScaleSets/([^/]+)/virtualMachines/([^/]+)$`)

// scaleSetInformationFromProviderID splits a provider's id belonging to an azure scale set into core components.
// A providerID for scale set VMs is build after the following schema:
// - 'azure:///subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.Compute/virtualMachineScaleSets/<scale-set-name>/virtualMachines/<instance-id>'
func scaleSetInformationFromProviderID(providerID string) (subscriptionID, resourceGroup, scaleSet, instanceID string, err error) {
	matches := azureVMSSProviderIDRegexp.FindStringSubmatch(providerID)
	if len(matches) != 5 {
		return "", "", "", "", errors.New("error splitting providerID")
	}
	return matches[1], matches[2], matches[3], matches[4], nil
}
