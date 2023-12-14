/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package components

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Components is a list of Kubernetes components.
type Components []*Component

type legacyComponent struct {
	URL         string `json:"URL,omitempty"`
	Hash        string `json:"Hash,omitempty"`
	InstallPath string `json:"InstallPath,omitempty"`
	Extract     bool   `json:"Extract,omitempty"`
}

// UnmarshalJSON implements a custom JSON unmarshaler to ensure backwards compatibility
// with older components lists which had a different format for all keys.
func (c *Components) UnmarshalJSON(b []byte) error {
	var legacyComponents []*legacyComponent
	if err := json.Unmarshal(b, &legacyComponents); err != nil {
		return err
	}
	var components []*Component
	if err := json.Unmarshal(b, &components); err != nil {
		return err
	}

	if len(legacyComponents) != len(components) {
		return errors.New("failed to unmarshal data: inconsistent number of components in list") // just a check, should never happen
	}

	// If a value is not set in the new format,
	// it might have been set in the old format.
	// In this case, we copy the value from the old format.
	comps := make(Components, len(components))
	for idx := 0; idx < len(components); idx++ {
		comps[idx] = components[idx]
		if comps[idx].Url == "" {
			comps[idx].Url = legacyComponents[idx].URL
		}
		if comps[idx].Hash == "" {
			comps[idx].Hash = legacyComponents[idx].Hash
		}
		if comps[idx].InstallPath == "" {
			comps[idx].InstallPath = legacyComponents[idx].InstallPath
		}
		if !comps[idx].Extract {
			comps[idx].Extract = legacyComponents[idx].Extract
		}
	}

	*c = comps
	return nil
}

// GetHash returns the hash over all component hashes.
func (c Components) GetHash() string {
	sha := sha256.New()
	for _, component := range c {
		sha.Write([]byte(component.Hash))
	}

	return fmt.Sprintf("sha256:%x", sha.Sum(nil))
}

// GetKubeadmComponent returns the kubeadm component.
func (c Components) GetKubeadmComponent() (*Component, error) {
	for _, component := range c {
		if strings.Contains(component.GetUrl(), "kubeadm") {
			return component, nil
		}
	}
	return nil, errors.New("kubeadm component not found")
}

// GetUpgradableComponents returns only those Components that should be passed to the upgrade-agent.
func (c Components) GetUpgradableComponents() Components {
	var cs Components
	for _, c := range c {
		if strings.HasPrefix(c.Url, "data:") || strings.HasSuffix(c.InstallPath, "kubeadm") {
			cs = append(cs, c)
		}
	}
	return cs
}
