package config

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/edgelesssys/constellation/debugd/internal/debugd/deploy"
	configc "github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/file"
)

// CDBGConfig describes the constellation-cli config file.
type CDBGConfig struct {
	ConstellationDebugConfig ConstellationDebugdConfig `yaml:"cdbg"`
}

// ConstellationDebugdConfig is the cdbg specific configuration.
type ConstellationDebugdConfig struct {
	AuthorizedKeys   []configc.UserKey    `yaml:"authorizedKeys"`
	BootstrapperPath string               `yaml:"bootstrapperPath"`
	SystemdUnits     []deploy.SystemdUnit `yaml:"systemdUnits,omitempty"`
}

// FromFile reads a debug configuration.
func FromFile(fileHandler file.Handler, name string) (*CDBGConfig, error) {
	conf := &CDBGConfig{}
	if err := fileHandler.ReadYAML(name, conf); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%s not found - consult the README on how to setup cdbg", name)
		}
		return nil, fmt.Errorf("loading config from file %s: %w", name, err)
	}
	return conf, nil
}
