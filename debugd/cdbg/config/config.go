package config

import (
	"fmt"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/debugd/debugd/deploy"
	"github.com/edgelesssys/constellation/debugd/ssh"
	configc "github.com/edgelesssys/constellation/internal/config"
)

// CDBGConfig describes the constellation-cli config file and extends it with a new field "cdbg".
type CDBGConfig struct {
	ConstellationDebugConfig ConstellationDebugdConfig `yaml:"cdbg"`
	configc.Config
}

// ConstellationDebugdConfig is the cdbg specific configuration.
type ConstellationDebugdConfig struct {
	AuthorizedKeys  []ssh.SSHKey         `yaml:"authorizedKeys"`
	CoordinatorPath string               `yaml:"coordinatorPath"`
	SystemdUnits    []deploy.SystemdUnit `yaml:"systemdUnits,omitempty"`
}

// Default returns a struct with the default config.
func Default() *CDBGConfig {
	return &CDBGConfig{
		ConstellationDebugConfig: ConstellationDebugdConfig{
			AuthorizedKeys:  []ssh.SSHKey{},
			CoordinatorPath: "coordinator",
			SystemdUnits:    []deploy.SystemdUnit{},
		},
		Config: *configc.Default(),
	}
}

// FromFile returns a default config that has been merged with a config file.
// If name is empty, the defaults are returned.
func FromFile(fileHandler file.Handler, name string) (*CDBGConfig, error) {
	conf := Default()
	if name == "" {
		return conf, nil
	}

	if err := fileHandler.ReadYAML(name, conf); err != nil {
		return nil, fmt.Errorf("could not load config from file %s: %w", name, err)
	}
	return conf, nil
}
