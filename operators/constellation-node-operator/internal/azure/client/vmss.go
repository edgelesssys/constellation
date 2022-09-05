/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"fmt"
	"regexp"
)

// vmssIDRegexp is describes the format of an azure virtual machine scale set (VMSS) id.
var vmssIDRegexp = regexp.MustCompile(`^/subscriptions/([^/]+)/resourceGroups/([^/]+)/providers/Microsoft.Compute/virtualMachineScaleSets/([^/]+)$`)

// joinVMSSID joins scale set parameters to generate a virtual machine scale set (VMSS) ID.
// Format: /subscriptions/<subscription>/resourceGroups/<resource-group>/providers/Microsoft.Compute/virtualMachineScaleSets/<scale-set> .
func joinVMSSID(subscriptionID, resourceGroup, scaleSet string) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachineScaleSets/%s", subscriptionID, resourceGroup, scaleSet)
}

// splitVMSSID splits a virtual machine scale set (VMSS) ID into its component.
func splitVMSSID(vmssID string) (subscriptionID, resourceGroup, scaleSet string, err error) {
	matches := vmssIDRegexp.FindStringSubmatch(vmssID)
	if len(matches) != 4 {
		return "", "", "", fmt.Errorf("error splitting vmssID")
	}
	return matches[1], matches[2], matches[3], nil
}
