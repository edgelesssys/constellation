package client

import (
	"encoding/json"
	"errors"

	"github.com/spf13/afero"
)

// cloudConfig uses same format as azure cloud controller manager to share the same kubernetes secret.
// this definition only contains the fields we need.
// reference: https://cloud-provider-azure.sigs.k8s.io/install/configs/ .
type cloudConfig struct {
	TenantID       string `json:"tenantId,omitempty"`
	SubscriptionID string `json:"subscriptionId,omitempty"`
}

// loadConfig loads the cloud config from the given path.
func loadConfig(fs afero.Fs, path string) (*cloudConfig, error) {
	rawConfig, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, err
	}
	var config cloudConfig
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		return nil, err
	}
	if config.TenantID == "" || config.SubscriptionID == "" {
		return nil, errors.New("invalid config: tenantId and subscriptionId are required")
	}
	return &config, nil
}
