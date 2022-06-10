package azureshared

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	azureVMProviderIDRegexp   = regexp.MustCompile(`^azure:///subscriptions/([^/]+)/resourceGroups/([^/]+)/providers/Microsoft.Compute/virtualMachines/([^/]+)$`)
	azureVMSSProviderIDRegexp = regexp.MustCompile(`^azure:///subscriptions/([^/]+)/resourceGroups/([^/]+)/providers/Microsoft.Compute/virtualMachineScaleSets/([^/]+)/virtualMachines/([^/]+)$`)
)

// BasicsFromProviderID extracts subscriptionID and resourceGroup from both types of valid azure providerID.
func BasicsFromProviderID(providerID string) (subscriptionID, resourceGroup string, err error) {
	subscriptionID, resourceGroup, _, err = VMInformationFromProviderID(providerID)
	if err == nil {
		return subscriptionID, resourceGroup, nil
	}
	subscriptionID, resourceGroup, _, _, err = ScaleSetInformationFromProviderID(providerID)
	if err == nil {
		return subscriptionID, resourceGroup, nil
	}
	return "", "", fmt.Errorf("providerID %v is malformatted", providerID)
}

// UIDFromProviderID extracts our own generated unique ID, which is the
// suffix at the resource group, e.g., resource-group-J18dB
// J18dB is the UID.
func UIDFromProviderID(providerID string) (string, error) {
	_, resourceGroup, err := BasicsFromProviderID(providerID)
	if err != nil {
		return "", err
	}

	parts := strings.Split(resourceGroup, "-")
	return parts[len(parts)-1], nil
}

// VMInformationFromProviderID splits a provider's id belonging to a single azure instance into core components.
// A providerID  for individual VMs is build after the following schema:
// - 'azure:///subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.Compute/virtualMachines/<instance-name>'
func VMInformationFromProviderID(providerID string) (subscriptionID, resourceGroup, instanceName string, err error) {
	matches := azureVMProviderIDRegexp.FindStringSubmatch(providerID)
	if len(matches) != 4 {
		return "", "", "", errors.New("error splitting providerID")
	}
	return matches[1], matches[2], matches[3], nil
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
